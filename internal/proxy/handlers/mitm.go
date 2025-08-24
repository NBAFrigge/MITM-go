package handlers

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/http2"

	"httpDebugger/internal/certs"
	"httpDebugger/internal/clientHello"
	"httpDebugger/internal/headerParser"
	"httpDebugger/internal/proxy/connections"
	"httpDebugger/internal/proxy/types"
	"httpDebugger/internal/proxy/utils"
	"httpDebugger/internal/sessiondata"
	"httpDebugger/internal/sortedMap"
)

const (
	ConnectionTimeoutSeconds = 30
	HTTP10                   = "HTTP/1.0"
	ConnectionHeaderClose    = "close"
	HTTPSScheme              = "https"
	InitialBufferIndex       = 0
)

type MITMHandler struct {
	config     *types.Config
	certsCache *certs.CertCache
	tlsCache   *clientHello.ClientHelloCache
}

func NewMITMHandler(config *types.Config, caCerts *certs.CertCache) *MITMHandler {
	return &MITMHandler{
		config:     config,
		certsCache: caCerts,
		tlsCache:   clientHello.NewClientHelloCache(),
	}

}

// Handle processes incoming CONNECT requests for HTTPS interception
func (h *MITMHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Extract host without port for certificate generation
	hostWithoutPort, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		hostWithoutPort = r.Host
	}

	// Hijack the connection to take over the TCP connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		h.config.Logger.LogError(errors.New("hijacking not supported"), "hijacking client connection")
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, clientBuffer, err := hijacker.Hijack()
	if err != nil {
		h.config.Logger.LogError(err, "hijacking client connection")
		return
	}
	defer clientConn.Close()

	// Read any buffered data from the client
	bufferedData := []byte{}
	if clientBuffer != nil && clientBuffer.Reader.Buffered() > InitialBufferIndex {
		bufferedSize := clientBuffer.Reader.Buffered()
		bufferedData = make([]byte, bufferedSize)
		_, err := clientBuffer.Reader.Read(bufferedData)
		if err != nil {
			h.config.Logger.LogError(err, "reading buffered data")
		}
	}

	// Send 200 OK response to the client to establish the tunnel
	_, err = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	if err != nil {
		h.config.Logger.LogError(err, "writing CONNECT response")
		return
	}

	// Wrap the connection to handle any buffered data
	var baseConn net.Conn = clientConn
	if len(bufferedData) > InitialBufferIndex {
		baseConn = connections.NewBufferedConn(clientConn, bufferedData)
	}

	// Parse the ClientHello message to extract TLS parameters
	var clientHelloData []byte
	var tlsConfig *tls.Config

	if len(bufferedData) > 5 && bufferedData[0] == 0x16 && bufferedData[5] == 0x01 {
		clientHelloData = bufferedData
	} else {
		buffer := make([]byte, 4096)
		n, err := baseConn.Read(buffer)
		if err != nil && !errors.Is(err, io.EOF) {
			h.config.Logger.LogError(err, "reading HTTPS request")
			return
		}
		clientHelloData = buffer[:n]
	}

	tlsConfig, found := h.tlsCache.Get(clientHelloData)
	if !found {
		tlsConfig, err = clientHello.ParseClientHello(clientHelloData)
		if err != nil {
			h.config.Logger.LogError(err, "reading client hello message")
			tlsConfig = &tls.Config{
				MinVersion: tls.VersionTLS12,
				MaxVersion: tls.VersionTLS13,
			}
		}
		h.tlsCache.Set(clientHelloData, tlsConfig)
	}

	// Generate a certificate for the requested host
	cert, err := h.certsCache.GetHostCert(hostWithoutPort, h.config.CACert)
	if err != nil {
		h.config.Logger.LogError(err, "failed to generate certificate for host: "+r.Host)
		http.Error(w, "Certificate Error", http.StatusInternalServerError)
		return
	}

	tlsConfig.Certificates = []tls.Certificate{cert}

	// Wrap the connection to replay the ClientHello data
	replayConn := connections.NewReplayConn(baseConn, clientHelloData)
	tlsConn := tls.Server(replayConn, tlsConfig)
	defer tlsConn.Close()

	// Perform the TLS handshake
	err = tlsConn.Handshake()
	if err != nil {
		h.config.Logger.LogError(err, fmt.Sprintf("TLS handshake failed for %s", r.Host))
		return
	}

	// Route based on negotiated protocol
	switch tlsConn.ConnectionState().NegotiatedProtocol {
	case "h2":
		h.handleTLSHTTP2(tlsConn, r.Host, tlsConfig)
	case "http/1.1", "":
		h.handleTLSHTTP1(tlsConn, r.Host, tlsConfig)
	}
}

