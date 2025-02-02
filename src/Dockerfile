FROM golang:1.23 as build

WORKDIR /app

# ENV CGO_ENABLED 1
# ENV GOPATH /go
# ENV GOCACHE /go-build

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod .
COPY go.sum .

# ADD go.mod /app/
# ADD go.sum /app/

RUN go mod download
# RUN go install github.com/vakenbolt/go-test-report@latest

COPY . .
# ADD cmd /app/cmd/
# ADD configs /app/configs/
# ADD e2e /app/e2e/
# ADD migrations /app/migrations
# ADD internal /app/internal/
# ADD Makefile /app/Makefile


# CMD ["go", "run", "./cmd/main.go"]

############################################

# RUN make build
RUN go build -o main cmd/main.go

FROM golang:1.23
# FROM apline:latest

# COPY --from=build /app/main /usr/bin

# RUN apk add --no-cache postgresql-client

# Copy the Pre-built binary file from the previous stage
COPY --from=build /app/main /main

EXPOSE 9000

# RUN chmod +x ./main

CMD ["/main"]
