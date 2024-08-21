FROM golang:1.18

ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o go-ssh-app

EXPOSE 8080

CMD ["./go-ssh-app"]
