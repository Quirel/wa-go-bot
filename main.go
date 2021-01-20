package main

import (
	"github.com/Rhymen/go-whatsapp"
	"log"
	"time"
)

func main() {
	//create new WhatsApp connection
	wac, err := whatsapp.NewConn(5 * time.Second)
	if err != nil {
		log.Fatalf("error creating connection: %v\n", err)
	}

	err = login(wac)
	if err != nil {
		log.Fatalf("error logging in: %v\n", err)
	}

	//wac.AddHandler(&waHandler{wac})
}
