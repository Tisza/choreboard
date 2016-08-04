package main

import (
	"fmt"
	"net/http"
)

/**
	TODO: create stub handler for userStatus
	TODO: create stub handler for signChore
	TODO: create stub handler for choreBoard
	TODO: create stub handler for loginUser
	TODO: create stub handler for reportChore
 */

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
	fmt.Println("Hi there, I love %s!", r.URL.Path[1:])

}

func main() {
	const host string = "localhost"
	const port string = "8080"

	http.HandleFunc("/", handler)
	http.ListenAndServe(host + ":" + port, nil)
}