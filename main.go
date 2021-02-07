package main

import (
	"fmt"
	"github.com/Rhymen/go-whatsapp"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var isDebug bool

// init - loads .env
func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln("No .env file found")
	}
}

func main() {
	// initiate telegram-logger instance
	createTgLoggerInstance()
	isDebug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	if isDebug {
		fmt.Println("⚠️ DEBUG MODE")
		tgLogger.Debug("DEBUG MODE")
	} else {
		fmt.Println("⚠️ PRODUCTION MODE")
	}

	clientNameShort := os.Getenv("CLIENT_NAME_SHORT")
	clientNameLong := os.Getenv("CLIENT_NAME_LONG")
	clientVersion := os.Getenv("CLIENT_NAME_VERSION")
	clientTimeout, _ := strconv.Atoi(os.Getenv("WA_CLIENT_TIMEOUT"))

	//create new WhatsApp connection
	wac, err := whatsapp.NewConnWithOptions(&whatsapp.Options{
		Timeout:         time.Duration(clientTimeout) * time.Second,
		ShortClientName: clientNameShort,
		LongClientName:  clientNameLong,
		ClientVersion:   clientVersion,
	})
	if err != nil {
		log.Fatalf("error creating connection: %v\n", err)
	}

	//get message for current day from schedule
	msgText, err := getMessageFromSchedule()
	if err != nil {
		tgLogger.Error(fmt.Sprintf("Schedule error: %v", err))
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
		graceShutDown("⚠️ Day is empty. Terminating", wac)
	}

	//add custom handlers
	wac.AddHandler(&waHandler{wac, uint64(time.Now().Unix()), msgText})

	//login or restore session
	err = login(wac)
	if err != nil {
		tgLogger.Error(fmt.Sprintf("Error logging in: %v\n", err))
		log.Fatalf("error logging in: %v\n", err)
	}
	tgLogger.Info(fmt.Sprintf("Login successful\nMessage to send:\n%v", msgText))

	// check phone connectivity
	go ping(wac)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	//Disconnect safe
	fmt.Println("\nShutting down now.")
	tgLogger.Warn("Shutting down now.")
	session, err := wac.Disconnect()
	if err != nil {
		tgLogger.Error(fmt.Sprintf("Error disconnecting: %v", err))
		log.Fatalf("error disconnecting: %v\n", err)
	}
	if err = writeSession(session); err != nil {
		tgLogger.Error(fmt.Sprintf("Error saving session: %v", err))
		log.Fatalf("error saving session: %v", err)
	}
}
