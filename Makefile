
.DEFAULT_GOAL := build

release/powerline-go-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -o release/powerline-go-darwin-amd64

.PHONY: build
build: release/powerline-go-darwin-amd64

.PHONY: clean
clean:
	rm -rvf release
