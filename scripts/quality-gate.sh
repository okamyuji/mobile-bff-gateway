#!/usr/bin/env sh
set -eu

export GOCACHE="${GOCACHE:-/tmp/mobile-bff-gateway-go-build}"
export STATICCHECK_CACHE="${STATICCHECK_CACHE:-/tmp/mobile-bff-gateway-staticcheck}"
export GOLANGCI_LINT_CACHE="${GOLANGCI_LINT_CACHE:-/tmp/mobile-bff-gateway-golangci-lint}"

unformatted="$(gofmt -l .)"
if [ -n "$unformatted" ]; then
  printf '%s\n' "$unformatted"
  exit 1
fi

go vet ./...
go test -count=1 -shuffle=on -cover ./...
staticcheck ./...
golangci-lint run ./...
go build ./cmd/gateway ./cmd/mockservice
