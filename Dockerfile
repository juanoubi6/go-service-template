# Use the official Golang image as a parent image
FROM golang:1.19 AS build

# Set the working directory
WORKDIR /app

# Copy the current directory contents into the container at /app
COPY . .

# Build the Go app
RUN go build -o main .

# Use a minimal base image
FROM alpine:latest

# Copy the compiled binary from the previous stage
COPY --from=build /app/main /app/main

# Expose port 8080
EXPOSE 8080

# Set the command to run the binary
CMD ["/app/main"]
