package lint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAcceptsRootFilesAndValidDomainStructures(t *testing.T) {
	root := createRepository(t)
	writeFile(t, root, "policy/회원/POLICY.md")
	writeFile(t, root, "policy/행사/POLICY.md")
	writeHistoryFile(t, root, "policy/행사/HISTORY.md")
	writeFile(t, root, "policy/english_domain/POLICY.md")

	assertNoErrors(t, root)
}

func TestRunRequiresStandardAndGlossaryFiles(t *testing.T) {
	root := createRepositoryWithoutRootFiles(t)

	errors, err := Run(root)
	if err != nil {
		t.Fatal(err)
	}
	assertMessages(t, errors,
		"policy/GLOSSARY.md: 필수 파일이 없습니다",
		"policy/STANDARD.md: 필수 파일이 없습니다",
	)
}

func TestRunRejectsUnexpectedRootEntries(t *testing.T) {
	root := createRepository(t)
	writeFile(t, root, "policy/README.md")
	makeSymlink(t, root, "policy/STANDARD.md", "policy/standard-link.md")

	errors, err := Run(root)
	if err != nil {
		t.Fatal(err)
	}
	assertMessages(t, errors,
		"policy/README.md: 허용되지 않은 항목입니다",
		"policy/standard-link.md: 허용되지 않은 항목입니다",
	)
}

func TestRunRequiresPolicyAndAllowsOptionalHistory(t *testing.T) {
	root := createRepository(t)
	makeDirectory(t, root, "policy/회원")
	writeFile(t, root, "policy/행사/POLICY.md")
	writeHistoryFile(t, root, "policy/행사/HISTORY.md")

	errors, err := Run(root)
	if err != nil {
		t.Fatal(err)
	}
	assertMessages(t, errors, "policy/회원/POLICY.md: 필수 파일이 없습니다")
}

func TestRunRejectsUnexpectedDomainEntries(t *testing.T) {
	root := createRepository(t)
	writeFile(t, root, "policy/행사/POLICY.md")
	writeFile(t, root, "policy/행사/README.md")
	makeDirectory(t, root, "policy/행사/decisions")
	makeSymlink(t, root, "policy/행사/POLICY.md", "policy/행사/policy-link.md")

	errors, err := Run(root)
	if err != nil {
		t.Fatal(err)
	}
	assertMessages(t, errors,
		"policy/행사/README.md: 허용되지 않은 항목입니다",
		"policy/행사/decisions: 허용되지 않은 항목입니다",
		"policy/행사/policy-link.md: 허용되지 않은 항목입니다",
	)
}

func TestRunRequiresAllowedNamesToBeRegularFiles(t *testing.T) {
	root := createRepositoryWithoutRootFiles(t)
	makeDirectory(t, root, "policy/STANDARD.md")
	writeFile(t, root, "policy/glossary-target.md")
	makeSymlink(t, root, "policy/glossary-target.md", "policy/GLOSSARY.md")
	makeDirectory(t, root, "policy/행사")
	makeDirectory(t, root, "policy/행사/POLICY.md")
	makeDirectory(t, root, "policy/행사/HISTORY.md")

	errors, err := Run(root)
	if err != nil {
		t.Fatal(err)
	}
	assertMessages(t, errors,
		"policy/GLOSSARY.md: 일반 파일이어야 합니다",
		"policy/STANDARD.md: 일반 파일이어야 합니다",
		"policy/행사/HISTORY.md: 일반 파일이어야 합니다",
		"policy/행사/POLICY.md: 일반 파일이어야 합니다",
	)
}

func createRepository(t *testing.T) string {
	t.Helper()
	root := createRepositoryWithoutRootFiles(t)
	writeFile(t, root, "policy/STANDARD.md")
	writeFile(t, root, "policy/GLOSSARY.md")
	return root
}

func createRepositoryWithoutRootFiles(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	makeDirectory(t, root, "policy")
	return root
}

func writeFile(t *testing.T, root string, relativePath string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("# document\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeHistoryFile(t *testing.T, root string, relativePath string) {
	t.Helper()
	writeContent(t, root, relativePath, "## 정책 원칙 변경 (2026-08-09)\n\n변경 내용이다.\n")
}

func makeDirectory(t *testing.T, root string, relativePath string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func makeSymlink(t *testing.T, root string, target string, link string) {
	t.Helper()
	targetPath := filepath.Join(root, filepath.FromSlash(target))
	linkPath := filepath.Join(root, filepath.FromSlash(link))
	if err := os.MkdirAll(filepath.Dir(linkPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(targetPath, linkPath); err != nil {
		t.Fatal(err)
	}
}

func assertNoErrors(t *testing.T, root string) {
	t.Helper()
	errors, err := Run(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(errors) > 0 {
		t.Fatalf("unexpected errors: %v", errors)
	}
}

func assertMessages(t *testing.T, errors []Error, expected ...string) {
	t.Helper()
	joined := make([]string, 0, len(errors))
	for _, item := range errors {
		joined = append(joined, item.String())
	}
	actual := strings.Join(joined, "\n")
	for _, message := range expected {
		if !strings.Contains(actual, message) {
			t.Errorf("expected error containing %q, got:\n%s", message, actual)
		}
	}
}
