FROM golang:1.26-alpine

WORKDIR /app

COPY . .

RUN go build -o workers ./cmd/workers

CMD ["./workers"]
