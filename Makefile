.PHONY: all  prepare-bin build-nix

all: prepare-bin build-nix 

prepare-bin:
	rm -rf ./bin || true
	mkdir -p ./bin || true

build-nix:
	go build  -o ./bin/zhub4 

install:
	go install
	