FROM golang

RUN go get -u github.com/gorilla/mux

COPY api.go /broker/api.go

WORKDIR /broker

ENTRYPOINT go run api.go
