package lint

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type taskSchema struct {
	Version            int               `yaml:"version"`
	Type               string            `yaml:"type"`
	Frontmatter        frontmatterSchema `yaml:"frontmatter"`
	Sections           []sectionSchema   `yaml:"sections"`
	AdditionalSections *bool             `yaml:"additionalSections"`
}

type frontmatterSchema struct {
	AdditionalFields *bool         `yaml:"additionalFields"`
	Fields           []fieldSchema `yaml:"fields"`
}

type fieldSchema struct {
	Name     string  `yaml:"name"`
	Type     string  `yaml:"type"`
	Required *bool   `yaml:"required"`
	Const    *string `yaml:"const"`
}

type sectionSchema struct {
	Name     string `yaml:"name"`
	Required *bool  `yaml:"required"`
}

func loadSchemas(repositoryRoot string) (map[string]taskSchema, []Error, error) {
	schemasRoot := filepath.Join(repositoryRoot, "tasks", "_schema")
	entries, err := os.ReadDir(schemasRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("read task schemas directory %s: %w", schemasRoot, err)
	}

	schemas := make(map[string]taskSchema)
	var errors []Error
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		taskType := strings.TrimSuffix(entry.Name(), ".yaml")
		path := filepath.Join(schemasRoot, entry.Name())
		relativePath, err := filepath.Rel(repositoryRoot, path)
		if err != nil {
			return nil, nil, err
		}
		if !isValidTaskType(taskType) {
			errors = append(errors, Error{
				Path:    relativePath,
				Line:    1,
				Message: "Schema 파일명은 소문자로 시작하고 소문자, 숫자, hyphen만 포함해야 합니다",
			})
			continue
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, err
		}
		var schema taskSchema
		decoder := yaml.NewDecoder(bytes.NewReader(content))
		decoder.KnownFields(true)
		if err := decoder.Decode(&schema); err != nil {
			errors = append(errors, Error{
				Path:    relativePath,
				Line:    1,
				Message: fmt.Sprintf("YAML schema를 해석할 수 없습니다: %v", err),
			})
			continue
		}

		schemaErrors := validateSchema(repositoryRoot, path, taskType, schema)
		errors = append(errors, schemaErrors...)
		if len(schemaErrors) == 0 {
			schemas[taskType] = schema
		}
	}
	if len(schemas) == 0 && len(errors) == 0 {
		relativePath, err := filepath.Rel(repositoryRoot, schemasRoot)
		if err != nil {
			return nil, nil, err
		}
		errors = append(errors, Error{
			Path:    relativePath,
			Line:    1,
			Message: "'{type}.yaml' 형식의 Task schema가 없습니다",
		})
	}

	sortErrors(errors)
	return schemas, errors, nil
}

func validateSchema(
	repositoryRoot string,
	path string,
	taskType string,
	schema taskSchema,
) []Error {
	relativePath, err := filepath.Rel(repositoryRoot, path)
	if err != nil {
		return []Error{{Path: path, Line: 1, Message: err.Error()}}
	}
	addError := func(errors *[]Error, message string) {
		*errors = append(*errors, Error{Path: relativePath, Line: 1, Message: message})
	}

	var errors []Error
	if schema.Version != 1 {
		addError(&errors, "schema version은 1이어야 합니다")
	}
	if schema.Type != taskType {
		addError(&errors, fmt.Sprintf("schema type은 파일명과 같은 '%s'이어야 합니다", taskType))
	}
	if schema.Frontmatter.AdditionalFields == nil {
		addError(&errors, "frontmatter의 additionalFields를 명시해야 합니다")
	}
	if schema.AdditionalSections == nil {
		addError(&errors, "additionalSections를 명시해야 합니다")
	}

	fieldNames := make(map[string]bool)
	for _, field := range schema.Frontmatter.Fields {
		if field.Name == "" {
			addError(&errors, "frontmatter field의 name이 비어 있습니다")
			continue
		}
		if fieldNames[field.Name] {
			addError(&errors, fmt.Sprintf("frontmatter field '%s'이 중복되었습니다", field.Name))
		}
		fieldNames[field.Name] = true
		if !isSupportedFieldType(field.Type) {
			addError(&errors, fmt.Sprintf("frontmatter field '%s'의 type '%s'을 지원하지 않습니다", field.Name, field.Type))
		}
		if field.Const != nil && field.Type != "string" {
			addError(&errors, fmt.Sprintf("frontmatter field '%s'의 const는 string type에서만 사용할 수 있습니다", field.Name))
		}
		if field.Required == nil {
			addError(&errors, fmt.Sprintf("frontmatter field '%s'의 required를 명시해야 합니다", field.Name))
		}
	}
	typeFieldFound := false
	for _, field := range schema.Frontmatter.Fields {
		if field.Name == "type" && isRequired(field.Required) && field.Type == "string" &&
			field.Const != nil && *field.Const == taskType {
			typeFieldFound = true
		}
	}
	if !typeFieldFound {
		addError(&errors, "frontmatter fields에는 schema type과 같은 const를 가진 필수 string type field가 있어야 합니다")
	}

	sectionNames := make(map[string]bool)
	for _, section := range schema.Sections {
		if section.Name == "" {
			addError(&errors, "section의 name이 비어 있습니다")
			continue
		}
		if sectionNames[section.Name] {
			addError(&errors, fmt.Sprintf("section '%s'이 중복되었습니다", section.Name))
		}
		sectionNames[section.Name] = true
		if section.Required == nil {
			addError(&errors, fmt.Sprintf("section '%s'의 required를 명시해야 합니다", section.Name))
		}
	}
	if len(schema.Sections) == 0 {
		addError(&errors, "sections가 비어 있습니다")
	}

	standardPath := filepath.Join(repositoryRoot, "tasks", "_standard", taskType+".md")
	if info, err := os.Stat(standardPath); err != nil || info.IsDir() {
		addError(&errors, fmt.Sprintf("tasks/_standard/%s.md가 없습니다", taskType))
	}
	return errors
}

func isSupportedFieldType(value string) bool {
	switch value {
	case "string", "integer", "boolean", "string-list":
		return true
	default:
		return false
	}
}

func isRequired(value *bool) bool {
	return value != nil && *value
}

func allowsAdditional(value *bool) bool {
	return value != nil && *value
}

func sortErrors(errors []Error) {
	sort.SliceStable(errors, func(i, j int) bool {
		if errors[i].Path != errors[j].Path {
			return errors[i].Path < errors[j].Path
		}
		return errors[i].Line < errors[j].Line
	})
}
