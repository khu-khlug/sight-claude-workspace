package lint

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extensionast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

const (
	rootHeadingLevel    = 1
	sectionHeadingLevel = 2
	policyHeadingLevel  = 3

	scopeSection  = "적용 범위"
	policySection = "정책"
)

var policyMarkdown = goldmark.New(goldmark.WithExtensions(extension.Table))

type policyDocumentLinter struct {
	repositoryRoot string
	relativePath   string
	absolutePath   string
	source         []byte
	document       ast.Node
	errors         []Error

	currentSection       string
	sectionContentCount  int
	policyItemCount      int
	policyItemActive     bool
	currentPolicyHeading string
	currentPolicyHasList bool
	seenSections         map[string]bool
	seenPolicyHeadings   map[string]bool
}

func lintPolicyDocument(repositoryRoot string, relativePath string) ([]Error, error) {
	absolutePath := filepath.Join(repositoryRoot, relativePath)
	source, err := os.ReadFile(absolutePath)
	if err != nil {
		return nil, fmt.Errorf("read policy document %s: %w", absolutePath, err)
	}

	document := policyMarkdown.Parser().Parse(text.NewReader(source))
	linter := &policyDocumentLinter{
		repositoryRoot:     repositoryRoot,
		relativePath:       relativePath,
		absolutePath:       absolutePath,
		source:             source,
		document:           document,
		seenSections:       make(map[string]bool),
		seenPolicyHeadings: make(map[string]bool),
	}
	linter.lintStructure()
	linter.lintLinks()
	return linter.errors, nil
}

func (l *policyDocumentLinter) lintStructure() {
	first := l.document.FirstChild()
	if first == nil {
		l.addError("비어 있지 않은 H1 제목이 필요합니다")
		return
	}
	if heading, ok := first.(*ast.Heading); !ok || heading.Level != rootHeadingLevel {
		l.addError("문서의 첫 번째 내용은 H1 제목이어야 합니다")
	}

	h1Count := 0
	for node := first; node != nil; node = node.NextSibling() {
		heading, isHeading := node.(*ast.Heading)
		if isHeading {
			switch heading.Level {
			case rootHeadingLevel:
				h1Count++
				if strings.TrimSpace(string(heading.Text(l.source))) == "" {
					l.addError("H1 제목은 비어 있을 수 없습니다")
				}
				if node != first {
					l.addError("H1 제목은 문서 최상단에 하나만 존재해야 합니다")
				}
			case sectionHeadingLevel:
				l.finishSection()
				l.startSection(heading)
			case policyHeadingLevel:
				l.startPolicyItem(heading)
			default:
				l.addError("H4 이하 제목은 사용할 수 없습니다")
			}
			continue
		}

		l.lintSectionBlock(node)
	}

	l.finishSection()
	if h1Count == 0 {
		l.addError("H1 제목이 필요합니다")
	} else if h1Count > 1 {
		l.addError("H1 제목은 정확히 하나만 존재해야 합니다")
	}
}

func (l *policyDocumentLinter) startSection(heading *ast.Heading) {
	name := strings.TrimSpace(string(heading.Text(l.source)))
	l.currentSection = name
	l.sectionContentCount = 0
	l.policyItemCount = 0
	l.policyItemActive = false
	l.currentPolicyHeading = ""
	l.currentPolicyHasList = false

	if name != scopeSection && name != policySection {
		l.addError(fmt.Sprintf("허용되지 않은 H2 섹션입니다: %q", name))
		return
	}
	if l.seenSections[name] {
		l.addError(fmt.Sprintf("H2 섹션은 한 번만 사용할 수 있습니다: %q", name))
	}
	if name == scopeSection && l.seenSections[policySection] {
		l.addError(fmt.Sprintf("%q 섹션은 %q 섹션보다 앞에 있어야 합니다", scopeSection, policySection))
	}
	l.seenSections[name] = true
}

func (l *policyDocumentLinter) finishSection() {
	l.finishPolicyItem()
	switch l.currentSection {
	case scopeSection:
		if l.sectionContentCount == 0 {
			l.addError(fmt.Sprintf("%q 섹션은 비어 있을 수 없습니다", scopeSection))
		}
	case policySection:
		if l.policyItemCount == 0 {
			l.addError(fmt.Sprintf("%q 섹션에는 하나 이상의 H3 정책 항목이 필요합니다", policySection))
		}
	}
	l.currentSection = ""
}

