# Build stage
FROM golang:1.22-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY ./ ./

RUN go build -o /redirector src/main.go


# Deploy stage
FROM alpine

HEALTHCHECK CMD /redirector ping -q

WORKDIR /

COPY --from=build /redirector /redirector

ENTRYPOINT ["/redirector"]
