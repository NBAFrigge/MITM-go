package helpers

import (
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"httpDebugger/pkg/sessiondata"
)

type SessionFilter struct {
	Method      string
	Host        string
	StatusCode  int
	StatusRange string
	MinDuration time.Duration
	MaxDuration time.Duration
	HasError    *bool
	TimeRange   TimeRange
	ContentType string
}

type TimeRange struct {
	Start time.Time
	End   time.Time
}

func ApplyFilter(sessions []*sessiondata.Session, filter SessionFilter) []*sessiondata.Session {
	var filtered []*sessiondata.Session

	for _, session := range sessions {
		if matchesFilter(session, filter) {
			filtered = append(filtered, session)
		}
	}

	return filtered
}

func matchesFilter(session *sessiondata.Session, filter SessionFilter) bool {
	if filter.Method != "" && !strings.EqualFold(session.Request.Method, filter.Method) {
		return false
	}

	if filter.Host != "" {
		sessionHost := extractHost(session.Request.URL)
		if !strings.Contains(strings.ToLower(sessionHost), strings.ToLower(filter.Host)) {
			return false
		}
	}

	if filter.StatusCode > 0 {
		if session.Response == nil || session.Response.StatusCode != filter.StatusCode {
			return false
		}
	}

	if filter.StatusRange != "" && session.Response != nil {
		if !matchesStatusRange(session.Response.StatusCode, filter.StatusRange) {
			return false
		}
	}

	if filter.MinDuration > 0 && session.Duration < filter.MinDuration {
		return false
	}

	if filter.MaxDuration > 0 && session.Duration > filter.MaxDuration {
		return false
	}

	if filter.HasError != nil {
		hasError := session.Error != nil || (session.Response != nil && session.Response.StatusCode >= 400)
		if *filter.HasError != hasError {
			return false
		}
	}

	if !filter.TimeRange.Start.IsZero() && session.Timestamp.Before(filter.TimeRange.Start) {
		return false
	}
	if !filter.TimeRange.End.IsZero() && session.Timestamp.After(filter.TimeRange.End) {
		return false
	}

	if filter.ContentType != "" {
		contentType := getContentType(session)
		if !strings.Contains(strings.ToLower(contentType), strings.ToLower(filter.ContentType)) {
			return false
		}
	}

	return true
}

func matchesStatusRange(statusCode int, statusRange string) bool {
	switch strings.ToLower(statusRange) {
	case "1xx":
		return statusCode >= 100 && statusCode < 200
	case "2xx":
		return statusCode >= 200 && statusCode < 300
	case "3xx":
		return statusCode >= 300 && statusCode < 400
	case "4xx":
		return statusCode >= 400 && statusCode < 500
	case "5xx":
		return statusCode >= 500 && statusCode < 600
	default:
		return false
	}
}

func getContentType(session *sessiondata.Session) string {
	if session.Request.Headers != nil {
		ct, ok := session.Request.Headers.Get("Content-Type")
		if ok {
			return ct.(string)
		}
	}

	if session.Response != nil && session.Response.Headers != nil {
		ct, ok := session.Request.Headers.Get("Content-Type")
		if ok {
			return ct.(string)
		}
	}
	return ""
}

type SessionSorter struct {
	Field     string
	Direction string
}

func SortSessions(sessions []*sessiondata.Session, sorter SessionSorter) {
	if sorter.Direction == "" {
		sorter.Direction = "desc"
	}

	sort.Slice(sessions, func(i, j int) bool {
		var less bool

		switch strings.ToLower(sorter.Field) {
		case "time", "timestamp":
			less = sessions[i].Timestamp.Before(sessions[j].Timestamp)
		case "duration":
			less = sessions[i].Duration < sessions[j].Duration
		case "status":
			statusI := getStatusCode(sessions[i])
			statusJ := getStatusCode(sessions[j])
			less = statusI < statusJ
		case "method":
			less = sessions[i].Request.Method < sessions[j].Request.Method
		case "url":
			less = sessions[i].Request.URL < sessions[j].Request.URL
		case "size":
			sizeI := getResponseSize(sessions[i])
			sizeJ := getResponseSize(sessions[j])
			less = sizeI < sizeJ
		default:

			less = sessions[i].Timestamp.Before(sessions[j].Timestamp)
		}

		if strings.ToLower(sorter.Direction) == "desc" {
			return !less
		}
		return less
	})
}

func getStatusCode(session *sessiondata.Session) int {
	if session.Response != nil {
		return session.Response.StatusCode
	}
	return 0
}

func getResponseSize(session *sessiondata.Session) int64 {
	if session.Response != nil {
		return int64(len(session.Response.Body))
	}
	return 0
}

