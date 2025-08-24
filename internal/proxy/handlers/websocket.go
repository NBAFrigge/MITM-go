package handlers

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"encoding/binary"
	"httpDebugger/internal/certs"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"httpDebugger/internal/proxy/types"
	"httpDebugger/internal/sessiondata"

	"github.com/google/uuid"
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	config  *types.Config
	caCache *certs.CertCache
}

func NewWebSocketHandler(config *types.Config, caCache *certs.CertCache) *WebSocketHandler {
	return &WebSocketHandler{config: config, caCache: caCache}
}

// Handle processes incoming WebSocket upgrade requests and manages the WebSocket connection
func (h *WebSocketHandler) Handle(r *http.Request, session *sessiondata.Session, clientConn net.Conn) {
	h.config.Logger.LogRequest(session)

	// Adjust the URL scheme for WebSocket
	url := r.URL
	if url.Scheme == "http" {
		url.Scheme = "ws"
		session.Request.URL = url.String()
		url.Scheme = "http"
	} else if url.Scheme == "https" {
		url.Scheme = "wss"
		session.Request.URL = url.String()
		url.Scheme = "https"
	}

	// Initialize WebSocket session data
	session.WebSocket = &sessiondata.WebSocketData{
		State:          sessiondata.WSConnecting,
		ConnectedAt:    time.Now(),
		MessageCount:   sessiondata.MessageStats{},
		Messages:       []sessiondata.WebSocketMessage{},
		UpgradeRequest: session.Request,
	}
	h.config.SessionStore.Store(session)

	targetAddr := r.Host
	if !strings.Contains(targetAddr, ":") {
		requireHTTPS := r.URL.Scheme == "https"
		if requireHTTPS {
			targetAddr = net.JoinHostPort(r.Host, "443")
		} else {
			targetAddr = net.JoinHostPort(r.Host, "80")
		}
	}

	var backendConn net.Conn
	var err error

	// Parse host and port
	_, port, parseErr := net.SplitHostPort(targetAddr)
	if parseErr != nil {
		h.config.Logger.LogError(parseErr, "failed to parse target address")
		session.WebSocket.State = sessiondata.WSFailed
		session.Error = parseErr
		session.Duration = time.Since(session.Timestamp)
		return
	}

	useTLS := port == "443"

	// Establish connection to the backend server
	if useTLS {
		backendConn, err = tls.Dial("tcp", targetAddr, &tls.Config{
			ServerName: r.Host,
		})
	} else {
		backendConn, err = net.Dial("tcp", targetAddr)
	}

	// Handle connection errors
	if err != nil {
		h.config.Logger.LogError(err, "failed to dial backend")
		session.WebSocket.State = sessiondata.WSFailed
		session.Error = err
		session.Duration = time.Since(session.Timestamp)
		return
	}
	defer backendConn.Close()

	if err := r.Write(backendConn); err != nil {
		h.config.Logger.LogError(err, "failed to write request to backend")
		session.WebSocket.State = sessiondata.WSFailed
		session.Error = err
		session.Duration = time.Since(session.Timestamp)
		return
	}

	// Read the handshake response from the backend
	handshakeResp, err := http.ReadResponse(bufio.NewReader(backendConn), r)
	if err != nil {
		h.config.Logger.LogError(err, "failed to read handshake response")
		session.WebSocket.State = sessiondata.WSFailed
		session.Error = err
		session.Duration = time.Since(session.Timestamp)
		return
	}

	// Extract and store response data
	session.Response = h.extractResponseData(handshakeResp)
	session.WebSocket.UpgradeResponse = session.Response

	// Write the handshake response back to the client
	if err := handshakeResp.Write(clientConn); err != nil {
		h.config.Logger.LogError(err, "failed to write handshake response to client")
		session.WebSocket.State = sessiondata.WSFailed
		session.Error = err
		session.Duration = time.Since(session.Timestamp)
		return
	}

	// Check if the handshake was successful
	if handshakeResp.StatusCode == http.StatusSwitchingProtocols {
		// Update session state to open
		session.WebSocket = &sessiondata.WebSocketData{
			State:           sessiondata.WSOpen,
			ConnectedAt:     time.Now(),
			MessageCount:    sessiondata.MessageStats{},
			Messages:        []sessiondata.WebSocketMessage{},
			UpgradeRequest:  session.Request,
			UpgradeResponse: session.Response,
			Subprotocol:     handshakeResp.Header.Get("Sec-WebSocket-Protocol"),
		}

		h.config.Logger.LogResponse(session)

		errChan := make(chan error, 2)

		// Start goroutines to handle bidirectional frame forwarding
		go func() {
			defer clientConn.Close()
			defer backendConn.Close()
			err := h.forwardWebSocketFrames(backendConn, clientConn, session, sessiondata.Inbound)
			errChan <- err
		}()

		go func() {
			defer clientConn.Close()
			defer backendConn.Close()
			err := h.forwardWebSocketFrames(clientConn, backendConn, session, sessiondata.Outbound)
			errChan <- err
		}()

		<-errChan

		// Update session state to closed
		session.WebSocket.State = sessiondata.WSClosed
		session.WebSocket.DisconnectedAt = time.Now()
		session.WebSocket.ConnectionDuration = session.WebSocket.DisconnectedAt.Sub(session.WebSocket.ConnectedAt)
		session.Duration = time.Since(session.Timestamp)
	} else {
		session.WebSocket = &sessiondata.WebSocketData{
			State:           sessiondata.WSFailed,
			ConnectedAt:     time.Now(),
			MessageCount:    sessiondata.MessageStats{},
			Messages:        []sessiondata.WebSocketMessage{},
			UpgradeRequest:  session.Request,
			UpgradeResponse: session.Response,
		}
		session.Duration = time.Since(session.Timestamp)
		h.config.Logger.LogResponse(session)
	}
}

