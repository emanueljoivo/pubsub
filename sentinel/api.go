package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)



func store(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w,"hey, this storing in a storage")
}

func get(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w,"hey, getting this from a storage")
}

func checkForBalance() {
	fmt.Println("just checking if everybody is ok")
	for ;; {
		fmt.Println("checking")
	}
}

func main() {
	router := mux.NewRouter().StrictSlash(true)	router.HandleFunc("/rebalance", store)

	go checkForBalance()
	router.HandleFunc("/store", store)
	router.HandleFunc("/get", store)

	log.Fatal(http.ListenAndServe(":8080", router))
}

