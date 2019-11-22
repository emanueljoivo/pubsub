FROM golang as build-stage

RUN go get -u github.com/gorilla/mux
RUN go get github.com/fatih/structs

WORKDIR /build

COPY storage.go /build/storage.go

RUN go mod init storage.go

RUN go get

RUN go build -o storage

FROM golang as production-stage

WORKDIR /service

EXPOSE 8003

COPY --from=build-stage /build/storage /service

ENTRYPOINT ["./storage"]
