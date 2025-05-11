BINARY_NAME=email-fetcher
PACKAGE=main.go

build:
	go build -o $(BINARY_NAME) $(PACKAGE)

run:
	go run $(PACKAGE) -username="your-email@outlook.com" -password="your-app-password"

clean:
	rm -f $(BINARY_NAME) emails.csv

install:
	go mod tidy

help:
	@echo "Usage:"
	@echo "  make build     # Compile the binary"
	@echo "  make run       # Run with current .go source"
	@echo "  make clean     # Remove binary and output"
	@echo "  make install   # Install/update dependencies"
	