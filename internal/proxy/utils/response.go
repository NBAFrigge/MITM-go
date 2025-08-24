package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"httpDebugger/internal/bodyParser"
	"httpDebugger/internal/proxy/types"
	"httpDebugger/internal/sessiondata"
	"httpDebugger/internal/sortedMap"
)

const (
	maxBodySize = 10 * 1024 * 1024
)

// ProcessAndStoreHTTPSession processes an HTTP request, forwards it, and stores the session data
func ProcessAndStoreHTTPSession(w io.Writer, r *http.Request, session *sessiondata.Session, bodyBytes []byte, config *types.Config) {
	config.Logger.LogRequest(session)

	forwardedReq, err := http.NewRequestWithContext(r.Context(), r.Method, r.URL.String(), bytes.NewBuffer(bodyBytes))
	if err != nil {
		HandleProxyError(w, r, err, "Bad Gateway", http.StatusBadGateway, session, config)
		return
	}

	CleanHeader(forwardedReq.Header, r.Header)

	start := time.Now()
	resp, err := config.HTTPClient.Do(forwardedReq)
	if err != nil {
		session.Error = err
		session.Duration = time.Since(start)
		HandleProxyError(w, r, err, "Bad Gateway", http.StatusBadGateway, session, config)
		return
	}

	session.Duration = time.Since(start)
	session.Response = ExtractResponseData(resp, config)
	config.Logger.LogResponse(session)
	config.SessionStore.Store(session)

	if httpWriter, ok := w.(http.ResponseWriter); ok {
		CopyResponse(httpWriter, resp, config)
	} else {
		WriteHTTPResponse(w, resp)
	}
}

// ExtractResponseData extracts relevant data from an HTTP response
func ExtractResponseData(resp *http.Response, config *types.Config) *sessiondata.ResponseData {
	// Read and close the response body
	bodyBytes, err := ReadAndCloseBody(resp)
	if err != nil {
		config.Logger.LogError(err, "reading response body")
		bodyBytes = []byte{}
	}

	if len(bodyBytes) > 0 {
		resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(bodyBytes)))
	}

	// Copy headers to a sorted map
	headers := sortedMap.New()
	for k, v := range resp.Header {
		headers.Put(k, v)
	}

	var parsedBody string
	bpOpt := bodyParser.NewBodyParserOptions()
	bpOpt.PopulateFromHeaders(resp.Header)
	parsedBody, err = bodyParser.Parse(string(bodyBytes), bpOpt)
	if err != nil {
		config.Logger.LogError(err, "parsing response body")
	}

	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	respCookie := resp.Cookies()
	cookies := make(map[string]string, len(respCookie))
	for _, cookie := range respCookie {
		cookies[cookie.Name] = cookie.Value
	}

	return &sessiondata.ResponseData{
		StatusCode:  resp.StatusCode,
		Status:      resp.Status,
		Headers:     headers,
		Cookies:     cookies,
		Body:        parsedBody,
		ContentType: resp.Header.Get("Content-Type"),
	}
}

// WriteHTTPResponse writes an HTTP response to the given connection
func WriteHTTPResponse(conn io.Writer, resp *http.Response) error {
	_, err := fmt.Fprintf(conn, "HTTP/1.1 %d %s\r\n", resp.StatusCode, http.StatusText(resp.StatusCode))
	if err != nil {
		return err
	}
	err = WriteFilteredHeaders(conn, resp.Header)
	if err != nil {
		return err
	}
	_, err = conn.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	_, err = io.Copy(conn, resp.Body)
	return err
}

// CopyResponse copies the HTTP response to the ResponseWriter
func CopyResponse(w http.ResponseWriter, resp *http.Response, config *types.Config) {
	CleanHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	_, err := io.Copy(w, resp.Body)
	if err != nil {
		config.Logger.LogError(err, "copying response body")
	}
}

// ReadAndCloseBody reads all data from the response body and closes it
func ReadAndCloseBody(resp *http.Response) ([]byte, error) {
	limitedReader := io.LimitReader(resp.Body, maxBodySize)

	data, err := io.ReadAll(limitedReader)
	resp.Body.Close()
	return data, err
}
