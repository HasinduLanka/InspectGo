FROM golang:1.18 AS builder

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Move to working directory /src
WORKDIR /src

COPY go.mod ./
RUN go mod download && go mod verify

# Copy the code into the container
COPY . .

RUN go mod tidy

# Run test
RUN go test -v ./...

# Build the application
RUN go build -o main .

# Move to /dist directory as the place for resulting binary folder
WORKDIR /dist

# Copy binary and frontend from build to main folder
RUN cp /src/main .
RUN cp -r /src/frontend ./frontend

EXPOSE 20000

ENTRYPOINT ["/dist/main"]
