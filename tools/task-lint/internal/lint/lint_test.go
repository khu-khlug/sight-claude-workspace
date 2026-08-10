package lint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAcceptsValidDocumentsAndIgnoresOtherDirectories(t *testing.T) {
	root := createRepository(t)
	writeTask(t, root, "open/lf.md", validDocument("\n"))
	writeTask(t, root, "completed/crlf.md", validDocument("\r\n"))
	writeTask(t, root, "group/reference.md", "# Task 형식이 아닌 참고 문서\n")

	assertNoErrors(t, root)
}

func TestRunReportsMissingDuplicateOutOfOrderAndEmptySections(t *testing.T) {
	root := createRepository(t)
	writeTask(t, root, "open/invalid.md", strings.Join([]string{
		"---",
		"type: backend",
		"---",
		"",
		"# Invalid",
		"",
		"## Database",
		"",
		"value",
		"",
		"## Overview",
		"",
		"   ",
		"",
		"## Database",
		"",
		"duplicate",
	}, "\n"))

	errors, err := Run(root)
	if err != nil {
		t.Fatal(err)
	}
	assertMessages(t, errors,
		"섹션 'Overview'의 내용이 비어 있습니다",
		"필수 섹션 'Behavior'이 없습니다",
		"섹션 'Database'이 중복되었습니다",
	)
}

func TestRunValidatesOptionalSectionsWhenPresent(t *testing.T) {
	root := createRepository(t)
	writeSchema(t, root, "backend", schemaWithOptionalSection)
	writeTask(t, root, "open/without-optional.md", validDocument("\n"))
	writeTask(t, root, "completed/with-optional.md", validDocument("\n")+strings.Join([]string{
		"## Notes",
		"",
		"optional value",
	}, "\n"))

	assertNoErrors(t, root)
}

func TestRunChecksSectionOrderAndAdditionalSectionPolicy(t *testing.T) {
	root := createRepository(t)
	writeSchema(t, root, "backend", strings.Replace(
		defaultSchema,
		"additionalSections: true",
		"additionalSections: false",
		1,
	))
	writeTask(t, root, "open/order.md", strings.Join([]string{
		"---",
		"type: backend",
		"---",
		"",
		"# Order",
		"",
		"## Behavior",
		"",
		"value",
		"",
		"## Additional",
		"",
		"value",
		"",
		"## Overview",
		"",
		"value",
		"",
		"## Database",
		"",
		"value",
	}, "\n"))

	errors, err := Run(root)
	if err != nil {
		t.Fatal(err)
	}
	assertMessages(t, errors,
		"섹션 'Behavior'은 'Overview' 뒤에 있어야 합니다",
		"schema에 정의되지 않은 섹션 'Additional'입니다",
	)
}

func TestRunSelectsSchemaAndStandardFromTaskType(t *testing.T) {
	root := createRepository(t)
	writeStandard(t, root, "frontend", "# Frontend Standard\n")
	writeSchema(t, root, "frontend", `version: 1
type: frontend
frontmatter:
  additionalFields: false
  fields:
    - name: type
      type: string
      required: true
      const: frontend
sections:
  - name: UI
    required: true
additionalSections: true
`)
	writeTask(t, root, "open/frontend.md", strings.Join([]string{
		"---",
		"type: frontend",
		"---",
		"",
		"# Frontend",
		"",
		"## UI",
		"",
		"value",
	}, "\n"))

	assertNoErrors(t, root)
}

func TestRunValidatesFrontmatterFieldsFromSchema(t *testing.T) {
	root := createRepository(t)
	writeSchema(t, root, "backend", schemaWithFrontmatterFields)
	writeTask(t, root, "open/invalid-fields.md", strings.Join([]string{
		"---",
		"type: backend",
		"priority: high",
		"labels: backend",
		"unexpected: true",
		"---",
	}, "\n"))

	errors, err := Run(root)
	if err != nil {
		t.Fatal(err)
	}
	assertMessages(t, errors,
		"필수 frontmatter field 'approved'이 없습니다",
		"frontmatter field 'priority'은 integer type이어야 합니다",
		"frontmatter field 'labels'은 string-list type이어야 합니다",
		"schema에 정의되지 않은 frontmatter field 'unexpected'입니다",
	)
}

func TestRunReportsInvalidDuplicateOrUnknownTaskType(t *testing.T) {
	root := createRepository(t)
	writeTask(t, root, "open/missing.md", "# Missing frontmatter\n")
	writeTask(t, root, "open/multiple.md", "---\ntype: [backend, frontend]\n---\n")
	writeTask(t, root, "open/duplicate.md", "---\ntype: backend\ntype: backend\n---\n")
	writeTask(t, root, "open/unknown.md", "---\ntype: frontend\n---\n")

	errors, err := Run(root)
	if err != nil {
		t.Fatal(err)
	}
	assertMessages(t, errors,
		"문서 시작에 YAML frontmatter가 없습니다",
		"type은 소문자로 시작하고 소문자, 숫자, hyphen만 포함하는 단일 string 값이어야 합니다",
		"frontmatter의 'type' field가 중복되었습니다",
		"type 'frontend'에 해당하는 tasks/_schema/frontend.yaml이 없습니다",
	)
}

