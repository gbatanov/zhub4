.PHONY: all  prepare-bin 

all: prepare-bin zhub4 

prepare-bin:
	rm -rf ./bin || true
	mkdir -p ./bin || true

zhub4:
	go build  -o ./bin/zhub4 

install:
	go install
	cp config /usr/local/etc/zhub4/
	cp http_server/gsb_style.css /usr/local/etc/zhub4/web/
