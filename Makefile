
build: compile

COMPILE_FILES := main.go $(wildcard tasscript/*.go)

compile: $(COMPILE_FILES)
	go build -o compile ./cmd/compile

tasscript/tas.go: tasscript/tas.y
	go generate ./...

clean:
	rm -v tasscript/y.output tasscript/tas.go compile

.PHONY: build clean
