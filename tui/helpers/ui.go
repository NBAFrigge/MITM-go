package helpers

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
)

var (
	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#A6E3A1")).
			Padding(0, 1).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#F38BA8")).
			Padding(0, 1).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#F9E2AF")).
			Padding(0, 1).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#89B4FA")).
			Padding(0, 1)

	MutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C7086"))

	HighlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CBA6F7")).
			Bold(true)
)

func WrapText(text string, width int) string {
	if width <= 0 || text == "" {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	var currentLine strings.Builder
	currentLength := 0

	for _, word := range words {
		wordLen := len(word)

		if wordLen > width {
			if currentLine.Len() > 0 {
				lines = append(lines, currentLine.String())
				currentLine.Reset()
				currentLength = 0
			}

			for len(word) > width {
				lines = append(lines, word[:width])
				word = word[width:]
			}

			if len(word) > 0 {
				currentLine.WriteString(word)
				currentLength = len(word)
			}
			continue
		}

		if currentLength+1+wordLen > width && currentLine.Len() > 0 {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLength = 0
		}

		if currentLine.Len() > 0 {
			currentLine.WriteString(" ")
			currentLength++
		}

		currentLine.WriteString(word)
		currentLength += wordLen
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return strings.Join(lines, "\n")
}

func PadString(s string, width int) string {
	if len(s) >= width {
		return s
	}

	padding := width - len(s)
	return s + strings.Repeat(" ", padding)
}

func PadStringCenter(s string, width int) string {
	if len(s) >= width {
		return s
	}

	padding := width - len(s)
	leftPad := padding / 2
	rightPad := padding - leftPad

	return strings.Repeat(" ", leftPad) + s + strings.Repeat(" ", rightPad)
}

func CreateProgressBar(current, total int, width int) string {
	if total <= 0 || width <= 0 {
		return ""
	}

	if current > total {
		current = total
	}

	filled := int(float64(current) / float64(total) * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	percentage := float64(current) / float64(total) * 100

	return fmt.Sprintf("%s %.1f%%", bar, percentage)
}

func ColorizeByStatus(text string, statusCode int) string {
	color := GetStatusColor(statusCode)
	style := lipgloss.NewStyle().Foreground(color)
	return style.Render(text)
}

func ColorizeByMethod(text string, method string) string {
	color := GetMethodColor(method)
	style := lipgloss.NewStyle().Foreground(color)
	return style.Render(text)
}

func CreateBox(content string, title string, width int) string {
	if width < 4 {
		return content
	}

	lines := strings.Split(content, "\n")
	boxWidth := width - 2

	var result strings.Builder

	if title != "" {
		titleLine := "┌─ " + title + " "
		remaining := width - len(titleLine) - 1
		if remaining > 0 {
			titleLine += strings.Repeat("─", remaining) + "┐"
		} else {
			titleLine = titleLine[:width-1] + "┐"
		}
		result.WriteString(titleLine + "\n")
	} else {
		result.WriteString("┌" + strings.Repeat("─", width-2) + "┐\n")
	}

	for _, line := range lines {
		if len(line) > boxWidth {
			line = line[:boxWidth]
		}
		padding := boxWidth - len(line)
		result.WriteString("│" + line + strings.Repeat(" ", padding) + "│\n")
	}

	result.WriteString("└" + strings.Repeat("─", width-2) + "┘")

	return result.String()
}

func CreateTable(headers []string, rows [][]string, colWidths []int) string {
	if len(headers) == 0 || len(colWidths) != len(headers) {
		return ""
	}

	var result strings.Builder

	result.WriteString("┌")
	for i, width := range colWidths {
		result.WriteString(strings.Repeat("─", width))
		if i < len(colWidths)-1 {
			result.WriteString("┬")
		}
	}
	result.WriteString("┐\n")

	result.WriteString("│")
	for i, header := range headers {
		padded := PadStringCenter(header, colWidths[i])
		result.WriteString(padded)
		if i < len(headers)-1 {
			result.WriteString("│")
		}
	}
	result.WriteString("│\n")

	result.WriteString("├")
	for i, width := range colWidths {
		result.WriteString(strings.Repeat("─", width))
		if i < len(colWidths)-1 {
			result.WriteString("┼")
		}
	}
	result.WriteString("┤\n")

	for _, row := range rows {
		result.WriteString("│")
		for i, cell := range row {
			if i >= len(colWidths) {
				break
			}
			padded := PadString(TruncateString(cell, colWidths[i]), colWidths[i])
			result.WriteString(padded)
			if i < len(headers)-1 {
				result.WriteString("│")
			}
		}
		result.WriteString("│\n")
	}

	result.WriteString("└")
	for i, width := range colWidths {
		result.WriteString(strings.Repeat("─", width))
		if i < len(colWidths)-1 {
			result.WriteString("┴")
		}
	}
	result.WriteString("┘")

	return result.String()
}

func HighlightSearchTerm(text, searchTerm string) string {
	if searchTerm == "" {
		return text
	}

	lowerText := strings.ToLower(text)
	lowerTerm := strings.ToLower(searchTerm)

	if !strings.Contains(lowerText, lowerTerm) {
		return text
	}

	var result strings.Builder
	lastIndex := 0

	for {
		index := strings.Index(lowerText[lastIndex:], lowerTerm)
		if index == -1 {
			break
		}

		actualIndex := lastIndex + index

		result.WriteString(text[lastIndex:actualIndex])

		match := text[actualIndex : actualIndex+len(searchTerm)]
		result.WriteString(HighlightStyle.Render(match))

		lastIndex = actualIndex + len(searchTerm)
	}

	result.WriteString(text[lastIndex:])

	return result.String()
}

func CreateStatusIndicator(status string, isActive bool) string {
	var style lipgloss.Style

	if isActive {
		style = SuccessStyle
	} else {
		style = ErrorStyle
	}

	return style.Render(status)
}

func FormatKeyValue(key, value string, keyWidth int) string {
	paddedKey := PadString(key+":", keyWidth)
	return MutedStyle.Render(paddedKey) + " " + value
}

func CreateSeparator(width int, char string) string {
	if char == "" {
		char = "─"
	}
	return strings.Repeat(char, width)
}

func CleanText(text string) string {
	var cleaned strings.Builder

	for _, r := range text {
		if unicode.IsPrint(r) || r == '\n' || r == '\t' {
			cleaned.WriteRune(r)
		} else {
			cleaned.WriteRune('?')
		}
	}

	result := cleaned.String()
	result = strings.ReplaceAll(result, "\t", "    ")

	return result
}

func CreateLoadingSpinner(frame int) string {
	spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	return spinners[frame%len(spinners)]
}
