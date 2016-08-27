package main

import (
	"fmt"
	"log"
)

func main() {

	t, err := NewTranslator("./test-data")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(t.Get("PAGE.LOADING", "de"))
	fmt.Println(t.Set("PAGE.LOADING", "Lappen", "de"))
	fmt.Println(t.Get("PAGE.LOADING", "de"))

	err = t.Sync("de", true)
	if err != nil {
		log.Fatal(err)
	}

	err = t.Save(true)
	if err != nil {
		log.Fatal(err)
	}

}
