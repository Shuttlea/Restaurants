package main

import (
	"fmt"
	"net/http"
)

func main() {

	http.HandleFunc("/", Handler)
	http.HandleFunc("/exit", ExitHandler)
	http.HandleFunc("/api/places", JsonHandler)
	http.HandleFunc("/api/recommend", RecomHandler)
	http.HandleFunc("/api/get_token", TokenHandler)

	fmt.Println("Server is listening...")
	http.ListenAndServe(":8888", nil)
	fmt.Println("Server down")
}
