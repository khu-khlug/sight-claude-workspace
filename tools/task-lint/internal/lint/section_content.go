package lint

import (
	"fmt"
	"strings"
)

type contentRecord struct {
	Name   string
	Line   int
	Fields []contentRecordField
}

type contentRecordField struct {
	Name  string
	Value string
	Line  int
}

func lintSectionContent(
	relativePath string,
	item section,
	schema sectionContentSchema,
) []Error {
	for _, variant := range schema.OneOf {
		if variant.Literal != nil && item.Content == *variant.Literal {
			return nil
		}
	}

	for _, variant := range schema.OneOf {
		if variant.Type == "records" {
			return lintRecordsContent(relativePath, item, variant)
		}
	}

	return []Error{{
		Path:    relativePath,
		Line:    item.Line,
		Message: fmt.Sprintf("섹션 '%s'의 내용이 schema에 정의된 형식과 일치하지 않습니다", item.Name),
	}}
}

func lintRecordsContent(
	relativePath string,
	item section,
	schema sectionContentVariantSchema,
) []Error {
	records, errors := parseContentRecords(relativePath, item, *schema.HeadingLevel)
	if len(records) < *schema.MinItems {
		errors = append(errors, Error{
			Path: relativePath,
			Line: item.Line,
			Message: fmt.Sprintf(
				"섹션 '%s'에는 level %d 제목의 record가 최소 %d개 있어야 합니다",
				item.Name,
				*schema.HeadingLevel,
				*schema.MinItems,
			),
		})
	}

	recordNames := make(map[string]contentRecord)
	for _, record := range records {
		if first, exists := recordNames[record.Name]; exists {
			errors = append(errors, Error{
				Path: relativePath,
				Line: record.Line,
				Message: fmt.Sprintf(
					"섹션 '%s'의 record '%s'이 중복되었습니다 (첫 번째 위치: %d줄)",
					item.Name,
					record.Name,
					first.Line,
				),
			})
		} else {
			recordNames[record.Name] = record
		}
		errors = append(errors, lintContentRecord(relativePath, item.Name, record, schema)...)
	}

	sortErrors(errors)
	return errors
}

func parseContentRecords(
	relativePath string,
	item section,
	headingLevel int,
) ([]contentRecord, []Error) {
	lines := strings.Split(item.RawContent, "\n")
	var records []contentRecord
	var current *contentRecord
	var activeFence *fence
	var errors []Error

	for index, line := range lines {
		lineNumber := item.Line + 1 + index
		trimmed := strings.TrimSpace(line)
		if activeFence != nil {
			if marker, ok := fenceMarker(line); ok && marker.char == activeFence.char &&
				marker.count >= activeFence.count && isClosingFence(line, marker) {
				activeFence = nil
			}
			continue
		}
		if marker, ok := fenceMarker(line); ok {
			activeFence = &marker
			errors = append(errors, Error{
				Path:    relativePath,
				Line:    lineNumber,
				Message: fmt.Sprintf("섹션 '%s'의 records 형식에는 fenced code block을 사용할 수 없습니다", item.Name),
			})
			continue
		}
		if trimmed == "" {
			continue
		}

		if name, ok := heading(line, headingLevel); ok {
			records = append(records, contentRecord{Name: name, Line: lineNumber})
			current = &records[len(records)-1]
			continue
		}

		field, ok := parseContentRecordField(line, lineNumber)
		if !ok {
			errors = append(errors, Error{
				Path: relativePath,
				Line: lineNumber,
				Message: fmt.Sprintf(
					"섹션 '%s'의 records 내용은 level %d 제목 또는 '- 이름: 값' 형식이어야 합니다",
					item.Name,
					headingLevel,
				),
			})
			continue
		}
		if current == nil {
			errors = append(errors, Error{
				Path: relativePath,
				Line: lineNumber,
				Message: fmt.Sprintf(
					"섹션 '%s'의 field '%s' 앞에 level %d record 제목이 필요합니다",
					item.Name,
					field.Name,
					headingLevel,
				),
			})
			continue
		}
		current.Fields = append(current.Fields, field)
	}

	if activeFence != nil {
		errors = append(errors, Error{
			Path:    relativePath,
			Line:    item.Line,
			Message: fmt.Sprintf("섹션 '%s'에 닫히지 않은 fenced code block이 있습니다", item.Name),
		})
	}
	return records, errors
}

