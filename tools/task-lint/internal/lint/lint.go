package lint

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Error struct {
	Path    string
	Line    int
	Message string
}

func (e Error) String() string {
	return fmt.Sprintf("%s:%d: %s", filepath.ToSlash(e.Path), e.Line, e.Message)
}

func Run(repositoryRoot string) ([]Error, error) {
	schemas, schemaErrors, err := loadSchemas(repositoryRoot)
	if err != nil {
		return nil, err
	}
	if len(schemaErrors) > 0 {
		return schemaErrors, nil
	}

	var paths []string
	tasksRoot := filepath.Join(repositoryRoot, "tasks")
	for _, status := range []string{"open", "completed"} {
		if err := collectMarkdownFiles(filepath.Join(tasksRoot, status), &paths); err != nil {
			return nil, err
		}
	}
	sort.Strings(paths)

	var errors []Error
	for _, path := range paths {
		relativePath, err := filepath.Rel(repositoryRoot, path)
		if err != nil {
			return nil, err
		}
		frontmatter, frontmatterErrors, err := parseFrontmatter(path)
		if err != nil {
			return nil, fmt.Errorf("read task document %s: %w", path, err)
		}
		for _, item := range frontmatterErrors {
			errors = append(errors, Error{
				Path:    relativePath,
				Line:    item.Line,
				Message: item.Message,
			})
		}
		if len(frontmatterErrors) > 0 {
			continue
		}

		typeValue, exists := frontmatter.Fields["type"]
		if !exists || !nodeMatchesType(typeValue.Value, "string") ||
			!isValidTaskType(typeValue.Value.Value) {
			line := 1
			if exists {
				line = typeValue.Line
			}
			errors = append(errors, Error{
				Path:    relativePath,
				Line:    line,
				Message: "frontmatter의 type은 소문자로 시작하고 소문자, 숫자, hyphen만 포함하는 단일 string 값이어야 합니다",
			})
			continue
		}
		taskType := typeValue.Value.Value
		schema, exists := schemas[taskType]
		if !exists {
			errors = append(errors, Error{
				Path: relativePath,
				Line: typeValue.Line,
				Message: fmt.Sprintf(
					"type '%s'에 해당하는 tasks/_schema/%s.yaml이 없습니다",
					taskType,
					taskType,
				),
			})
			continue
		}

		fieldErrors := lintFrontmatter(relativePath, frontmatter, schema.Frontmatter)
		errors = append(errors, fieldErrors...)

		documentErrors, err := lintDocument(repositoryRoot, path, schema)
		if err != nil {
			return nil, err
		}
		errors = append(errors, documentErrors...)
	}
	sortErrors(errors)
	return errors, nil
}

func collectMarkdownFiles(root string, paths *[]string) error {
	info, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("read task directory %s: %w", root, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("task path is not a directory: %s", root)
	}

	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type().IsRegular() &&
			strings.EqualFold(filepath.Ext(entry.Name()), ".md") {
			*paths = append(*paths, path)
		}
		return nil
	})
}

func lintDocument(
	repositoryRoot string,
	path string,
	schema taskSchema,
) ([]Error, error) {
	sections, err := parseSections(path)
	if err != nil {
		return nil, fmt.Errorf("read task document %s: %w", path, err)
	}

	relativePath, err := filepath.Rel(repositoryRoot, path)
	if err != nil {
		return nil, err
	}

	byName := make(map[string][]section)
	for _, item := range sections {
		byName[item.Name] = append(byName[item.Name], item)
	}

	var errors []Error
	positions := make(map[string]int)
	declared := make(map[string]bool)
	for _, expected := range schema.Sections {
		declared[expected.Name] = true
		matches := byName[expected.Name]
		switch len(matches) {
		case 0:
			if isRequired(expected.Required) {
				errors = append(errors, Error{
					Path:    relativePath,
					Line:    1,
					Message: fmt.Sprintf("필수 섹션 '%s'이 없습니다", expected.Name),
				})
			}
		case 1:
			positions[expected.Name] = matches[0].Line
			if matches[0].Content == "" {
				errors = append(errors, Error{
					Path:    relativePath,
					Line:    matches[0].Line,
					Message: fmt.Sprintf("섹션 '%s'의 내용이 비어 있습니다", expected.Name),
				})
			}
		default:
			for _, duplicate := range matches[1:] {
				errors = append(errors, Error{
					Path: relativePath,
					Line: duplicate.Line,
					Message: fmt.Sprintf(
						"섹션 '%s'이 중복되었습니다 (첫 번째 위치: %d줄)",
						expected.Name,
						matches[0].Line,
					),
				})
			}
		}
	}
	if !allowsAdditional(schema.AdditionalSections) {
		for _, item := range sections {
			if !declared[item.Name] {
				errors = append(errors, Error{
					Path:    relativePath,
					Line:    item.Line,
					Message: fmt.Sprintf("schema에 정의되지 않은 섹션 '%s'입니다", item.Name),
				})
			}
		}
	}

	lastLine := 0
	lastName := ""
	for _, expected := range schema.Sections {
		line, exists := positions[expected.Name]
		if !exists {
			continue
		}
		if line < lastLine {
			errors = append(errors, Error{
				Path: relativePath,
				Line: line,
				Message: fmt.Sprintf(
					"섹션 '%s'은 '%s' 뒤에 있어야 합니다",
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

	sortErrors(errors)
	return errors, nil
}