func TestRunRequiresMatchingStandardForSchema(t *testing.T) {
	root := createRepository(t)
	writeSchema(t, root, "frontend", strings.ReplaceAll(defaultSchema, "backend", "frontend"))

	errors, err := Run(root)
	if err != nil {
		t.Fatal(err)
	}
	assertMessages(t, errors, "tasks/_standard/frontend.md가 없습니다")
}

func TestRunRejectsUnknownSchemaFields(t *testing.T) {
	root := createRepository(t)
	writeSchema(t, root, "backend", strings.Replace(
		defaultSchema,
		"required: true",
		"requred: true",
		1,
	))

	errors, err := Run(root)
	if err != nil {
		t.Fatal(err)
	}
	assertMessages(t, errors, "field requred not found")
}

func TestRunRequiresExplicitRequiredAndAdditionalPolicies(t *testing.T) {
	root := createRepository(t)
	writeSchema(t, root, "backend", `version: 1
type: backend
frontmatter:
  fields:
    - name: type
      type: string
      const: backend
sections:
  - name: Overview
`)

	errors, err := Run(root)
	if err != nil {
		t.Fatal(err)
	}
	assertMessages(t, errors,
		"frontmatter의 additionalFields를 명시해야 합니다",
		"frontmatter field 'type'의 required를 명시해야 합니다",
		"section 'Overview'의 required를 명시해야 합니다",
		"additionalSections를 명시해야 합니다",
	)
}

func TestRunIgnoresHeadingsInsideFencedCodeBlocks(t *testing.T) {
	root := createRepository(t)
	document := validDocument("\n") + strings.Join([]string{
		"## Additional",
		"",
		"```markdown",
		"## Overview",
		"```",
	}, "\n")
	writeTask(t, root, "open/fence.md", document)

	assertNoErrors(t, root)
}

func TestRunIgnoresHeadingsInsideFrontmatter(t *testing.T) {
	root := createRepository(t)
	writeSchema(t, root, "backend", strings.Replace(
		schemaWithFrontmatterFields,
		"    - name: priority\n      type: integer\n      required: false",
		"    - name: description\n      type: string\n      required: false",
		1,
	))
	writeTask(t, root, "open/frontmatter-heading.md", strings.Join([]string{
		"---",
		"type: backend",
		"approved: true",
		"description: |",
		"  ## Overview",
		"---",
	}, "\n"))

	assertNoErrors(t, root)
}

func TestRunAcceptsStructuredSectionLiteralAndRecords(t *testing.T) {
	root := createRepository(t)
	writeSchema(t, root, "backend", schemaWithStructuredSection)
	writeTask(t, root, "open/literal.md", strings.Join([]string{
		"---",
		"type: backend",
		"---",
		"",
		"# Literal",
		"",
		"## Policies",
		"",
		"No changes",
	}, "\n"))
	writeTask(t, root, "completed/records.md", strings.Join([]string{
		"---",
		"type: backend",
		"---",
		"",
		"# Records",
		"",
		"## Policies",
		"",
		"### First policy",
		"",
		"- Actor: manager",
		"- Decision: Denied",
		"- Recovery: wait for completion",
		"",
		"### Second policy",
		"",
		"- Actor: member",
		"- Decision: Allowed",
		"- Recovery: None",
	}, "\n"))

	assertNoErrors(t, root)
}

func TestRunValidatesStructuredSectionRecords(t *testing.T) {
	root := createRepository(t)
	writeSchema(t, root, "backend", schemaWithStructuredSection)
	writeTask(t, root, "open/invalid-records.md", strings.Join([]string{
		"---",
		"type: backend",
		"---",
		"",
		"# Invalid records",
		"",
		"## Policies",
		"",
		"### Duplicate policy",
		"",
		"- Decision: Sometimes",
		"- Actor: manager",
		"- Actor: operator",
		"- Unexpected: value",
		"- Recovery:",
		"",
		"### Duplicate policy",
		"",
		"- Actor: member",
		"- Decision: Allowed",
		"",
		"### Out of order",
		"",
		"- Decision: Denied",
		"- Actor: member",
		"- Recovery: None",
	}, "\n"))

	errors, err := Run(root)
	if err != nil {
		t.Fatal(err)
	}
	assertMessages(t, errors,
		"record 'Duplicate policy' field 'Actor'이 중복되었습니다",
		"field 'Decision'은 다음 값 중 하나여야 합니다: Allowed, Denied",
		"schema로 정의되지 않은 field 'Unexpected'",
		"field 'Recovery'의 값이 비어 있습니다",
		"record 'Duplicate policy'에 필수 field 'Recovery'이 없습니다",
		"record 'Duplicate policy'이 중복되었습니다",
		"record 'Out of order' field 'Decision'은 'Actor' 뒤에 있어야 합니다",
	)
}

