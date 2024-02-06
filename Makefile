
# This is the default install path for Mac/Linux. Change it as needed for your system
FACTORIO_DIR := ${HOME}/factorio
FACTORIO_BIN := ${FACTORIO_DIR}/bin/x64/factorio


gen: fmr
	./fmr mods/MinPctTAS_0.0.1/tasks.lua

fmr:
	go build -o fmr .

start_factorio:
	${FACTORIO_BIN} --mod-directory ./mods


.PHONY: start_factorio gen
