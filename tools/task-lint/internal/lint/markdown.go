package lint

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type section struct {
	Name       string
	Line       int
	Content    string
	RawContent string
}

type fence struct {
	char  byte
	count int
}

func isValidTaskType(value string) bool {
	if value == "" || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for _, char := range value[1:] {
		if (char < 'a' || char > 'z') &&
			(char < '0' || char > '9') &&
			char != '-' {
			return false
		}
	}
	return true
}

func parseSections(path string) ([]section, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var sections []section
	var current *section
	var content []string
	var activeFence *fence
	inFrontmatter := false

	finishCurrent := func() {
		if current == nil {
			return
		}
		current.RawContent = strings.Join(content, "\n")
		current.Content = strings.TrimSpace(current.RawContent)
		sections = append(sections, *current)
		current = nil
		content = nil
	}

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if lineNumber == 1 && strings.TrimSpace(line) == "---" {
			inFrontmatter = true
			continue
		}
		if inFrontmatter {
			if strings.TrimSpace(line) == "---" {
				inFrontmatter = false
			}
			continue
		}

		if marker, ok := fenceMarker(line); ok {
			if activeFence == nil {
				activeFence = &marker
			} else if marker.char == activeFence.char &&
				marker.count >= activeFence.count &&
				isClosingFence(line, marker) {
				activeFence = nil
			}

			if current != nil {
				content = append(content, line)
			}
			continue
		}

		if activeFence == nil {
			if name, ok := heading(line, 2); ok {
				finishCurrent()
				current = &section{Name: name, Line: lineNumber}
				continue
			}
		}

		if current != nil {
			content = append(content, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read markdown: %w", err)
	}

	finishCurrent()
	return sections, nil
}

func heading(line string, level int) (string, bool) {
	trimmed := trimUpToThreeLeadingSpaces(line)
	marker := strings.Repeat("#", level)
	if !strings.HasPrefix(trimmed, marker) ||
		strings.HasPrefix(trimmed, marker+"#") {
		return "", false
	}
	if len(trimmed) == level ||
		(trimmed[level] != ' ' && trimmed[level] != '\t') {
		return "", false
	}

	name := strings.TrimSpace(trimmed[level:])
	name = strings.TrimSpace(strings.TrimRight(name, "#"))
	if name == "" {
		return "", false
	}
	return name, true
}

func fenceMarker(line string) (fence, bool) {
	trimmed := trimUpToThreeLeadingSpaces(line)
	if len(trimmed) < 3 || (trimmed[0] != '`' && trimmed[0] != '~') {
		return fence{}, false
	}

	char := trimmed[0]
	count := 0
	for count < len(trimmed) && trimmed[count] == char {
		count++
	}
	if count < 3 {
		return fence{}, false
	}
	return fence{char: char, count: count}, true
}

func isClosingFence(line string, marker fence) bool {
	trimmed := trimUpToThreeLeadingSpaces(line)
	index := 0
	for index < len(trimmed) && trimmed[index] == marker.char {
		index++
	}
	return strings.TrimSpace(trimmed[index:]) == ""
}

func trimUpToThreeLeadingSpaces(line string) string {
	spaces := 0
	for spaces < len(line) && spaces < 3 && line[spaces] == ' ' {
		spaces++
	}
	return line[spaces:]
}
