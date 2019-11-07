FROM golang

RUN go get -u github.com/gorilla/mux
COPY api.go /replicator/api.go

WORKDIR /replicator

ENTRYPOINT go run api.go
