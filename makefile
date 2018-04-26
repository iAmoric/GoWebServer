GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BINARY_NAME=server

all: build

build:
		$(GOBUILD) -o ./bin/$(BINARY_NAME) ./src

clean:
		$(GOCLEAN)
		rm -f ./bin/$(BINARY_NAME)
run:
		$(GOBUILD) -o ./bin/$(BINARY_NAME) -v ./src
		./bin/$(BINARY_NAME)
