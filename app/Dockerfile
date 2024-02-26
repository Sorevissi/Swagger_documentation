FROM golang:alpine as builder

WORKDIR /app

COPY . .

RUN go mod init test && go mod tidy\
    && go build -o main

FROM alpine:latest

COPY --from=builder /app/main /main

EXPOSE 8080

CMD ["/main"]