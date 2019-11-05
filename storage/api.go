package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)



func store(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w,"hey, this is a store")
}


func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/store", store)

	log.Fatal(http.ListenAndServe(":8080", router))
}

