FROM golang:1.21-alpine

WORKDIR /app

# Copy go files and modules
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o email-fetcher main.go

CMD ["./email-fetcher"]
