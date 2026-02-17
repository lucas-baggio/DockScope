.PHONY: build run cli tidy

BINARY := dockscope
CMD   := ./cmd/dockscope

build:
	go build -o $(BINARY) $(CMD)

run: build
	./$(BINARY)

cli: build
	./$(BINARY) --cli --all

tidy:
	go mod tidy
