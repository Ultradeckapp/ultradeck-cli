From golang:1.8

WORKDIR /go/src/ultradeck-cli
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

EXPOSE 8080

CMD ["ultradeck-server", "-redisAddr=redis:6379","-ultradeckBackendAddr=backend:3001"]

