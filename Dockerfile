FROM golang:alpine as builder

WORKDIR /usr/src/app

COPY . .

RUN CGO_ENABLED=0 go build -o sync cmd/synchronizer/main.go

FROM golang:alpine as runner

WORKDIR /home/sync

COPY --from=builder /usr/src/app/sync /home/sync

CMD [ "./sync" ]
