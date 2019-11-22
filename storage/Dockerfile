FROM golang

RUN go get -u github.com/gorilla/mux
RUN go get github.com/fatih/structs

COPY api.go /storage/api.go

WORKDIR /storage

ENTRYPOINT go run api.go
