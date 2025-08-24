package connections

import (
	"bytes"
	"net"
	"sync"

	"httpDebugger/internal/proxy/interfaces"
	"httpDebugger/internal/sortedMap"

	"golang.org/x/net/http2/hpack"
)

type HTTP2FrameWrapper struct {
	net.Conn
	logger            interfaces.Logger
	buffer            []byte
	offset            int
	decoder           *hpack.Decoder
	http2Started      bool
	streams           map[uint32]*HTTP2StreamData
	onHeadersCallback func(streamID uint32, headers *sortedMap.SortedMap)
}

type HTTP2StreamData struct {
	Headers         *sortedMap.SortedMap
	HeadersComplete bool
	Data            []byte
}

var frameBufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 64*1024)
	},
}

var hpackDecoderPool = sync.Pool{
	New: func() interface{} {
		return hpack.NewDecoder(4096, nil)
	},
}

func NewHTTP2FrameWrapper(conn net.Conn, logger interfaces.Logger) *HTTP2FrameWrapper {
	return &HTTP2FrameWrapper{
		Conn:         conn,
		logger:       logger,
		buffer:       frameBufferPool.Get().([]byte),
		offset:       0,
		decoder:      hpackDecoderPool.Get().(*hpack.Decoder),
		http2Started: false,
		streams:      make(map[uint32]*HTTP2StreamData),
	}
}

func (w *HTTP2FrameWrapper) SetHeadersCallback(callback func(streamID uint32, headers *sortedMap.SortedMap)) {
	w.onHeadersCallback = callback
}

// read from the underlying connection and process HTTP/2 frames
func (w *HTTP2FrameWrapper) Read(b []byte) (n int, err error) {
	n, err = w.Conn.Read(b)
	if n > 0 {
		requiredCap := len(w.buffer) + n

		if cap(w.buffer) < requiredCap {
			var newBuffer []byte
			if requiredCap <= 64*1024 {
				newBuffer = frameBufferPool.Get().([]byte)
			} else {
				newBuffer = make([]byte, 0, requiredCap*2)
			}

			if cap(newBuffer) < requiredCap {
				frameBufferPool.Put(newBuffer[:0])
				newBuffer = make([]byte, 0, requiredCap*2)
			}

			newBuffer = newBuffer[:len(w.buffer)]
			copy(newBuffer, w.buffer)

			if cap(w.buffer) <= 64*1024 {
				frameBufferPool.Put(w.buffer[:0])
			}

			w.buffer = newBuffer
		}

		w.buffer = append(w.buffer, b[:n]...)
		w.processBufferedFrames()
	}
	return n, err
}

// process buffered data to extract and log HTTP/2 frames
func (w *HTTP2FrameWrapper) processBufferedFrames() {
	if !w.http2Started {
		w.findHTTP2Start()
		if !w.http2Started {
			return
		}
	}

	for {
		if len(w.buffer)-w.offset < 9 {
			break
		}

		frameHeader := w.buffer[w.offset : w.offset+9]
		length := (uint32(frameHeader[0]) << 16) | (uint32(frameHeader[1]) << 8) | uint32(frameHeader[2])
		frameType := frameHeader[3]
		flags := frameHeader[4]
		streamID := (uint32(frameHeader[5])<<24 | uint32(frameHeader[6])<<16 |
			uint32(frameHeader[7])<<8 | uint32(frameHeader[8])) & 0x7fffffff

		if length > 1048576 {
			w.offset++
			continue
		}

		if frameType > 9 {
			w.offset++
			continue
		}

		totalFrameSize := 9 + int(length)
		if len(w.buffer)-w.offset < totalFrameSize {
			break
		}

		completeFrame := w.buffer[w.offset : w.offset+totalFrameSize]
		w.routeFrame(frameType, flags, streamID, length, completeFrame[9:])

		w.offset += totalFrameSize

		if w.offset > 64*1024 {
			w.buffer = w.buffer[w.offset:]
			w.offset = 0
		}
	}
}

// find the start of HTTP/2 communication in the buffered data
func (w *HTTP2FrameWrapper) findHTTP2Start() {
	preface := []byte("PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n")

	for i := 0; i <= len(w.buffer)-len(preface); i++ {
		if bytes.Equal(w.buffer[i:i+len(preface)], preface) {
			w.offset = i + len(preface)
			w.http2Started = true
			return
		}
	}

	for i := 0; i <= len(w.buffer)-9; i++ {
		if w.buffer[i+3] == 4 {
			length := (uint32(w.buffer[i]) << 16) | (uint32(w.buffer[i+1]) << 8) | uint32(w.buffer[i+2])
			if length <= 1024 && length%6 == 0 {
				w.offset = i
				w.http2Started = true
				return
			}
		}
	}
}

// route frame to appropriate handler based on frame type
func (w *HTTP2FrameWrapper) routeFrame(frameType, flags byte, streamID uint32, length uint32, payload []byte) {
	switch frameType {
	case 0x01:
		headers := w.processHeadersFrame(streamID, flags, payload)
		_ = headers
	case 0x00:
		data := w.processDataFrame(streamID, flags, payload)
		_ = data
	case 0x04:
		w.processSettingsFrame(flags, payload)
	case 0x09:
		headers := w.processContinuationFrame(streamID, flags, payload)
		_ = headers
	case 0x08:
		w.processWindowUpdateFrame(streamID, payload)
	}
}

