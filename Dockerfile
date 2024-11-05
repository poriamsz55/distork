# Use an official Golang image as the base image
FROM golang:1.23.2

# Set the working directory inside the container
WORKDIR /app

# Copy the Go mod and sum files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -tags dev -o distork .

# Expose the port the app will run on
EXPOSE 8080

# FROM scratch
# COPY --from=0 /usr/local/bin/drive /usr/local/bin/drive

# Run the executable
CMD ["/app/distork"]