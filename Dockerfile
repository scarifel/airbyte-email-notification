FROM golang:1.21.5-alpine as base

FROM base as builder
WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o main ./cmd/main.go

FROM base
WORKDIR /app 

COPY --from=builder /app/main /app/main
ENTRYPOINT ["./main"]