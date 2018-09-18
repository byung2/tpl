## simple makefile
.PHONY: all clean build install

all: build 

build:
	go build -o tpl cmd/tpl/main.go

install: build 
	sudo mv tpl /usr/local/bin/tpl

clean:
	go clean 
	rm tpl

## EOF
