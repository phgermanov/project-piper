package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, Go!")
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", homeHandler).Methods("GET")

	fmt.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}

