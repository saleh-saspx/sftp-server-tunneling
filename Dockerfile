FROM golang:1.18

ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o go-ssh-app

RUN sed -i 's|http://deb.debian.org|http://mirrors.ustc.edu.cn|g' /etc/apt/sources.list && \
    apt-get update && \
    apt-get install -y ca-certificates && \
    apt-get clean

EXPOSE 8080

CMD ["./go-ssh-app"]
