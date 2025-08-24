package session

import (
	"fmt"
	"httpDebugger/internal/sessiondata"
	"reflect"
)

type SearchOptions struct {
	URL        string
	HeadersKey string
	HeadersVal interface{}
	CookiesKey string
	CookiesVal interface{}
	Body       string
}

func (s *InMemoryStore) Search(opt SearchOptions) ([]*sessiondata.Session, error) {
	if opt.URL == "" && opt.HeadersKey == "" && opt.HeadersVal == nil && opt.CookiesKey == "" && opt.CookiesVal == nil && opt.Body == "" {
		return nil, fmt.Errorf("no search criteria provided")
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var results []*sessiondata.Session
	for _, session := range s.order {
		fmt.Printf("search session %s\n", session.Request.URL)
		if s.CheckIfMatch(session, opt) {
			results = append(results, session)
		}
	}

	fmt.Printf("%+v\n", results)

	return results, nil
}

func (s *InMemoryStore) CheckIfMatch(ses *sessiondata.Session, opt SearchOptions) bool {
	if opt.URL != "" {
		if !matchString(ses.Request.URL, opt.URL) {
			return false
		}
	}

	if opt.HeadersKey != "" || opt.HeadersVal != nil {
		if !s.checkHeaders(ses, opt) {
			return false
		}
	}

	if opt.CookiesKey != "" || !reflect.ValueOf(opt.CookiesVal).IsZero() {
		if !s.checkCookies(ses, opt) {
			return false
		}
	}

	if opt.Body != "" {
		if !matchString(ses.Request.Body, opt.Body) && !matchString(ses.Response.Body, opt.Body) {
			return false
		}
	}

	return true
}

func (s *InMemoryStore) checkHeaders(ses *sessiondata.Session, opt SearchOptions) bool {
	foundKey := opt.HeadersKey == ""
	foundVal := opt.HeadersVal == nil

	valStr := toString(opt.HeadersVal)

	for key, val := range ses.Request.Headers.Entries {
		if !foundKey && opt.HeadersKey != "" && matchString(key, opt.HeadersKey) {
			foundKey = true
		}

		if !foundVal && opt.HeadersVal != nil {
			strval := toString(val)
			if matchString(strval, valStr) {
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
	foundVal := opt.CookiesVal == nil

	valStr := toString(opt.CookiesVal)

	for key, val := range ses.Request.Cookies {
		if !foundKey && opt.CookiesKey != "" && matchString(key, opt.CookiesKey) {
			foundKey = true
		}

		if !foundVal && opt.CookiesVal != nil {
			cookieVal := toString(val)
			if matchString(cookieVal, valStr) {
				foundVal = true
			}
		}

		if foundKey && foundVal {
			return true
		}
	}

	return foundKey && foundVal
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
