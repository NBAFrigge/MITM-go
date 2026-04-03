package helpers

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"

	"httpDebugger/pkg/sessiondata"
)

type SessionListItem struct {
	Session *sessiondata.Session
}

func (i SessionListItem) FilterValue() string {
	return i.Session.Request.URL + " " + i.Session.Request.Method
}

func (i SessionListItem) Title() string {
	status := ""
	if i.Session.Response != nil && i.Session.Response.StatusCode > 0 {
		status = fmt.Sprintf(" [%s]", FormatHTTPStatus(i.Session.Response.StatusCode))
	}

	method := FormatMethod(i.Session.Request.Method)
	url := FormatURL(i.Session.Request.URL, 60)

	return fmt.Sprintf("%s %s%s", method, url, status)
}

func (i SessionListItem) Description() string {
	sessionType := GetSessionTypeText(i.Session)
	duration := FormatDuration(i.Session.Duration)
	timestamp := FormatTimestamp(i.Session.Timestamp)

	return fmt.Sprintf("%s | %s | %s", sessionType, duration, timestamp)
}

func GetSessionTypeIcon(session *sessiondata.Session) string {
	switch session.Type {
	case sessiondata.HTTPSession:
		return "🌐"
	case sessiondata.WebSocketSession:
		return "🔌"
	default:
		return "❓"
	}
}

func GetSessionTypeText(session *sessiondata.Session) string {
	switch session.Type {
	case sessiondata.HTTPSession:
		return "HTTP"
	case sessiondata.WebSocketSession:
		return "WebSocket"
	default:
		return "Unknown"
	}
}

func GetStatusColor(statusCode int) lipgloss.Color {
	switch {
	case statusCode == 0:
		return lipgloss.Color("#6C7086")
	case statusCode >= 200 && statusCode < 300:
		return lipgloss.Color("#A6E3A1")
	case statusCode >= 300 && statusCode < 400:
		return lipgloss.Color("#F9E2AF")
	case statusCode >= 400 && statusCode < 500:
		return lipgloss.Color("#F38BA8")
	case statusCode >= 500:
		return lipgloss.Color("#F38BA8")
	default:
		return lipgloss.Color("#6C7086")
	}
}

func GetMethodColor(method string) lipgloss.Color {
	switch strings.ToUpper(method) {
	case "GET":
		return lipgloss.Color("#89B4FA")
	case "POST":
		return lipgloss.Color("#A6E3A1")
	case "PUT":
		return lipgloss.Color("#F9E2AF")
	case "DELETE":
		return lipgloss.Color("#F38BA8")
	case "PATCH":
		return lipgloss.Color("#CBA6F7")
	case "HEAD":
		return lipgloss.Color("#94E2D5")
	case "OPTIONS":
		return lipgloss.Color("#FAB387")
	default:
		return lipgloss.Color("#6C7086")
	}
}

func SessionToListItem(session *sessiondata.Session) list.Item {
	return SessionListItem{Session: session}
}

func SessionsToListItems(sessions []*sessiondata.Session) []list.Item {
	items := make([]list.Item, len(sessions))
	for i, session := range sessions {
		items[i] = SessionToListItem(session)
	}
	return items
}

func FilterSessions(sessions []*sessiondata.Session, query string) []*sessiondata.Session {
	if query == "" {
		return sessions
	}

	query = strings.ToLower(query)
	var filtered []*sessiondata.Session

	for _, session := range sessions {
		if matchesQuery(session, query) {
			filtered = append(filtered, session)
		}
	}

	return filtered
}

func matchesQuery(session *sessiondata.Session, query string) bool {
	if strings.Contains(strings.ToLower(session.Request.URL), query) {
		return true
	}

	if strings.Contains(strings.ToLower(session.Request.Method), query) {
		return true
	}

	if session.Request.Headers != nil {
		for name, value := range session.Request.Headers.Entries {
			if strings.Contains(strings.ToLower(name), query) {
				return true
			}
			valStr := fmt.Sprintf("%v", value)
			if strings.Contains(strings.ToLower(valStr), query) {
				return true
			}
		}
	}

	if len(session.Request.Body) > 0 && isPrintable(session.Request.Body) {
		if strings.Contains(strings.ToLower(session.Request.Body), query) {
			return true
		}
	}

	if session.Response != nil {
		if strings.Contains(strings.ToLower(FormatHTTPStatus(session.Response.StatusCode)), query) {
			return true
		}

		if session.Response.Headers != nil {
			for name, value := range session.Response.Headers.Entries {
				if strings.Contains(strings.ToLower(name), query) {
					return true
				}
				valStr := fmt.Sprintf("%v", value)
				if strings.Contains(strings.ToLower(valStr), query) {
					return true
				}
			}
		}
	}

	return false
}

func SortSessionsByTime(sessions []*sessiondata.Session) {
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Timestamp.After(sessions[j].Timestamp)
	})
}

func SortSessionsByDuration(sessions []*sessiondata.Session) {
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Duration > sessions[j].Duration
	})
}

func SortSessionsByStatus(sessions []*sessiondata.Session) {
	sort.Slice(sessions, func(i, j int) bool {
		statusI := 0
		statusJ := 0
		if sessions[i].Response != nil {
			statusI = sessions[i].Response.StatusCode
		}
		if sessions[j].Response != nil {
			statusJ = sessions[j].Response.StatusCode
		}
		return statusI > statusJ
	})
}

func GroupSessionsByHost(sessions []*sessiondata.Session) map[string][]*sessiondata.Session {
	groups := make(map[string][]*sessiondata.Session)

	for _, session := range sessions {
		host := extractHost(session.Request.URL)
		groups[host] = append(groups[host], session)
	}

	return groups
}

func extractHost(rawURL string) string {
	if strings.HasPrefix(rawURL, "https://") {
		rawURL = strings.TrimPrefix(rawURL, "https://")
	} else if strings.HasPrefix(rawURL, "http://") {
		rawURL = strings.TrimPrefix(rawURL, "http://")
	}

	if idx := strings.Index(rawURL, "/"); idx != -1 {
		rawURL = rawURL[:idx]
	}

	if idx := strings.Index(rawURL, ":"); idx != -1 {
		rawURL = rawURL[:idx]
	}

	if rawURL == "" {
		return "unknown"
	}

	return rawURL
}

func GetSessionByID(sessions []*sessiondata.Session, id string) *sessiondata.Session {
	for _, session := range sessions {
		if session.ID == id {
			return session
		}
	}
	return nil
}

func GetSessionStats(sessions []*sessiondata.Session) SessionStats {
	stats := SessionStats{}

	if len(sessions) == 0 {
		return stats
	}

	stats.Total = len(sessions)

	var totalDuration int64
	statusCounts := make(map[int]int)

	for _, session := range sessions {
		totalDuration += session.Duration.Nanoseconds()

		switch session.Type {
		case sessiondata.HTTPSession:
			stats.HTTP++
		case sessiondata.WebSocketSession:
			stats.WebSocket++
		}

		if session.Response != nil {
			statusCounts[session.Response.StatusCode]++
		} else {
			stats.NoResponse++
		}
	}

	if stats.Total > 0 {
		stats.AvgDuration = totalDuration / int64(stats.Total)
	}

	for status, count := range statusCounts {
		if status >= 400 {
			stats.Errors += count
		}
	}

	return stats
}

type SessionStats struct {
	Total       int
	HTTP        int
	WebSocket   int
	Errors      int
	NoResponse  int
	AvgDuration int64
}
