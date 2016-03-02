SRC := $(shell find . -type f -name '*.go')
ASSETS := $(shell find ./assets/assets -type f)
CROSSARCH := amd64 386
CROSSOS := darwin linux openbsd netbsd freebsd
CROSS := $(foreach os,$(CROSSOS),$(foreach arch,$(CROSSARCH),bin/$(os).$(arch)))

.PHONY: run lint install reset publish

bin/tiif: $(SRC) assets/assets.go
	@-mkdir bin 2>/dev/null || true
	go build -o bin/tiif

install:
	go install

lint:
	-golint ./... | grep -v "exported .* should have comment"

cross: assets/assets.go
	$(MAKE) $(CROSS)

$(CROSS): $(SRC) assets/assets.go
	@-mkdir bin 2>/dev/null || true
	gox -osarch=$(shell basename $@ | sed 's/\./\//') -output="bin/{{.OS}}.{{.Arch}}"

assets/assets.go: $(ASSETS)
	go-bindata -pkg assets -o assets/assets.go -prefix assets/assets assets/assets/...

publish: reset
	$(MAKE) cross
	s3cmd put -P ./bin/* s3://ogresarelikeonions.fyi/bin/tiif/

reset:
	-rm assets/assets.go
	-rm -rf bin

