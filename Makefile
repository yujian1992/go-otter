all: build

build:
        go build -ldflags "-s -w" -gcflags "-N -l" -o ./GoProducer ./producer/producer_main.go
        go build -ldflags "-s -w" -gcflags "-N -l" -o ./GoConsumer ./consumer/consumer_main.go
clean:
	@rm -rf bin
test:
	go test ./go/... -race
