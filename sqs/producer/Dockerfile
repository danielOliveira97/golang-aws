FROM golang:1.17.1

WORKDIR /app
COPY . .
RUN go build -o producer .

CMD "./producer"