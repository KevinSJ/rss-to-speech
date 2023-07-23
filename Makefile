export GOPROXY 		:= https://proxy.golang.org,https://gocenter.io,direct
export PATH 		:= ./bin:$(PATH)
export GO111MODULE 	:= on


setup:
	#curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh
	curl -sfL https://goreleaser.com/static/run | bash
	go mod tidy

.PHONY: setup

test:
	go test $(TEST_OPTIONS) -v -failfast -race -coverpkg=./... -covermode=atomic -coverprofile=coverage.out $(SOURCE_FILES) -run $(TEST_PATTERN) -timeout=2m

cover: test
	go tool cover -html=coverage.out

fmt:
	find . -name '*.go' -not -wholename './vendor/*' | while read -r file; do gofmt -w -s "$$file"; goimports -w "$$file"; done

lint:
	./bin/golangci-lint run --fix ./...

ci: test clean build

clean:
	go clean && rm -rf ./dist

build: clean
	goreleaser release --snapshot

build-aarm64-linux:
	env GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o rss-to-speech-arm ./main.go

.DEFAULT_GOAL := ci

#all: build-armv7 build-x64
#build-armv7:
	#env GOOS=linux GOARCH=arm GOARM=7 go build
#build-x64:
	#go build
#clean:
	#rm -rf ./rss-to-podcast
	#find . -type d  -name "*2023*" -exec rm -rf {} \;
