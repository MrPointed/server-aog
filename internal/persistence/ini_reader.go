package persistence

import (
	"bufio"
	"io"
	"os"
	"strings"
)

func ReadINI(path string) (map[string]map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read all bytes
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Decode Latin-1/Windows-1252 to UTF-8
	var sb strings.Builder
	sb.Grow(len(bytes))
	for _, b := range bytes {
		sb.WriteRune(rune(b)) // Simple cast works for ISO-8859-1 mapping
	}
	decodedContent := sb.String()

	result := make(map[string]map[string]string)
	var currentSection string
	scanner := bufio.NewScanner(strings.NewReader(decodedContent))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "'") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			endIdx := strings.Index(line, "]")
			if endIdx != -1 {
				currentSection = strings.ToUpper(strings.TrimSpace(line[1:endIdx]))
				result[currentSection] = make(map[string]string)
			}
		} else if currentSection != "" {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.ToUpper(strings.TrimSpace(parts[0]))
				val := strings.TrimSpace(parts[1])
				
				// Handle trailing comments in values if any (common in some DATs)
				if commIdx := strings.Index(val, "'"); commIdx != -1 {
					val = strings.TrimSpace(val[:commIdx])
				}

				result[currentSection][key] = val
			}
		}
	}
	return result, scanner.Err()
}
