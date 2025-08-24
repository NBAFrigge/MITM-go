package handlers

import (
	"errors"
	"io"
	"net/http"

	"httpDebugger/internal/proxy/types"
	"httpDebugger/internal/proxy/utils"
	"httpDebugger/internal/sessiondata"
	"httpDebugger/internal/sortedMap"
)

// HTTPHandler handles standard HTTP requests
type HTTPHandler struct {
	config *types.Config
}

func NewHTTPHandler(config *types.Config) *HTTPHandler {
	return &HTTPHandler{config: config}
}

// TODO add protocol
// Handle processes incoming HTTP requests
func (h *HTTPHandler) Handle(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		utils.HandleProxyError(w, r, errors.New("reading request body"), "Bad Gateway", http.StatusBadGateway, nil, h.config)
		return
	}
	r.Body.Close()

	rawHeaders := sortedMap.New()
	for name, values := range r.Header {
		for _, value := range values {
			rawHeaders.Put(name, value)
		}
	}

	session := sessiondata.NewSessionData(r, bodyBytes, rawHeaders, nil, "")

	switch session.Type {
	case sessiondata.HTTPSession:
		utils.ProcessAndStoreHTTPSession(w, r, session, bodyBytes, h.config)
	case sessiondata.WebSocketSession:
		utils.HandleProxyError(w, r, errors.New("websocket not supported in HTTP handler"),
			"Bad Request", http.StatusBadRequest, session, h.config)
	}
}

func (h *HTTPHandler) copyResponse(w http.ResponseWriter, resp *http.Response) {
	utils.CleanHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	_, err := io.Copy(w, resp.Body)
	if err != nil {
		h.config.Logger.LogError(err, "copying response body")
	}
}