func (l *policyDocumentLinter) startPolicyItem(heading *ast.Heading) {
	if l.currentSection != policySection {
		l.addError(fmt.Sprintf("H3 정책 항목은 %q 섹션 안에서만 사용할 수 있습니다", policySection))
		return
	}

	l.finishPolicyItem()
	title := strings.TrimSpace(string(heading.Text(l.source)))
	l.currentPolicyHeading = title
	l.policyItemActive = true
	l.currentPolicyHasList = false
	l.policyItemCount++
	if title == "" {
		l.addError("H3 정책 항목 제목은 비어 있을 수 없습니다")
		return
	}
	if l.seenPolicyHeadings[title] {
		l.addError(fmt.Sprintf("H3 정책 항목 제목은 중복될 수 없습니다: %q", title))
	}
	l.seenPolicyHeadings[title] = true
}

func (l *policyDocumentLinter) finishPolicyItem() {
	if !l.policyItemActive {
		return
	}
	if !l.currentPolicyHasList {
		l.addError(fmt.Sprintf("H3 정책 항목에는 하나 이상의 '-' unordered list가 필요합니다: %q", l.currentPolicyHeading))
	}
	l.policyItemActive = false
	l.currentPolicyHeading = ""
	l.currentPolicyHasList = false
}

func (l *policyDocumentLinter) lintSectionBlock(node ast.Node) {
	switch l.currentSection {
	case scopeSection:
		l.sectionContentCount++
		if _, ok := node.(*ast.Paragraph); !ok {
			l.addError(fmt.Sprintf("%q 섹션에는 일반 문단만 사용할 수 있습니다", scopeSection))
		}
	case policySection:
		if !l.policyItemActive {
			l.addError(fmt.Sprintf("%q 섹션의 내용은 H3 정책 항목 아래에 작성해야 합니다", policySection))
			return
		}
		l.lintPolicyBlock(node)
	default:
		l.addError("H1 아래에는 허용된 H2 섹션만 작성할 수 있습니다")
	}
}

func (l *policyDocumentLinter) lintPolicyBlock(node ast.Node) {
	switch typed := node.(type) {
	case *ast.List:
		if !typed.IsOrdered() {
			l.currentPolicyHasList = true
		}
		l.lintList(typed)
	case *extensionast.Table:
		l.lintTable(typed)
	case *ast.FencedCodeBlock:
		l.lintMermaid(typed)
	default:
		l.addError(fmt.Sprintf(
			"H3 정책 항목에는 '-' unordered list, 표와 Mermaid code block만 사용할 수 있습니다: %q",
			l.currentPolicyHeading,
		))
	}
}

func (l *policyDocumentLinter) lintList(list *ast.List) {
	_ = ast.Walk(list, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch typed := node.(type) {
		case *ast.List:
			if typed.IsOrdered() {
				l.addError(fmt.Sprintf("ordered list는 사용할 수 없습니다: %q", l.currentPolicyHeading))
			} else if typed.Marker != '-' {
				l.addError(fmt.Sprintf("unordered list marker는 '-'만 사용할 수 있습니다: %q", l.currentPolicyHeading))
			}
		case *ast.ListItem:
			l.lintListItem(typed)
		}
		return ast.WalkContinue, nil
	})
}

func (l *policyDocumentLinter) lintListItem(item *ast.ListItem) {
	hasContent := false
	for child := item.FirstChild(); child != nil; child = child.NextSibling() {
		switch child.(type) {
		case *ast.List:
			continue
		case *ast.Paragraph, *ast.TextBlock:
			if strings.TrimSpace(string(child.Text(l.source))) != "" {
				hasContent = true
			}
		default:
			l.addError(fmt.Sprintf("list item에는 문장과 중첩 unordered list만 사용할 수 있습니다: %q", l.currentPolicyHeading))
		}
	}
	if !hasContent {
		l.addError(fmt.Sprintf("list item은 비어 있을 수 없습니다: %q", l.currentPolicyHeading))
	}
}