// forwardWebSocketFrames reads WebSocket frames from the 'from' connection, processes them, and writes them to the 'to' connection
func (h *WebSocketHandler) forwardWebSocketFrames(from io.Reader, to io.Writer, session *sessiondata.Session, direction sessiondata.MessageDirection) error {
	for {
		// Read the frame header
		header := make([]byte, 2)
		if _, err := io.ReadFull(from, header); err != nil {
			if err != io.EOF {
				h.config.Logger.LogError(err, "reading frame header")
			}
			return err
		}

		fin := header[0]&0x80 != 0
		opcode := header[0] & 0x0F
		masked := header[1]&0x80 != 0
		payloadLen := int64(header[1] & 0x7F)

		// Determine the actual payload length
		if payloadLen == 126 {
			ext := make([]byte, 2)
			if _, err := io.ReadFull(from, ext); err != nil {
				h.config.Logger.LogError(err, "reading extended length")
				return err
			}
			payloadLen = int64(binary.BigEndian.Uint16(ext))
			header = append(header, ext...)
		} else if payloadLen == 127 {
			ext := make([]byte, 8)
			if _, err := io.ReadFull(from, ext); err != nil {
				h.config.Logger.LogError(err, "reading extended length")
				return err
			}
			payloadLen = int64(binary.BigEndian.Uint64(ext))
			header = append(header, ext...)
		}

		var maskKey []byte
		if masked {
			maskKey = make([]byte, 4)
			if _, err := io.ReadFull(from, maskKey); err != nil {
				h.config.Logger.LogError(err, "reading mask key")
				return err
			}
		}

		// Read the payload data
		payload := make([]byte, payloadLen)
		if payloadLen > 0 {
			if _, err := io.ReadFull(from, payload); err != nil {
				h.config.Logger.LogError(err, "reading payload")
				return err
			}
		}

		originalPayload := make([]byte, len(payload))
		copy(originalPayload, payload)

		if masked {
			for i := range payload {
				payload[i] ^= maskKey[i%4]
			}
		}

		// Create and store the WebSocket message
		id, _ := uuid.NewUUID()
		msg := sessiondata.WebSocketMessage{
			ID:         id.String(),
			Timestamp:  time.Now(),
			Direction:  direction,
			Opcode:     opcode,
			Payload:    payload,
			IsMasked:   masked,
			IsFragment: !fin,
			Size:       int(payloadLen),
		}

		// route based on opcode
		switch opcode {
		case 0x1:
			msg.PayloadText = string(payload)
			msg.Type = sessiondata.TextMessage
		case 0x2:
			msg.Type = sessiondata.BinaryMessage
		case 0x8:
			msg.Type = sessiondata.CloseMessage
		case 0x9:
			msg.Type = sessiondata.PingMessage
		case 0xA:
			msg.Type = sessiondata.PongMessage
		default:
			msg.Type = sessiondata.ContinuationMessage
		}

		// Store the message in the session
		h.addMessageToSession(session, msg)

		// Handle control frames and forwarding
		switch opcode {
		case 0x8:
			closeCode := sessiondata.CloseNoStatusReceived
			closeReason := ""

			if len(payload) >= 2 {
				closeCode = int(binary.BigEndian.Uint16(payload[:2]))
				if len(payload) > 2 {
					closeReason = string(payload[2:])
				}
			}

			session.WebSocket.CloseCode = closeCode
			session.WebSocket.CloseReason = closeReason

			h.forwardRawFrame(to, header, maskKey, originalPayload, direction)
			return nil

		default:
			h.forwardRawFrame(to, header, maskKey, originalPayload, direction)
		}
	}
}

