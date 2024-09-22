# Start with the official Golang image
FROM golang:1.22.4-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download necessary Go modules
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the application
RUN go build -o main .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the application
CMD ["./main"]
