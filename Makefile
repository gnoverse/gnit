.PHONY: install build

all: install

build:
	cd client && CGO_ENABLED=0 go build -o ../gnit ./cmd/gnit

install:
	CGO_ENABLED=0 go install -C ./client/cmd/gnit
