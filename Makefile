.PHONY: all build run install clean help

all: run

build:
	go build ./cmd/startmc

run:
	go run ./cmd/startmc

install:
	go install ./cmd/startmc

clean:
	go clean
	rm startmc

help:
	@echo "Usage: make [target]"
	@echo "Targets:"
	@echo "  all     : run (default)"
	@echo "  build   : compile the project"
	@echo "  install : build and install the project"
	@echo "  run     : run the project"
	@echo "  clean   : remove build objects and caches" 
