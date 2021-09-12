#!/bin/make
Binary=torrodle
Main=./cmd/torrodle/main.go

# install for android termux
termux:
	go build -o ${Main} ${Binary}
	chmod +x ${Binary}
	mv ${Binary} ~/../usr/bin/
	
