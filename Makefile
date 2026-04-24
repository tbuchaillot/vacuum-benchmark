# vacuum-benchmark

COUNT ?= 5

.PHONY: all fetch bench contracts clean tidy

all: fetch bench

fetch:
	./scripts/fetch.sh

# The orchestrator writes both results.md and contracts.md in one pass, so
# 'bench' and 'contracts' both route through it. Keeping them as separate
# phony targets makes the intent visible in the Makefile's API even though
# they share the same invocation.
bench: fetch
	@mkdir -p results/raw
	cd cmd/report && go run . -count=$(COUNT)

contracts: bench

# `go mod tidy` on each sub-module. Runners use GOWORK=off because they
# are deliberately not in the Go workspace (see go.work comment).
tidy:
	cd internal/scenarios && go mod tidy
	cd runners/upstream   && GOWORK=off go mod tidy
	cd runners/fork       && GOWORK=off go mod tidy
	cd contracts          && go mod tidy
	cd cmd/report         && go mod tidy

clean:
	rm -rf results/