func parseContentRecordField(line string, lineNumber int) (contentRecordField, bool) {
	trimmed := trimUpToThreeLeadingSpaces(line)
	if len(trimmed) < 3 || trimmed[0] != '-' ||
		(trimmed[1] != ' ' && trimmed[1] != '\t') {
		return contentRecordField{}, false
	}

	field := strings.TrimSpace(trimmed[2:])
	separator := strings.Index(field, ":")
	if separator < 1 {
		return contentRecordField{}, false
	}
	name := strings.TrimSpace(field[:separator])
	if name == "" {
		return contentRecordField{}, false
	}
	return contentRecordField{
		Name:  name,
		Value: strings.TrimSpace(field[separator+1:]),
		Line:  lineNumber,
	}, true
}

func lintContentRecord(
	relativePath string,
	sectionName string,
	record contentRecord,
	schema sectionContentVariantSchema,
) []Error {
	byName := make(map[string][]contentRecordField)
	for _, field := range record.Fields {
		byName[field.Name] = append(byName[field.Name], field)
	}

	var errors []Error
	positions := make(map[string]int)
	declared := make(map[string]bool)
	for _, expected := range schema.Fields {
		declared[expected.Name] = true
		matches := byName[expected.Name]
		switch len(matches) {
		case 0:
			if isRequired(expected.Required) {
				errors = append(errors, Error{
					Path: relativePath,
					Line: record.Line,
					Message: fmt.Sprintf(
						"섹션 '%s'의 record '%s'에 필수 field '%s'이 없습니다",
						sectionName,
						record.Name,
						expected.Name,
					),
				})
			}
		case 1:
			positions[expected.Name] = matches[0].Line
			if matches[0].Value == "" {
				errors = append(errors, Error{
					Path: relativePath,
					Line: matches[0].Line,
					Message: fmt.Sprintf(
						"섹션 '%s'의 record '%s' field '%s'의 값이 비어 있습니다",
						sectionName,
						record.Name,
						expected.Name,
					),
				})
			} else if expected.Type == "enum" && !contains(expected.Values, matches[0].Value) {
				errors = append(errors, Error{
					Path: relativePath,
					Line: matches[0].Line,
					Message: fmt.Sprintf(
						"섹션 '%s'의 record '%s' field '%s'은 다음 값 중 하나여야 합니다: %s",
						sectionName,
						record.Name,
						expected.Name,
						strings.Join(expected.Values, ", "),
					),
				})
			}
		default:
			for _, duplicate := range matches[1:] {
				errors = append(errors, Error{
					Path: relativePath,
					Line: duplicate.Line,
					Message: fmt.Sprintf(
						"섹션 '%s'의 record '%s' field '%s'이 중복되었습니다 (첫 번째 위치: %d줄)",
						sectionName,
						record.Name,
						expected.Name,
						matches[0].Line,
					),
				})
			}
		}
	}

	if !allowsAdditional(schema.AdditionalFields) {
		for _, field := range record.Fields {
			if !declared[field.Name] {
				errors = append(errors, Error{
					Path: relativePath,
					Line: field.Line,
					Message: fmt.Sprintf(
						"섹션 '%s'의 record '%s'에 schema로 정의되지 않은 field '%s'이 있습니다",
						sectionName,
						record.Name,
						field.Name,
					),
				})
			}
		}
	}

	lastLine := 0
	lastName := ""
	for _, expected := range schema.Fields {
		line, exists := positions[expected.Name]
		if !exists {
			continue
		}
		if line < lastLine {
			errors = append(errors, Error{
				Path: relativePath,
				Line: line,
				Message: fmt.Sprintf(
					"섹션 '%s'의 record '%s' field '%s'은 '%s' 뒤에 있어야 합니다",
					sectionName,
					record.Name,
					expected.Name,
					lastName,
				),
			})
		}
		if line > lastLine {
			lastLine = line
			lastName = expected.Name
		}
	}

	return errors
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
