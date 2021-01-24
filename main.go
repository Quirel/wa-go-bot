package main

import (
	"fmt"
	"github.com/Rhymen/go-whatsapp"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

// init - loads .env
func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln("No .env file found")
	}
}

func main() {
	isDebug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	if isDebug {
		fmt.Println("⚠️ DEBUG MODE")
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

	//create Telegram connection for logs
	tgToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	tgBot, err := tgbotapi.NewBotAPI(tgToken)
	if err != nil {
		log.Fatalf("error creating telegram connection: %v\n", err)
	}

	//get message for current day from schedule
	msgText, err := getMessageFromSchedule()
	if err != nil {
		tgLog(fmt.Sprintf("❌ schedule error: %v", err), tgBot)
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
		graceShutDown("⚠️ Day is empty. Terminating", tgBot, wac)
	}

	//add custom handlers
	wac.AddHandler(&waHandler{wac, uint64(time.Now().Unix()), msgText, tgBot})

	//login or restore session
	err = login(wac)
	if err != nil {
		tgLog(fmt.Sprintf("❌ error logging in: %v\n", err), tgBot)
		log.Fatalf("error logging in: %v\n", err)
	}
	tgLog(fmt.Sprintf("Login successful\nMessage to send:\n%v", msgText), tgBot)

	// check phone connectivity
	go ping(wac, tgBot)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	//Disconnect safe
	fmt.Println("\nShutting down now.")
	tgLog("⚠️ Shutting down now.", tgBot)
	session, err := wac.Disconnect()
	if err != nil {
		tgLog(fmt.Sprintf("❌ error disconnecting: %v", err), tgBot)
		log.Fatalf("error disconnecting: %v\n", err)
	}
	if err := writeSession(session); err != nil {
		tgLog(fmt.Sprintf("❌ error saving session: %v", err), tgBot)
		log.Fatalf("error saving session: %v", err)
	}
}
