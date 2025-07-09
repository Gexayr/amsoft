# Use the official Go image as a base
FROM golang:1.22-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod ./
COPY go.sum* ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 go build -o /app/contact-mailer .

# Expose the port the application will listen on
EXPOSE 8080

# Command to run the application
CMD ["/app/contact-mailer"]