FROM golang:1.26-alpine

WORKDIR /app

COPY . .

RUN go build -o /usr/local/bin/workers ./cmd/workers
RUN go build -o /usr/local/bin/ctl ./cmd/cli

CMD ["workers"]
