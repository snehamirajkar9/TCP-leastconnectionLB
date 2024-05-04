FROM golang:1.22 AS build-stage

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download 

# Copy source code
COPY balancer balancer/
COPY server server/
COPY main.go main.go

# Build the app
RUN CGO_ENABLED=0 GOOS=linux go build -a -o loadbalancer main.go

# Production image
FROM gcr.io/distroless/base-debian12 AS run-stage

WORKDIR /

# Copy binary from build stage
COPY --from=build-stage /app/loadbalancer loadbalancer
COPY certs certs/

EXPOSE 9443
ENTRYPOINT [ "./loadbalancer" ]