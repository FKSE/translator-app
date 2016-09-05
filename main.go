package main

import (
	"log"
)

func main() {

	t, err := NewTranslator("./test-data")
	if err != nil {
		log.Fatal(err)
	}

	server := NewServer(t)
	server.Run()

}
