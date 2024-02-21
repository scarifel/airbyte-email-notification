FROM golang:1.25.1-apphine as base

FROM base as builder
WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o main ./cmd/main.go

FROM base
WORKDIR /app 

COPY --from=builder /app /app
ENTRYPOINT ["./main"]