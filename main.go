package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	res, err := http.Post("https://minhaapi.com", "application/json", nil)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
	return

	// res, err = http.Get("https://google.com")
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	data, err = io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}
