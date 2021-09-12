#!/bin/make
Binary=torrodle

# install for android termux
termux:
	go build -o torrodle.go ${Binary}
	chmod +x ${Binary}
	mv ${Binary} ~/../usr/bin/
	
