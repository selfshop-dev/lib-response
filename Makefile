.PHONY: help
help:
	@awk -f help.awk $(MAKEFILE_LIST)
	@echo ""

.DEFAULT_GOAL := help

# === Generation ===

.PHONY: code-gen
code-gen: ## Сгенерировать код
	go generate ./...

# === CI/CD ===

.PHONY: lint
lint: ## Запустить линтер
	golangci-lint cache clean
	golangci-lint run

.PHONY: test
test: ## Запустить тесты с покрытием
	go generate ./...
	go test -race -coverprofile=coverage.out -covermode=atomic \
		$(shell go list ./... | grep -v -E '/(vendor|gen|cmd|testdata|mocks)/')
	sed -i -E '/\/gen\/|_mock\.go:|\.pb(\.gw)?\.go:|\.sql\.go:|\.gen\.go:/d' coverage.out || true
	go tool cover -func=coverage.out | tee coverage.humanize | tail -1

# === Profiling ===

.PHONY: prof
prof: ## Собрать профили (cpu, mem, block, mutex)
	go test -run=. -bench=. \
		-cpuprofile=cpu.out -blockprofile=block.out \
		-memprofile=mem.out -mutexprofile=mutex.out ./...

.PHONY: prof-view
prof-view: ## Открыть профиль в браузере (e.g. make prof-view FILE=mem.out)
	go tool pprof -http=:8082 $(or $(FILE),cpu.out)