package lint

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type frontmatterDocument struct {
	Fields map[string]frontmatterValue
}

type frontmatterValue struct {
	Value *yaml.Node
	Line  int
}

type frontmatterError struct {
	Line    int
	Message string
}

func parseFrontmatter(path string) (frontmatterDocument, []frontmatterError, error) {
	file, err := os.Open(path)
	if err != nil {
		return frontmatterDocument{}, nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return frontmatterDocument{}, nil, err
		}
		return frontmatterDocument{}, []frontmatterError{{
			Line:    1,
			Message: "문서 시작에 YAML frontmatter가 없습니다",
		}}, nil
	}
	if strings.TrimSpace(scanner.Text()) != "---" {
		return frontmatterDocument{}, []frontmatterError{{
			Line:    1,
			Message: "문서 시작에 YAML frontmatter가 없습니다",
		}}, nil
	}

	var lines []string
	closed := false
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == "---" {
			closed = true
			break
		}
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return frontmatterDocument{}, nil, err
	}
	if !closed {
		return frontmatterDocument{}, []frontmatterError{{
			Line:    1,
			Message: "YAML frontmatter를 닫는 '---'가 없습니다",
		}}, nil
	}

	var document yaml.Node
	if err := yaml.Unmarshal([]byte(strings.Join(lines, "\n")), &document); err != nil {
		return frontmatterDocument{}, []frontmatterError{{
			Line:    1,
			Message: fmt.Sprintf("YAML frontmatter를 해석할 수 없습니다: %v", err),
		}}, nil
	}
	if len(document.Content) == 0 || len(document.Content[0].Content) == 0 {
		return frontmatterDocument{Fields: map[string]frontmatterValue{}}, nil, nil
	}
	mapping := document.Content[0]
	if mapping.Kind != yaml.MappingNode {
		return frontmatterDocument{}, []frontmatterError{{
			Line:    2,
			Message: "YAML frontmatter는 key-value mapping이어야 합니다",
		}}, nil
	}

	fields := make(map[string]frontmatterValue)
	var errors []frontmatterError
	for index := 0; index < len(mapping.Content); index += 2 {
		key := mapping.Content[index]
		value := mapping.Content[index+1]
		line := key.Line + 1
		if _, exists := fields[key.Value]; exists {
			errors = append(errors, frontmatterError{
				Line:    line,
				Message: fmt.Sprintf("frontmatter의 '%s' field가 중복되었습니다", key.Value),
			})
			continue
		}
		fields[key.Value] = frontmatterValue{Value: value, Line: line}
	}
	return frontmatterDocument{Fields: fields}, errors, nil
}

func lintFrontmatter(
	relativePath string,
	document frontmatterDocument,
	schema frontmatterSchema,
) []Error {
	declared := make(map[string]fieldSchema)
	var errors []Error
	for _, field := range schema.Fields {
		declared[field.Name] = field
		value, exists := document.Fields[field.Name]
		if !exists {
			if isRequired(field.Required) {
				errors = append(errors, Error{
					Path:    relativePath,
					Line:    1,
					Message: fmt.Sprintf("필수 frontmatter field '%s'이 없습니다", field.Name),
				})
			}
			continue
		}
		if !nodeMatchesType(value.Value, field.Type) {
			errors = append(errors, Error{
				Path: relativePath,
				Line: value.Line,
				Message: fmt.Sprintf(
					"frontmatter field '%s'은 %s type이어야 합니다",
					field.Name,
					field.Type,
				),
			})
			continue
		}
		if field.Const != nil && value.Value.Value != *field.Const {
			errors = append(errors, Error{
				Path: relativePath,
				Line: value.Line,
				Message: fmt.Sprintf(
					"frontmatter field '%s'은 '%s'이어야 합니다",
					field.Name,
					*field.Const,
				),
			})
		}
	}
	if !allowsAdditional(schema.AdditionalFields) {
		for name, value := range document.Fields {
			if _, exists := declared[name]; !exists {
				errors = append(errors, Error{
					Path:    relativePath,
					Line:    value.Line,
					Message: fmt.Sprintf("schema에 정의되지 않은 frontmatter field '%s'입니다", name),
				})
			}
		}
	}
	sortErrors(errors)
	return errors
}

func nodeMatchesType(node *yaml.Node, expected string) bool {
	switch expected {
	case "string":
		return node.Kind == yaml.ScalarNode && node.Tag == "!!str"
	case "integer":
		return node.Kind == yaml.ScalarNode && node.Tag == "!!int"
	case "boolean":
		return node.Kind == yaml.ScalarNode && node.Tag == "!!bool"
	case "string-list":
		if node.Kind != yaml.SequenceNode {
			return false
		}
		for _, item := range node.Content {
			if item.Kind != yaml.ScalarNode || item.Tag != "!!str" {
				return false
			}
		}
		return true
	default:
		return false
	}
}