// handleTLSHTTP1 processes HTTPS connections using HTTP/1.1
func (h *MITMHandler) handleTLSHTTP1(clientConn *tls.Conn, originalHost string, tlsConfig *tls.Config) {
	// Initialize header parser
	wsHandler := NewWebSocketHandler(h.config, h.certsCache)

	for {
		clientConn.SetDeadline(time.Now().Add(ConnectionTimeoutSeconds * time.Second))

		// Capture raw request data
		var capturedData bytes.Buffer
		captureConn := connections.NewCapturingConn(clientConn, &capturedData)

		// Read the HTTP request
		reader := bufio.NewReader(captureConn)
		req, err := http.ReadRequest(reader)
		if err != nil {
			if err == io.EOF {
				return
			}
			h.config.Logger.LogError(err, "reading HTTPS request")
			return
		}

		// Parse raw headers
		rawRequest := capturedData.Bytes()
		headers, err := headerParser.ParseHeadersFromRaw(rawRequest)
		if err != nil {
			h.config.Logger.LogError(err, "reading raw header")
		}

		clientConn.SetDeadline(time.Time{})

		// Update request URL and Host
		req.URL.Scheme = HTTPSScheme
		req.URL.Host = originalHost
		if req.Host == "" {
			req.Host = originalHost
		}

		// Get the full request body
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			h.config.Logger.LogError(err, "reading HTTPS request body")
			return
		}
		req.Body.Close()

		session := sessiondata.NewSessionData(req, bodyBytes, headers, tlsConfig, sessiondata.HTTP11Protocol)

		// Handle based on session type
		switch session.Type {
		case sessiondata.HTTPSession:
			utils.ProcessAndStoreHTTPSession(clientConn, req, session, bodyBytes, h.config)
		case sessiondata.WebSocketSession:
			wsHandler.Handle(req, session, clientConn)
		}

		if req.Proto == HTTP10 || strings.ToLower(req.Header.Get("Connection")) == ConnectionHeaderClose {
			return
		}
	}
}

// handleTLSHTTP2 processes HTTPS connections using HTTP/2
func (h *MITMHandler) handleTLSHTTP2(clientConn *tls.Conn, originalHost string, tlsConfig *tls.Config) {
	// Wrap the connection to capture HTTP/2 frames
	wrappedConn := connections.NewHTTP2FrameWrapper(clientConn, h.config.Logger)
	defer wrappedConn.Close()

	// Map to store headers for each stream
	streamHeaders := make(map[uint32]*sortedMap.SortedMap)

	// Set callback to capture headers
	wrappedConn.SetHeadersCallback(func(streamID uint32, headers *sortedMap.SortedMap) {
		streamHeaders[streamID] = headers
	})

	// Initialize HTTP/2 server
	http2Server := &http2.Server{
		MaxHandlers:                  1000,
		MaxConcurrentStreams:         250,
		MaxReadFrameSize:             1048576,
		PermitProhibitedCipherSuites: false,
	}

	// Serve HTTP/2 requests
	http2Server.ServeConn(wrappedConn, &http2.ServeConnOpts{
		// Custom handler to process each request
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			//h.config.Logger.LogInfo(fmt.Sprintf("Processing HTTP/2 request: %s %s", req.Method, req.URL.Path))

			// Update request URL and Host
			req.URL.Scheme = HTTPSScheme
			req.URL.Host = originalHost
			if req.Host == "" {
				req.Host = originalHost
			}

			// Retrieve captured headers for the stream
			rawHeaders := sortedMap.New()
			streamID := h.extractStreamID(req)
			if capturedHeaders, exists := streamHeaders[streamID]; exists {
				rawHeaders = capturedHeaders
				delete(streamHeaders, streamID)
			} else {
				rawHeaders = sortedMap.New()
				for key, value := range req.Header {
					rawHeaders.Put(key, value)
				}
			}

			// Get the full request body
			bodyBytes, err := io.ReadAll(req.Body)
			if err != nil {
				h.config.Logger.LogError(err, "reading HTTPS request body")
				http.Error(w, "Error reading request body", http.StatusBadRequest)
				return
			}
			req.Body.Close()

			// Create session data
			session := sessiondata.NewSessionData(req, bodyBytes, rawHeaders, tlsConfig, sessiondata.HTTP2Protocol)

			// Handle based on session type
			switch session.Type {
			case sessiondata.HTTPSession:
				utils.ProcessAndStoreHTTPSession(w, req, session, bodyBytes, h.config)
			case sessiondata.WebSocketSession:
				wsHandler := NewWebSocketHandler(h.config, h.certsCache)
				wsHandler.Handle(req, session, clientConn)
			default:
				h.config.Logger.LogError(fmt.Errorf("unsupported session type for HTTP/2: %v", session.Type), "unsupported session type")
				http.Error(w, "Unsupported session type", http.StatusBadRequest)
			}
		}),
		BaseConfig: &http.Server{
			ReadTimeout:  ConnectionTimeoutSeconds * time.Second,
			WriteTimeout: ConnectionTimeoutSeconds * time.Second,
		},
	})
}

// extractStreamID generates a pseudo stream ID based on request path and method
func (h *MITMHandler) extractStreamID(req *http.Request) uint32 {
	hash := uint32(0)
	path := req.URL.Path + req.Method
	for _, b := range []byte(path) {
		hash = hash*31 + uint32(b)
	}
	return hash & 0x7fffffff
}
