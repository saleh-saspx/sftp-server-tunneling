FROM golang:1.20

    WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go mod tidy
RUN go build -o main

RUN sed -i 's|http://deb.debian.org|http://mirrors.ustc.edu.cn|g' /etc/apt/sources.list && \
    apt-get update && \
    apt-get install -y ca-certificates && \
    apt-get clean

EXPOSE 8080

CMD ["go", "run", "main.go"]