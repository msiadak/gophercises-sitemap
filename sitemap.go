package main

import (
	"fmt"
	"net/http"

	link "github.com/msiadak/gophercises-link"
)

func main() {
	w, err := http.Get("https://www.google.com")
	if err != nil {
		panic(err)
	}
	links, err := link.ExtractLinks(w.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v", links)
}
