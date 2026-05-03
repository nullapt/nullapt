FROM golang:1.25.9-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /out/registry-api ./cmd/registry-api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /out/registry-api /usr/local/bin/registry-api

EXPOSE 8080
CMD ["/usr/local/bin/registry-api"]
