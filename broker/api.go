package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func pub(w http.ResponseWriter, r *http.Request) {
	fmt.Println(w, "hey, this is pub")
	
}

func sub(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hey, this is sub")
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/pub", pub)
	router.HandleFunc("/sub", sub)

	log.Fatal(http.ListenAndServe(":8080", router))
}