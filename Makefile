.PHONY: all build run install clean help

ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

all: run

build:
	go build ./cmd/startmc

run:
	go run ./cmd/startmc

install: build
	sudo mv $(ROOT_DIR)/startmc /usr/local/bin/
	sudo cp $(ROOT_DIR)/startmc.service /etc/systemd/system

refresh: install
	sudo systemctl daemon-reload
	sudo systemctl start startmc
	sudo systemctl enable startmc

clean:
	go clean
	rm -f startmc

uninstall: clean
	sudo systemctl stop startmc
	sudo systemctl disable startmc
	sudo rm /etc/systemd/system/startmc.service
	sudo systemctl daemon-reload
	sudo rm /usr/local/bin/startmc

help:
	@echo "Usage: make [target]"
	@echo "Targets:"
	@echo "  all       : run (default)"
	@echo "  build     : compile the project"
	@echo "  run       : run the project"
	@echo "  install   : build and install the project"
	@echo "  refresh   : install project and reload systemd"
	@echo "  clean     : remove build objects and caches" 
	@echo "  uninstall : cleans and removes all installed files and units"
