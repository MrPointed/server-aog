package persistence

import (
	"bufio"
	"os"
	"strings"
)

func ReadINI(path string) (map[string]map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result := make(map[string]map[string]string)
	var currentSection string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.ToUpper(line[1 : len(line)-1])
			result[currentSection] = make(map[string]string)
		} else if currentSection != "" {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.ToUpper(strings.TrimSpace(parts[0]))
				val := strings.TrimSpace(parts[1])
				result[currentSection][key] = val
			}
		}
	}
	return result, scanner.Err()
}
