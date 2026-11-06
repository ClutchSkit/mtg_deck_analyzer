package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		fmt.Printf("Failed to create http server: %v", err)
		os.Exit(1)
	}
}
