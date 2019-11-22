FROM golang

RUN go get -u github.com/gorilla/mux

COPY api.go /sentinel/sentinel.go

WORKDIR /sentinel

ENTRYPOINT go run sentinel.go
