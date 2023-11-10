.PHONY: all  prepare-bin zhub4

all: prepare-bin zhub4 

prepare-bin:
	rm -rf ./bin || true
	mkdir -p ./bin || true

zhub4:
	go build  -o ./bin/zhub4 

install:
	cp ./bin/zhub4 /usr/local/bin 
	cp config.txt /usr/local/etc/zhub4/
	cp httpServer/gsb_style.css /usr/local/etc/zhub4/web/
