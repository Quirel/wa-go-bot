package main

import (
	"fmt"
	"github.com/Rhymen/go-whatsapp"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// init - loads .env
func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	clientName, _ := os.LookupEnv("CLIENT_NAME")

	//create new WhatsApp connection
	wac, err := whatsapp.NewConnWithOptions(&whatsapp.Options{
		Timeout:         10 * time.Second,
		ShortClientName: clientName,
		LongClientName:  clientName,
	})
	if err != nil {
		log.Fatalf("error creating connection: %v\n", err)
	}

	//get message for current day from schedule
	msgText, err := getMessageFromSchedule()
	if err != nil {
		log.Fatalf("schedule error: %v", err)
	}
	if msgText == "" {
		session, err := wac.Disconnect()
		if err != nil {
			log.Fatalf("error disconnecting: %v\n", err)
		}
		if err := writeSession(session); err != nil {
			log.Fatalf("error saving session: %v", err)
		}
		graceShutDown("Day is empty. Terminating")
	}

	//add custom handlers
	wac.AddHandler(&waHandler{wac, uint64(time.Now().Unix()), msgText})

	//login or restore
	err = login(wac)
	if err != nil {
		log.Fatalf("error logging in: %v\n", err)
	}

	//verifies phone connectivity
	pong, err := wac.AdminTest()

	if !pong || err != nil {
		log.Fatalf("error pinging in: %v\n", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	//Disconnect safe
	fmt.Println("Shutting down now.")
	session, err := wac.Disconnect()
	if err != nil {
		log.Fatalf("error disconnecting: %v\n", err)
	}
	if err := writeSession(session); err != nil {
		log.Fatalf("error saving session: %v", err)
	}
}

// graceShutDown - terminates script without error
func graceShutDown(msg string) {
	fmt.Println(msg)
	fmt.Println("Grace shutdown")
	os.Exit(0)
}