// forwardRawFrame writes the raw WebSocket frame to the destination, applying masking if necessary
func (h *WebSocketHandler) forwardRawFrame(to io.Writer, header []byte, maskKey []byte, payload []byte, direction sessiondata.MessageDirection) {
	// Write the frame header
	to.Write(header)

	// Write the mask key and payload
	if direction == sessiondata.Outbound {
		if len(maskKey) == 0 {
			maskKey = make([]byte, 4)
			rand.Read(maskKey)
			to.Write(maskKey)

			maskedPayload := make([]byte, len(payload))
			for i := range payload {
				maskedPayload[i] = payload[i] ^ maskKey[i%4]
			}
			to.Write(maskedPayload)
		} else {
			to.Write(maskKey)
			to.Write(payload)
		}
	} else {
		// Inbound messages are not masked
		to.Write(payload)
	}
}

// addMessageToSession updates the session with the new WebSocket message and updates statistics
func (h *WebSocketHandler) addMessageToSession(session *sessiondata.Session, msg sessiondata.WebSocketMessage) {
	h.config.Mutex.Lock()
	defer h.config.Mutex.Unlock()

	session.WebSocket.Messages = append(session.WebSocket.Messages, msg)
	session.WebSocket.MessageCount.TotalMessages++

	if msg.Type == sessiondata.TextMessage {
		session.WebSocket.MessageCount.TextMessages++
	}

	if msg.Type == sessiondata.BinaryMessage {
		session.WebSocket.MessageCount.BinaryMessages++
	}

	if msg.Direction == sessiondata.Outbound {
		session.WebSocket.MessageCount.OutboundMessages++
		session.WebSocket.MessageCount.OutboundBytes += int64(msg.Size)
	} else {
		session.WebSocket.MessageCount.InboundMessages++
		session.WebSocket.MessageCount.InboundBytes += int64(msg.Size)
	}

	session.WebSocket.MessageCount.TotalBytes += int64(msg.Size)
	if msg.Opcode >= 0x8 && msg.Opcode <= 0xA {
		session.WebSocket.MessageCount.ControlFrames++
	}
}

// extractResponseData extracts basic response data for WebSocket upgrade responses
func (h *WebSocketHandler) extractResponseData(resp *http.Response) *sessiondata.ResponseData {
	return &sessiondata.ResponseData{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
	}
}