func TestRunRejectsMalformedStructuredSectionContent(t *testing.T) {
	root := createRepository(t)
	writeSchema(t, root, "backend", schemaWithStructuredSection)
	writeTask(t, root, "open/malformed-records.md", strings.Join([]string{
		"---",
		"type: backend",
		"---",
		"",
		"# Malformed records",
		"",
		"## Policies",
		"",
		"No changes with additional text",
	}, "\n"))

	errors, err := Run(root)
	if err != nil {
		t.Fatal(err)
	}
	assertMessages(t, errors,
		"records 내용은 level 3 제목 또는 '- 이름: 값' 형식이어야 합니다",
		"level 3 제목의 record가 최소 1개 있어야 합니다",
	)
}

func TestRunValidatesStructuredSectionSchema(t *testing.T) {
	root := createRepository(t)
	writeSchema(t, root, "backend", `version: 1
type: backend
frontmatter:
  additionalFields: false
  fields:
    - name: type
      type: string
      required: true
      const: backend
sections:
  - name: Policies
    required: true
    content:
      oneOf:
        - literal: ""
          minItems: 1
        - type: records
          headingLevel: 2
          minItems: 0
          fields:
            - name: Decision
              type: enum
            - name: Note
              type: string
              required: true
              values: [value]
additionalSections: true
`)

	errors, err := Run(root)
	if err != nil {
		t.Fatal(err)
	}
	assertMessages(t, errors,
		"literal이 비어 있습니다",
		"literal variant에는 records 설정을 사용할 수 없습니다",
		"headingLevel은 3 이상 6 이하여야 합니다",
		"minItems는 1 이상이어야 합니다",
		"additionalFields를 명시해야 합니다",
		"field 'Decision'의 required를 명시해야 합니다",
		"field 'Decision'의 values가 비어 있습니다",
		"field 'Note'의 values는 enum type에서만 사용할 수 있습니다",
	)
}

const defaultStandard = `# Standard

Human-readable guidance.
`

const defaultSchema = `version: 1
type: backend
frontmatter:
  additionalFields: false
  fields:
    - name: type
      type: string
      required: true
      const: backend
sections:
  - name: Overview
    required: true
  - name: Behavior
    required: true
  - name: Database
    required: true
additionalSections: true
`

const schemaWithOptionalSection = `version: 1
type: backend
frontmatter:
  additionalFields: false
  fields:
    - name: type
      type: string
      required: true
      const: backend
sections:
  - name: Overview
    required: true
  - name: Behavior
    required: true
  - name: Database
    required: true
  - name: Notes
    required: false
additionalSections: true
`

const schemaWithFrontmatterFields = `version: 1
type: backend
frontmatter:
  additionalFields: false
  fields:
    - name: type
      type: string
      required: true
      const: backend
    - name: priority
      type: integer
      required: false
    - name: approved
      type: boolean
      required: true
    - name: labels
      type: string-list
      required: false
sections:
  - name: Overview
    required: false
additionalSections: true
`

const schemaWithStructuredSection = `version: 1
type: backend
frontmatter:
  additionalFields: false
  fields:
    - name: type
      type: string
      required: true
      const: backend
sections:
  - name: Policies
    required: true
    content:
      oneOf:
        - literal: No changes
        - type: records
          headingLevel: 3
          minItems: 1
          additionalFields: false
          fields:
            - name: Actor
              type: string
              required: true
            - name: Decision
              type: enum
              required: true
              values: [Allowed, Denied]
            - name: Recovery
              type: string
              required: true
additionalSections: true
`

func validDocument(newline string) string {
	return strings.Join([]string{
		"---",
		"type: backend",
		"---",
		"",
		"# Valid",
		"",
		"## Overview",
		"",
		"overview",
		"",
		"## Behavior",
		"",
		"behavior",
		"",
		"## Database",
		"",
		"database",
		"",
	}, newline)
}

func createRepository(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, directory := range []string{
		filepath.Join(root, "tasks", "_schema"),
		filepath.Join(root, "tasks", "_standard"),
		filepath.Join(root, "tasks", "open"),
		filepath.Join(root, "tasks", "completed"),
	} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeSchema(t, root, "backend", defaultSchema)
	writeStandard(t, root, "backend", defaultStandard)
	return root
}

func writeSchema(t *testing.T, root string, taskType string, content string) {
	t.Helper()
	path := filepath.Join(root, "tasks", "_schema", taskType+".yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeStandard(t *testing.T, root string, taskType string, content string) {
	t.Helper()
	path := filepath.Join(root, "tasks", "_standard", taskType+".md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeTask(t *testing.T, root string, relativePath string, content string) {
	t.Helper()
	path := filepath.Join(root, "tasks", relativePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertNoErrors(t *testing.T, root string) {
	t.Helper()
	errors, err := Run(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(errors) != 0 {
		t.Fatalf("expected no errors, got: %v", errors)
	}
}

func assertMessages(t *testing.T, errors []Error, expected ...string) {
	t.Helper()
	messages := errorMessages(errors)
	for _, message := range expected {
		if !strings.Contains(messages, message) {
			t.Errorf("expected %q in errors:\n%s", message, messages)
		}
	}
}

func errorMessages(errors []Error) string {
	var messages []string
	for _, item := range errors {
		messages = append(messages, item.String())
	}
	return strings.Join(messages, "\n")
}
