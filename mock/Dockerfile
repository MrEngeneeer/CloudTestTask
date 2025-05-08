# mock/Dockerfile

FROM golang:1.23-alpine

WORKDIR /app

COPY mock/main.go .

RUN go mod init mock && go mod tidy

RUN go build -o mock .

EXPOSE 9001

CMD ["./mock"]
