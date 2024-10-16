# Build stage
FROM golang:1.22-alpine AS build

RUN apk add git

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY ./ ./

RUN go build --ldflags "-s -w -X main.version=$(git describe --tags)" -o /redirector src/main.go


# Deploy stage
FROM alpine

HEALTHCHECK CMD /redirector ping -q

WORKDIR /

COPY --from=build /redirector /redirector
RUN chmod +x /redirector

ENTRYPOINT ["/redirector"]
