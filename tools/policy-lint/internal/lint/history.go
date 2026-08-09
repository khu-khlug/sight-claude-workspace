package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

const historyHeadingLevel = 2

var (
	historyHeadingPattern = regexp.MustCompile(`^(.+?) \((\d{4}-\d{2}-\d{2})\)$`)
	historyMarkdown       = goldmark.New(goldmark.WithExtensions(extension.Table))
)

type historyDocumentLinter struct {
	relativePath string
	source       []byte
	document     ast.Node
	errors       []Error

	entryCount         int
	currentEntry       string
	currentBodyCount   int
	previousValidDate  time.Time
	hasPreviousDate    bool
	contentBeforeEntry bool
}

func lintHistoryDocument(repositoryRoot string, relativePath string) ([]Error, error) {
	absolutePath := filepath.Join(repositoryRoot, relativePath)
	source, err := os.ReadFile(absolutePath)
	if err != nil {
		return nil, fmt.Errorf("read policy history document %s: %w", absolutePath, err)
	}

	document := historyMarkdown.Parser().Parse(text.NewReader(source))
	linter := &historyDocumentLinter{
		relativePath: relativePath,
		source:       source,
		document:     document,
	}
	linter.lintStructure()
	return linter.errors, nil
}

func (l *historyDocumentLinter) lintStructure() {
	for node := l.document.FirstChild(); node != nil; node = node.NextSibling() {
		if heading, ok := node.(*ast.Heading); ok {
			if heading.Level == historyHeadingLevel {
				l.finishEntry()
				l.startEntry(heading)
				continue
			}

			l.addError("변경 기록을 구분하는 H2 이외의 heading은 사용할 수 없습니다")
			continue
		}

		l.lintNestedHeadings(node)
		if l.entryCount == 0 {
			if !l.contentBeforeEntry {
				l.addError("첫 H2 변경 기록 앞에는 본문을 작성할 수 없습니다")
				l.contentBeforeEntry = true
			}
			continue
		}
		l.currentBodyCount++
	}

	l.finishEntry()
	if l.entryCount == 0 {
		l.addError("하나 이상의 H2 변경 기록이 필요합니다")
	}
}

func (l *historyDocumentLinter) startEntry(heading *ast.Heading) {
	title := strings.TrimSpace(string(heading.Text(l.source)))
	l.entryCount++
	l.currentEntry = title
	l.currentBodyCount = 0

	matches := historyHeadingPattern.FindStringSubmatch(title)
	if matches == nil || strings.TrimSpace(matches[1]) == "" || matches[1] != strings.TrimSpace(matches[1]) {
		l.addError(fmt.Sprintf("H2 변경 기록은 '변경 내용 (YYYY-MM-DD)' 형식이어야 합니다: %q", title))
		return
	}

	date, err := time.Parse("2006-01-02", matches[2])
	if err != nil {
		l.addError(fmt.Sprintf("H2 변경 기록에 유효한 날짜가 필요합니다: %q", title))
		return
	}
	if l.hasPreviousDate && date.Before(l.previousValidDate) {
		l.addError(fmt.Sprintf("변경 기록은 날짜 오름차순으로 작성해야 합니다: %q", title))
	}
	l.previousValidDate = date
	l.hasPreviousDate = true
}

func (l *historyDocumentLinter) finishEntry() {
	if l.entryCount == 0 {
		return
	}
	if l.currentBodyCount == 0 {
		l.addError(fmt.Sprintf("H2 변경 기록의 본문은 비어 있을 수 없습니다: %q", l.currentEntry))
	}
}

func (l *historyDocumentLinter) lintNestedHeadings(node ast.Node) {
	_ = ast.Walk(node, func(current ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if _, ok := current.(*ast.Heading); ok {
				l.addError("변경 기록을 구분하는 최상위 H2 이외의 heading은 사용할 수 없습니다")
			}
		}
		return ast.WalkContinue, nil
	})
}

func (l *historyDocumentLinter) addError(message string) {
	l.errors = append(l.errors, Error{Path: l.relativePath, Message: message})
}
