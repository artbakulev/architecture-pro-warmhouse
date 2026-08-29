FROM golang:1.26.6-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /src/server .

FROM alpine:3.21

WORKDIR /app

COPY --from=builder /src/server .

EXPOSE 8081

CMD ["./server"]
