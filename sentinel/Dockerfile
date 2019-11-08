FROM golang

RUN go get -u github.com/gorilla/mux

COPY api.go /sentinel/api.go

WORKDIR /sentinel

ENTRYPOINT go run api.go
