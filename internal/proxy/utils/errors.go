package utils

import (
	"fmt"
	"io"
	"net/http"

	"httpDebugger/internal/proxy/types"
	"httpDebugger/internal/sessiondata"
)

// HandleProxyError handles errors that occur during proxying
func HandleProxyError(w io.Writer, r *http.Request, err error, statusText string, statusCode int, session *sessiondata.Session, config *types.Config) {
	config.Logger.LogError(err, fmt.Sprintf("proxy error: %s", statusText))

	if session != nil {
		session.Error = err
		session.Response = &sessiondata.ResponseData{
			StatusCode: statusCode,
		}
		config.SessionStore.Store(session)
	}

	if httpWriter, ok := w.(http.ResponseWriter); ok {
		http.Error(httpWriter, statusText, statusCode)
	} else {
		SendErrorResponse(w, statusCode, statusText)
	}
}

// SendErrorResponse sends a simple HTTP error response over the given connection
func SendErrorResponse(conn io.Writer, statusCode int, statusText string) {
	response := fmt.Sprintf("HTTP/1.1 %d %s\r\nContent-Length: 0\r\nConnection: close\r\n\r\n", statusCode, statusText)
	conn.Write([]byte(response))
}
