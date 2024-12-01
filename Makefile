.PHONY: run-server
run-server:
	@go run cmd/main.go

.PHONY: run-client
run-client:
	@cd web && npm run dev