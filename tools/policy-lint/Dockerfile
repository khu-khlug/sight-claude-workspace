FROM golang:1.23.4-alpine3.20 AS builder

WORKDIR /workspace

COPY policy policy
COPY tools/policy-lint tools/policy-lint

WORKDIR /workspace/tools/policy-lint

RUN go test ./...

RUN set -eu; \
    source_hash="$(go run ./cmd/policy-lint --root /workspace --print-source-hash)"; \
    mkdir -p /out; \
    CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath \
      -ldflags="-s -w -X main.buildSourceHash=${source_hash}" \
      -o /out/policy-lint-darwin-arm64 ./cmd/policy-lint; \
    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath \
      -ldflags="-s -w -X main.buildSourceHash=${source_hash}" \
      -o /out/policy-lint-darwin-amd64 ./cmd/policy-lint; \
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath \
      -ldflags="-s -w -X main.buildSourceHash=${source_hash}" \
      -o /out/policy-lint-windows-amd64.exe ./cmd/policy-lint; \
    CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -trimpath \
      -ldflags="-s -w -X main.buildSourceHash=${source_hash}" \
      -o /out/policy-lint-windows-arm64.exe ./cmd/policy-lint; \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath \
      -ldflags="-s -w -X main.buildSourceHash=${source_hash}" \
      -o /out/policy-lint-linux-amd64 ./cmd/policy-lint; \
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath \
      -ldflags="-s -w -X main.buildSourceHash=${source_hash}" \
      -o /out/policy-lint-linux-arm64 ./cmd/policy-lint

FROM scratch AS binaries

COPY --from=builder /out/ /
