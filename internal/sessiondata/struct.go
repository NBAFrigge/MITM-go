package sessiondata

import (
	"httpDebugger/internal/sortedMap"
	"time"
)

type RequestDifference struct {
	Method      *FieldDiff
	URL         *FieldDiff
	Headers     *HeadersDiff
	Cookies     *CookiesDiff
	Body        *FieldDiff
	ContentType *FieldDiff
	HasDiffs    bool
}

type FieldDiff struct {
	Original string
	Other    string
	Changed  bool
}

type HeadersDiff struct {
	Added    map[string]interface{}
	Removed  map[string]interface{}
	Modified map[string]FieldDiff
	Changed  bool
}

type CookiesDiff struct {
	Added    map[string]interface{}
	Removed  map[string]interface{}
	Modified map[string]FieldDiff
	Changed  bool
}

type SessionType int

type RequestData struct {
	Method      string
	URL         string
	Headers     *sortedMap.SortedMap
	Cookies     map[string]string
	Body        string
	ContentType string
	IsUpgrade   bool
}

type ResponseData struct {
	StatusCode  int
	Status      string
	Headers     *sortedMap.SortedMap
	Cookies     map[string]string
	Body        string
	ContentType string
	IsUpgrade   bool
}

type WebSocketData struct {
	UpgradeRequest  *RequestData
	UpgradeResponse *ResponseData

	State              WebSocketState
	ConnectedAt        time.Time
	DisconnectedAt     time.Time
	ConnectionDuration time.Duration

	Messages     []WebSocketMessage
	MessageCount MessageStats

	Subprotocol string
	Extensions  []string
	CloseCode   int
	CloseReason string
}

type WebSocketState int

type WebSocketMessage struct {
	ID          string
	Timestamp   time.Time
	Direction   MessageDirection
	Type        MessageType
	Opcode      uint8
	Payload     []byte
	PayloadText string
	IsMasked    bool
	IsFragment  bool
	Size        int
}

type MessageDirection int

const (
	Inbound MessageDirection = iota
	Outbound
)

type MessageType int

const (
	TextMessage MessageType = iota
	BinaryMessage
	CloseMessage
	PingMessage
	PongMessage
	ContinuationMessage
)

type MessageStats struct {
	TotalMessages    int
	InboundMessages  int
	OutboundMessages int
	TextMessages     int
	BinaryMessages   int
	ControlFrames    int
	TotalBytes       int64
	InboundBytes     int64
	OutboundBytes    int64
}
