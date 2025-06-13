FROM golang:1.23-alpine

WORKDIR /app

# Download modules first
COPY go.mod go.sum ./
RUN go mod download

# Then copy everything else, including static/
COPY . .

# Build the application
RUN go build -o todo .

CMD ["./todo"]