func (l *policyDocumentLinter) lintTable(table *extensionast.Table) {
	headerCount := 0
	dataRowCount := 0
	for child := table.FirstChild(); child != nil; child = child.NextSibling() {
		switch typed := child.(type) {
		case *extensionast.TableHeader:
			headerCount++
			l.lintTableCells(typed)
		case *extensionast.TableRow:
			dataRowCount++
			l.lintTableCells(typed)
		}
	}
	if headerCount != 1 {
		l.addError(fmt.Sprintf("표에는 비어 있지 않은 header row가 하나 필요합니다: %q", l.currentPolicyHeading))
	}
	if dataRowCount == 0 {
		l.addError(fmt.Sprintf("표에는 비어 있지 않은 data row가 하나 이상 필요합니다: %q", l.currentPolicyHeading))
	}
}

func (l *policyDocumentLinter) lintTableCells(row ast.Node) {
	cellCount := 0
	for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
		cellCount++
		if strings.TrimSpace(string(cell.Text(l.source))) == "" {
			l.addError(fmt.Sprintf("표의 cell은 비어 있을 수 없습니다: %q", l.currentPolicyHeading))
		}
	}
	if cellCount == 0 {
		l.addError(fmt.Sprintf("표의 row는 비어 있을 수 없습니다: %q", l.currentPolicyHeading))
	}
}

func (l *policyDocumentLinter) lintMermaid(block *ast.FencedCodeBlock) {
	if string(block.Language(l.source)) != "mermaid" {
		l.addError(fmt.Sprintf("Mermaid 이외의 code block은 사용할 수 없습니다: %q", l.currentPolicyHeading))
		return
	}

	var content bytes.Buffer
	for index := 0; index < block.Lines().Len(); index++ {
		segment := block.Lines().At(index)
		content.Write(segment.Value(l.source))
	}
	if strings.TrimSpace(content.String()) == "" {
		l.addError(fmt.Sprintf("Mermaid code block은 비어 있을 수 없습니다: %q", l.currentPolicyHeading))
	}
}

func (l *policyDocumentLinter) lintLinks() {
	_ = ast.Walk(l.document, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		link, ok := node.(*ast.Link)
		if !ok {
			return ast.WalkContinue, nil
		}
		l.lintLink(string(link.Destination))
		return ast.WalkContinue, nil
	})
}

func (l *policyDocumentLinter) lintLink(destination string) {
	parsed, err := url.Parse(strings.TrimSpace(destination))
	if err != nil {
		l.addError(fmt.Sprintf("Markdown link를 해석할 수 없습니다: %q", destination))
		return
	}
	if parsed.IsAbs() || strings.HasPrefix(destination, "//") {
		return
	}
	if strings.HasPrefix(parsed.Path, "/") {
		l.addError(fmt.Sprintf("로컬 Markdown link는 상대경로를 사용해야 합니다: %q", destination))
		return
	}

	decodedPath, err := url.PathUnescape(parsed.Path)
	if err != nil {
		l.addError(fmt.Sprintf("Markdown link 경로를 해석할 수 없습니다: %q", destination))
		return
	}
	if decodedPath == "" {
		return
	}

	target := filepath.Clean(filepath.Join(filepath.Dir(l.absolutePath), filepath.FromSlash(decodedPath)))
	if !isWithin(l.repositoryRoot, target) {
		l.addError(fmt.Sprintf("로컬 Markdown link는 repository 밖을 가리킬 수 없습니다: %q", destination))
		return
	}
	info, err := os.Stat(target)
	if err != nil {
		if os.IsNotExist(err) {
			l.addError(fmt.Sprintf("로컬 Markdown link 대상이 존재하지 않습니다: %q", destination))
			return
		}
		l.addError(fmt.Sprintf("로컬 Markdown link 대상을 확인할 수 없습니다: %q", destination))
		return
	}
	if !info.Mode().IsRegular() {
		l.addError(fmt.Sprintf("로컬 Markdown link 대상은 일반 파일이어야 합니다: %q", destination))
	}
	if filepath.Base(target) == policyFile && parsed.Fragment != "" {
		l.addError(fmt.Sprintf("다른 정책 문서는 POLICY.md 전체를 참조해야 하며 fragment를 사용할 수 없습니다: %q", destination))
	}
}

func isWithin(root string, path string) bool {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func (l *policyDocumentLinter) addError(message string) {
	l.errors = append(l.errors, Error{Path: l.relativePath, Message: message})
}
