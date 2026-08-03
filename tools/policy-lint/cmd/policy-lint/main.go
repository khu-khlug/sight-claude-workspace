package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/khu-khlug/sight-workspace/tools/policy-lint/internal/lint"
	"github.com/khu-khlug/sight-workspace/tools/policy-lint/internal/sourcehash"
)

var buildSourceHash = "development"

const buildCommand = "docker buildx build --file tools/policy-lint/Dockerfile --output type=local,dest=tools/policy-lint/bin ."

func main() {
	os.Exit(run())
}

func run() int {
	var rootFlag string
	var printSourceHash bool
	flag.StringVar(&rootFlag, "root", "", "repository root (default: search from current directory)")
	flag.BoolVar(&printSourceHash, "print-source-hash", false, "print policy-lint source hash")
	flag.Parse()

	root, err := resolveRepositoryRoot(rootFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	currentSourceHash, err := sourcehash.Compute(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "policy-lint source hash 계산 실패: %v\n", err)
		return 2
	}
	if printSourceHash {
		fmt.Println(currentSourceHash)
		return 0
	}

	if buildSourceHash == "development" {
		fmt.Fprintln(os.Stderr, "policy-lint binary에 source hash가 없습니다.")
		printBuildInstruction()
		return 2
	}
	if buildSourceHash != currentSourceHash {
		fmt.Fprintln(os.Stderr, "policy-lint binary가 현재 source와 일치하지 않습니다.")
		fmt.Fprintf(os.Stderr, "binary: %s\nsource: %s\n", buildSourceHash, currentSourceHash)
		printBuildInstruction()
		return 2
	}

	errors, err := lint.Run(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "policy lint 실행 실패: %v\n", err)
		return 2
	}
	if len(errors) > 0 {
		for _, lintError := range errors {
			fmt.Fprintln(os.Stderr, lintError.String())
		}
		fmt.Fprintf(os.Stderr, "\nPolicy lint 실패: %d개 오류\n", len(errors))
		return 1
	}

	fmt.Println("Policy lint 성공")
	return 0
}

func printBuildInstruction() {
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "저장소 root에서 다음 명령을 실행한 뒤 lint를 다시 실행하세요:")
	fmt.Fprintln(os.Stderr, buildCommand)
}

func resolveRepositoryRoot(rootFlag string) (string, error) {
	if rootFlag != "" {
		root, err := filepath.Abs(rootFlag)
		if err != nil {
			return "", err
		}
		if isRepositoryRoot(root) {
			return root, nil
		}
		return "", fmt.Errorf("repository root가 아닙니다: %s", root)
	}

	current, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if isRepositoryRoot(current) {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("policy와 tools/policy-lint 디렉터리가 있는 repository root를 찾지 못했습니다")
		}
		current = parent
	}
}

func isRepositoryRoot(path string) bool {
	for _, directory := range []string{"policy", filepath.Join("tools", "policy-lint")} {
		info, err := os.Stat(filepath.Join(path, directory))
		if err != nil || !info.IsDir() {
			return false
		}
	}
	return true
}
