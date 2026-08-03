package lint

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

const (
	standardFile = "STANDARD.md"
	glossaryFile = "GLOSSARY.md"
	policyFile   = "POLICY.md"
	historyFile  = "HISTORY.md"
)

type Error struct {
	Path    string
	Message string
}

func (e Error) String() string {
	return fmt.Sprintf("%s: %s", filepath.ToSlash(e.Path), e.Message)
}

func Run(repositoryRoot string) ([]Error, error) {
	policyRoot := filepath.Join(repositoryRoot, "policy")
	entries, err := os.ReadDir(policyRoot)
	if err != nil {
		return nil, fmt.Errorf("read policy directory %s: %w", policyRoot, err)
	}

	seenRootFiles := make(map[string]bool)
	var errors []Error
	for _, entry := range entries {
		relativePath := filepath.Join("policy", entry.Name())
		switch entry.Name() {
		case standardFile, glossaryFile:
			seenRootFiles[entry.Name()] = true
			if !isRegularFile(entry) {
				errors = append(errors, Error{
					Path:    relativePath,
					Message: "일반 파일이어야 합니다",
				})
			}
		default:
			if isDirectory(entry) {
				domainErrors, err := lintDomainDirectory(repositoryRoot, relativePath)
				if err != nil {
					return nil, err
				}
				errors = append(errors, domainErrors...)
				continue
			}
			errors = append(errors, Error{
				Path: relativePath,
				Message: fmt.Sprintf(
					"허용되지 않은 항목입니다; %s, %s 또는 도메인 디렉터리만 둘 수 있습니다",
					standardFile,
					glossaryFile,
				),
			})
		}
	}

	for _, required := range []string{standardFile, glossaryFile} {
		if !seenRootFiles[required] {
			errors = append(errors, Error{
				Path:    filepath.Join("policy", required),
				Message: "필수 파일이 없습니다",
			})
		}
	}

	sortErrors(errors)
	return errors, nil
}

func lintDomainDirectory(repositoryRoot string, relativeDirectory string) ([]Error, error) {
	directory := filepath.Join(repositoryRoot, relativeDirectory)
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("read policy domain directory %s: %w", directory, err)
	}

	policyExists := false
	var errors []Error
	for _, entry := range entries {
		relativePath := filepath.Join(relativeDirectory, entry.Name())
		switch entry.Name() {
		case policyFile:
			policyExists = true
			if !isRegularFile(entry) {
				errors = append(errors, Error{
					Path:    relativePath,
					Message: "일반 파일이어야 합니다",
				})
			}
		case historyFile:
			if !isRegularFile(entry) {
				errors = append(errors, Error{
					Path:    relativePath,
					Message: "일반 파일이어야 합니다",
				})
			}
		default:
			errors = append(errors, Error{
				Path: relativePath,
				Message: fmt.Sprintf(
					"허용되지 않은 항목입니다; %s와 선택적인 %s만 둘 수 있습니다",
					policyFile,
					historyFile,
				),
			})
		}
	}

	if !policyExists {
		errors = append(errors, Error{
			Path:    filepath.Join(relativeDirectory, policyFile),
			Message: "필수 파일이 없습니다",
		})
	}

	sortErrors(errors)
	return errors, nil
}

func isDirectory(entry fs.DirEntry) bool {
	if entry.Type()&os.ModeSymlink != 0 {
		return false
	}
	return entry.IsDir()
}

func isRegularFile(entry fs.DirEntry) bool {
	if entry.Type()&os.ModeSymlink != 0 {
		return false
	}
	info, err := entry.Info()
	return err == nil && info.Mode().IsRegular()
}

func sortErrors(errors []Error) {
	sort.SliceStable(errors, func(i, j int) bool {
		if errors[i].Path != errors[j].Path {
			return errors[i].Path < errors[j].Path
		}
		return errors[i].Message < errors[j].Message
	})
}