func GroupSessions(sessions []*sessiondata.Session, groupBy string) map[string][]*sessiondata.Session {
	groups := make(map[string][]*sessiondata.Session)

	for _, session := range sessions {
		var key string

		switch strings.ToLower(groupBy) {
		case "host":
			key = extractHost(session.Request.URL)
		case "method":
			key = session.Request.Method
		case "status":
			if session.Response != nil {
				key = strconv.Itoa(session.Response.StatusCode)
			} else {
				key = "No Response"
			}
		case "type":
			key = GetSessionTypeText(session)
		case "hour":
			key = session.Timestamp.Format("15:00")
		case "date":
			key = session.Timestamp.Format("2006-01-02")
		case "contenttype":
			ct := getContentType(session)
			if ct == "" {
				key = "Unknown"
			} else {
				key = FormatContentType(ct)
			}
		default:
			key = "All"
		}

		groups[key] = append(groups[key], session)
	}

	return groups
}

func GetUniqueHosts(sessions []*sessiondata.Session) []string {
	hostMap := make(map[string]bool)

	for _, session := range sessions {
		host := extractHost(session.Request.URL)
		hostMap[host] = true
	}

	var hosts []string
	for host := range hostMap {
		hosts = append(hosts, host)
	}

	sort.Strings(hosts)
	return hosts
}

func GetUniqueMethods(sessions []*sessiondata.Session) []string {
	methodMap := make(map[string]bool)

	for _, session := range sessions {
		methodMap[session.Request.Method] = true
	}

	var methods []string
	for method := range methodMap {
		methods = append(methods, method)
	}

	sort.Strings(methods)
	return methods
}

func GetUniqueStatusCodes(sessions []*sessiondata.Session) []int {
	statusMap := make(map[int]bool)

	for _, session := range sessions {
		if session.Response != nil {
			statusMap[session.Response.StatusCode] = true
		}
	}

	var statuses []int
	for status := range statusMap {
		statuses = append(statuses, status)
	}

	sort.Ints(statuses)
	return statuses
}

func FindSessionsByURL(sessions []*sessiondata.Session, urlPattern string) []*sessiondata.Session {
	var matches []*sessiondata.Session

	pattern := strings.ToLower(urlPattern)

	for _, session := range sessions {
		if strings.Contains(strings.ToLower(session.Request.URL), pattern) {
			matches = append(matches, session)
		}
	}

	return matches
}

func FindSimilarSessions(sessions []*sessiondata.Session, target *sessiondata.Session) []*sessiondata.Session {
	var similar []*sessiondata.Session

	targetPath := extractPath(target.Request.URL)
	targetHost := extractHost(target.Request.URL)

	for _, session := range sessions {
		if session.ID == target.ID {
			continue
		}

		if extractHost(session.Request.URL) == targetHost &&
			session.Request.Method == target.Request.Method {

			sessionPath := extractPath(session.Request.URL)
			if pathsSimilar(targetPath, sessionPath) {
				similar = append(similar, session)
			}
		}
	}

	return similar
}

func extractPath(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {

		if idx := strings.Index(rawURL, ":
			rawURL = rawURL[idx+3:]
		}
		if idx := strings.Index(rawURL, "/"); idx != -1 {
			return rawURL[idx:]
		}
		return "/"
	}

	return parsedURL.Path
}

func pathsSimilar(path1, path2 string) bool {
	if path1 == path2 {
		return true
	}

	segments1 := strings.Split(strings.Trim(path1, "/"), "/")
	segments2 := strings.Split(strings.Trim(path2, "/"), "/")

	if len(segments1) != len(segments2) {
		return false
	}

	for i := 0; i < len(segments1); i++ {
		seg1 := segments1[i]
		seg2 := segments2[i]

		if isParameter(seg1) && isParameter(seg2) {
			continue
		}

		if seg1 != seg2 {
			return false
		}
	}

	return true
}

func isParameter(segment string) bool {
	if _, err := strconv.Atoi(segment); err == nil {
		return true
	}

	if len(segment) == 36 && strings.Count(segment, "-") == 4 {
		return true
	}

	if len(segment) > 10 && isAlphanumeric(segment) {
		return true
	}

	return false
}

func isAlphanumeric(s string) bool {
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}

func CalculateSessionsPerHour(sessions []*sessiondata.Session) map[string]int {
	hourCount := make(map[string]int)

	for _, session := range sessions {
		hour := session.Timestamp.Format("15:00")
		hourCount[hour]++
	}

	return hourCount
}

func CalculateAverageResponseTime(sessions []*sessiondata.Session) time.Duration {
	if len(sessions) == 0 {
		return 0
	}

	var total time.Duration
	count := 0

	for _, session := range sessions {
		if session.Response != nil {
			total += session.Duration
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return total / time.Duration(count)
}

func FindSlowSessions(sessions []*sessiondata.Session, threshold time.Duration) []*sessiondata.Session {
	var slow []*sessiondata.Session

	for _, session := range sessions {
		if session.Duration > threshold {
			slow = append(slow, session)
		}
	}

	sort.Slice(slow, func(i, j int) bool {
		return slow[i].Duration > slow[j].Duration
	})

	return slow
}

func FindErrorSessions(sessions []*sessiondata.Session) []*sessiondata.Session {
	var errors []*sessiondata.Session

	for _, session := range sessions {
		hasError := session.Error != nil
		if session.Response != nil && session.Response.StatusCode >= 400 {
			hasError = true
		}

		if hasError {
			errors = append(errors, session)
		}
	}

	return errors
}
