
# This is the default install path for Mac/Linux. Change it as needed for your system
FACTORIO_DIR := ${HOME}/factorio
FACTORIO_BIN := ${FACTORIO_DIR}/bin/x64/factorio

OUT_FILE := mods/MinPctTAS_0.0.1/tasks.lua

${OUT_FILE}: fmr
	./fmr ${OUT_FILE}

targets := $(wildcard *.go) $(wildcard **/*.go)
fmr: $(targets)
	go vet ./...
	go build -o fmr .

start_factorio:
	${FACTORIO_BIN} --mod-directory ./mods

test:
	go test -v -covermode=count ./...


.PHONY: start_factorio test
