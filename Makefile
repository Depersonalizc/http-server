all:
	go build -o client ./cmd/client
	go build -o server ./cmd/server
clean:
	rm -fv client server