package headerParser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"httpDebugger/pkg/sortedMap"
)

func ParseHeadersFromRaw(rawRequest []byte) (*sortedMap.SortedMap, error) {
	reader := bufio.NewReader(bytes.NewReader(rawRequest))

	header := sortedMap.New()

	_, _, err := reader.ReadLine()
	if err != nil {
		return nil, fmt.Errorf("error reading first line: %w", err)
	}

	for line, _, err := reader.ReadLine(); err != io.EOF; line, _, err = reader.ReadLine() {
		if string(line) == "\r\n" {
			break
		}

		index := strings.Index(string(line), ":")
		if index == -1 {
			continue
		}

		key := strings.TrimSpace(string(line[:index]))
		value := strings.TrimSpace(string(line[index+1:]))

		header.Put(key, value)
	}

	return header, nil
}
