# Start with a base image that has Go installed.
FROM golang:latest

# Set the working directory inside the container.
WORKDIR /app

# Copy the Go module files and download dependencies.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code.
COPY . .

# Build the Go application.
RUN go build -o /watchalong-server ./cmd/server.go

# Expose the port the application will run on.
EXPOSE 8080

# Set the entrypoint for the container.
ENTRYPOINT [ "/watchalong-server" ]
