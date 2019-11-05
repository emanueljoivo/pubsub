package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)



func podsUp(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w,"hey, let's create some pods")
}

func podsDown(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w,"hey, let's kill some pods")
}

func rebalance(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w,"hey, this let's rebalance")	
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/create", podsUp)
	router.HandleFunc("/destroy", podsDown)
	router.HandleFunc("/rebalance", rebalance)


	log.Fatal(http.ListenAndServe(":8080", router))
}

