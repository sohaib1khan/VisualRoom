FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o vroom .

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /src/vroom /usr/local/bin/vroom
COPY config/default.yaml /config/default.yaml
ENTRYPOINT ["vroom"]
