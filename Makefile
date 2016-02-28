SRC := $(shell find . -type f -name '*.go')
ASSETS := $(shell find ./assets/assets -type f)

.PHONY: run lint install

tiif: $(SRC) assets/assets.go
	go build

install:
	go install

lint:
	-golint ./... | grep -v "exported .* should have comment"

assets/assets.go: $(ASSETS)
	go-bindata -pkg assets -o assets/assets.go -prefix assets/assets assets/assets/...

