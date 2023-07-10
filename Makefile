
# This is the default install path for Mac/Linux. Change it as needed for your system
FACTORIO_DIR := ${HOME}/factorio
FACTORIO_BIN := ${FACTORIO_DIR}/bin/x64/factorio

start_factorio:
	${FACTORIO_BIN} --mod-directory ./mods


.PHONY: start_factorio
