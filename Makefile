
build: fmin

FMIN_FILES := main.go $(wildcard tasscript/*.go)

fmin: $(FMIN_FILES)
	go build -o fmin .

tasscript/tas.go: tasscript/tas.y
	go generate ./...

clean:
	rm -v tasscript/y.output tasscript/tas.go fmin

.PHONY: build clean
