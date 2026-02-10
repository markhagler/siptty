.PHONY: build test test-integration lint clean release

build:
	go build -o siptty ./cmd/siptty

test:
	go test ./...

test-integration:
	docker compose -f docker-compose.test.yml up -d --wait
	go test -tags integration ./tests/

lint:
	golangci-lint run

clean:
	rm -f siptty

release:
	GOOS=linux   GOARCH=amd64 go build -o siptty-linux-amd64       ./cmd/siptty
	GOOS=linux   GOARCH=arm64 go build -o siptty-linux-arm64       ./cmd/siptty
	GOOS=darwin  GOARCH=arm64 go build -o siptty-darwin-arm64      ./cmd/siptty
	GOOS=darwin  GOARCH=amd64 go build -o siptty-darwin-amd64      ./cmd/siptty
	GOOS=windows GOARCH=amd64 go build -o siptty-windows-amd64.exe ./cmd/siptty
