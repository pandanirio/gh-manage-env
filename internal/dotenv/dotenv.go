package dotenv

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ParseFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open dotenv file: %w", err)
	}
	defer f.Close()

	out := make(map[string]string)
	sc := bufio.NewScanner(f)

	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// allow: export KEY=VALUE
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		// Strip inline comment only when preceded by space (avoid breaking URLs)
		// ex: KEY=value # comment
		if idx := strings.Index(line, " #"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid dotenv line %d: %q", lineNo, line)
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		// Remove optional quotes
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}

		if key == "" {
			return nil, fmt.Errorf("empty key at line %d", lineNo)
		}

		out[key] = val
	}

	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("scan dotenv file: %w", err)
	}

	return out, nil
}
