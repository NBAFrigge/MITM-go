package sessiondata

const (
	HTTPSession SessionType = iota
	WebSocketSession
)

const (
	WSConnecting WebSocketState = iota
	WSOpen
	WSClosing
	WSClosed
	WSFailed
)

const (
	CloseNormalClosure           = 1000
	CloseGoingAway               = 1001
	CloseProtocolError           = 1002
	CloseUnsupportedData         = 1003
	CloseNoStatusReceived        = 1005
	CloseAbnormalClosure         = 1006
	CloseInvalidFramePayloadData = 1007
	ClosePolicyViolation         = 1008
	CloseMessageTooBig           = 1009
	CloseMandatoryExtension      = 1010
	CloseInternalServerErr       = 1011
	CloseServiceRestart          = 1012
	CloseTryAgainLater           = 1013
	CloseTLSHandshake            = 1015
)

const (
	HTTP2Protocol  = "HTTP/2"
	HTTP11Protocol = "HTTP/1.1"
	HTTP10Protocol = "HTTP/1.0"
)
