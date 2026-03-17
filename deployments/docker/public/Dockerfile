FROM golang:1.26-alpine

WORKDIR /app

COPY . .

RUN go build -o /usr/local/bin/public-api ./cmd/public

EXPOSE 8080

CMD ["public-api"]
