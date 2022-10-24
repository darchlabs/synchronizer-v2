FROM golang:alpine as builder

WORKDIR /app

COPY . .

# ENV INTERVAL_SECONDS=10
# ENV DATABASE_URL=./storage.db

RUN CGO_ENABLED=0 go build -o bin/synchronizer cmd/synchronizer/main.go

ENTRYPOINT [ "./bin/synchronizer" ]