.PHONY: all

lint: .PHONY
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

bench: .PHONY
	go test -bench=. -count=1 > performance.txt