# Start from the golang image
FROM golang:alpine as compiler

# Set the working directory for golang image
WORKDIR /Task1

# Copy all project files inside the container to /app
COPY . .

# Build our application
RUN go build -o golang-project

#specify the base image for the Docker container as Alpine Linux
FROM alpine

# Set the working directory for our application
WORKDIR /small

# Copy all project files from golang working directory to Alpine Linux Docker container 
COPY --from=compiler ./Task1/golang-project ./binary

# Specify the initial command that should be executed
ENTRYPOINT [ "./binary" ]

# To build the Image of our golang project you need to input the command:
# docker build --go-web 
