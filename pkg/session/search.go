package session

import (
	"fmt"

	"httpDebugger/pkg/sessiondata"
)

type SearchOptions struct {
	URL        string
	HeadersKey string
	HeadersVal string
	CookiesKey string
	CookiesVal string
	Body       string
}

func (s *InMemoryStore) Search(opt SearchOptions) ([]*sessiondata.Session, error) {
	if opt.URL == "" && opt.HeadersKey == "" && opt.HeadersVal == "" && opt.CookiesKey == "" && opt.CookiesVal == "" && opt.Body == "" {
		return nil, fmt.Errorf("no search criteria provided")
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var results []*sessiondata.Session
	for _, session := range s.order {
		if s.checkIfMatch(session, opt) {
			results = append(results, session)
		}
	}

	return results, nil
}

func (s *InMemoryStore) checkIfMatch(ses *sessiondata.Session, opt SearchOptions) bool {
	if opt.URL != "" {
		if !matchString(ses.Request.URL, opt.URL) {
			return false
		}
	}

	if opt.HeadersKey != "" || opt.HeadersVal != "" {
		if !s.checkHeaders(ses, opt) {
			return false
		}
	}

	if opt.CookiesKey != "" || opt.CookiesVal != "" {
		if !s.checkCookies(ses, opt) {
			return false
		}
	}

	if opt.Body != "" {
		reqMatch := matchString(ses.Request.Body, opt.Body)
		respMatch := ses.Response != nil && matchString(ses.Response.Body, opt.Body)
		if !reqMatch && !respMatch {
			return false
		}
	}

	return true
}

func (s *InMemoryStore) checkHeaders(ses *sessiondata.Session, opt SearchOptions) bool {
	foundKey := opt.HeadersKey == ""
	foundVal := opt.HeadersVal == ""

	for key, val := range ses.Request.Headers.Entries {
		if !foundKey && opt.HeadersKey != "" && matchString(key, opt.HeadersKey) {
			foundKey = true
		}

		if !foundVal && opt.HeadersVal != "" {
			strval := fmt.Sprintf("%v", val)
			if matchString(strval, opt.HeadersVal) {
				foundVal = true
			}
		}

		if foundKey && foundVal {
			return true
		}
	}

	return foundKey && foundVal
}

func (s *InMemoryStore) checkCookies(ses *sessiondata.Session, opt SearchOptions) bool {
	foundKey := opt.CookiesKey == ""
	foundVal := opt.CookiesVal == ""

	for key, val := range ses.Request.Cookies {
		if !foundKey && opt.CookiesKey != "" && matchString(key, opt.CookiesKey) {
			foundKey = true
		}

		if !foundVal && opt.CookiesVal != "" {
			if matchString(val, opt.CookiesVal) {
				foundVal = true
			}
		}

		if foundKey && foundVal {
			return true
		}
	}

	return foundKey && foundVal
}