// process different frame types
func (w *HTTP2FrameWrapper) processHeadersFrame(streamID uint32, flags byte, payload []byte) *sortedMap.SortedMap {
	headerBlock := payload

	if (flags & 0x20) != 0 {
		if len(payload) >= 5 {
			headerBlock = payload[5:]
		}
	}

	if (flags & 0x08) != 0 {
		if len(headerBlock) >= 1 {
			padLength := headerBlock[0]
			if len(headerBlock) > int(padLength) {
				headerBlock = headerBlock[1 : len(headerBlock)-int(padLength)]
			}
		}
	}

	if len(headerBlock) == 0 {
		return nil
	}

	tempDecoder := hpackDecoderPool.Get().(*hpack.Decoder)
	defer hpackDecoderPool.Put(tempDecoder)

	headers, err := tempDecoder.DecodeFull(headerBlock)
	if err != nil {
		w.logger.LogError(err, "Error decoding HPACK headers")
		return nil
	}

	h := sortedMap.New()

	if w.streams[streamID] == nil {
		w.streams[streamID] = &HTTP2StreamData{
			Headers:         h,
			HeadersComplete: false,
			Data:            make([]byte, 0),
		}
	}

	for _, hf := range headers {
		w.streams[streamID].Headers.Put(hf.Name, hf.Value)
	}

	endHeaders := (flags & 0x04) != 0
	if endHeaders {
		w.streams[streamID].HeadersComplete = true
		if w.onHeadersCallback != nil {
			w.onHeadersCallback(streamID, w.streams[streamID].Headers)
		}
	}

	return w.streams[streamID].Headers
}

// process DATA frame, handling padding if present
func (w *HTTP2FrameWrapper) processDataFrame(streamID uint32, flags byte, payload []byte) []byte {
	data := payload
	if (flags & 0x08) != 0 {
		if len(payload) >= 1 {
			padLength := payload[0]
			if len(payload) > int(padLength) {
				data = payload[1 : len(payload)-int(padLength)]
			}
		}
	}

	headers := sortedMap.New()

	if w.streams[streamID] == nil {
		w.streams[streamID] = &HTTP2StreamData{
			Headers:         headers,
			HeadersComplete: false,
			Data:            make([]byte, 0),
		}
	}

	w.streams[streamID].Data = append(w.streams[streamID].Data, data...)
	return data
}

// process SETTINGS frame, ignoring individual settings for now
func (w *HTTP2FrameWrapper) processSettingsFrame(flags byte, payload []byte) {
	if (flags & 0x01) == 0 {
		for i := 0; i < len(payload); i += 6 {
			if i+6 <= len(payload) {
				settingID := uint16(payload[i])<<8 | uint16(payload[i+1])
				value := uint32(payload[i+2])<<24 | uint32(payload[i+3])<<16 |
					uint32(payload[i+4])<<8 | uint32(payload[i+5])
				_ = settingID
				_ = value
			}
		}
	}
}

// process CONTINUATION frame to decode additional headers
func (w *HTTP2FrameWrapper) processContinuationFrame(streamID uint32, flags byte, payload []byte) *sortedMap.SortedMap {
	tempDecoder := hpackDecoderPool.Get().(*hpack.Decoder)
	defer hpackDecoderPool.Put(tempDecoder)

	headers, err := tempDecoder.DecodeFull(payload)
	if err != nil {
		w.logger.LogError(err, "Error decoding CONTINUATION headers")
		return nil
	}

	h := sortedMap.New()

	if w.streams[streamID] == nil {
		w.streams[streamID] = &HTTP2StreamData{
			Headers:         h,
			HeadersComplete: false,
			Data:            make([]byte, 0),
		}
	}

	for _, hf := range headers {
		w.streams[streamID].Headers.Put(hf.Name, hf.Value)
	}

	endHeaders := (flags & 0x04) != 0
	if endHeaders {
		w.streams[streamID].HeadersComplete = true
		if w.onHeadersCallback != nil {
			w.onHeadersCallback(streamID, w.streams[streamID].Headers)
		}
	}

	return w.streams[streamID].Headers
}

func (w *HTTP2FrameWrapper) GetStreamHeaders(streamID uint32) *sortedMap.SortedMap {
	if stream, exists := w.streams[streamID]; exists {
		return stream.Headers
	}

	headers := sortedMap.New()

	return headers
}

func (w *HTTP2FrameWrapper) GetStreamData(streamID uint32) []byte {
	if stream, exists := w.streams[streamID]; exists {
		return stream.Data
	}
	return []byte{}
}

func (w *HTTP2FrameWrapper) CleanupStream(streamID uint32) {
	delete(w.streams, streamID)
}

// process WINDOW_UPDATE frame to adjust flow control window
func (w *HTTP2FrameWrapper) processWindowUpdateFrame(streamID uint32, payload []byte) {
	if len(payload) >= 4 {
		increment := uint32(payload[0])<<24 | uint32(payload[1])<<16 |
			uint32(payload[2])<<8 | uint32(payload[3])
		increment = increment & 0x7fffffff
		_ = increment
	}
}

func (w *HTTP2FrameWrapper) Close() error {
	if w.buffer != nil && cap(w.buffer) <= 64*1024 {
		frameBufferPool.Put(w.buffer[:0])
	}

	if w.decoder != nil {
		hpackDecoderPool.Put(w.decoder)
	}

	return w.Conn.Close()
}
