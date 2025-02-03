FROM golang:1.23

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .
