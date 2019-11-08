FROM golang

RUN go get -u github.com/gorilla/mux

COPY api.go /storage/api.go

WORKDIR /storage

ENTRYPOINT go run api.go
