# Start from the official Go image
FROM golang:1.22

# Install PostgreSQL client
RUN apt-get update && apt-get install -y postgresql-client

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules and sum files
COPY go.mod go.sum ./

# Download Go module dependencies
RUN go mod download

# Install SQLC
# RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Install Goose
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Copy the rest of the application's source code
COPY . .

# Build the application
RUN go build -o myapp

# Command to run the executable
# CMD ["./myapp"]