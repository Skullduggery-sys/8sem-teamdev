FROM dashori/golang-dind:1.21.3 AS build

WORKDIR /app

COPY . .

RUN go mod download

ENTRYPOINT cd /app/cmd
