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
		log.Print("No .env file found")
	}
}

func main() {
	clientNameShort, _ := os.LookupEnv("CLIENT_NAME_SHORT")
	clientNameLong, _ := os.LookupEnv("CLIENT_NAME_LONG")
	clientVersion, _ := os.LookupEnv("CLIENT_NAME_VERSION")

	//create new WhatsApp connection
	wac, err := whatsapp.NewConnWithOptions(&whatsapp.Options{
		Timeout:         5 * time.Second,
		ShortClientName: clientNameShort,
		LongClientName:  clientNameLong,
		ClientVersion:   clientVersion,
	})
	if err != nil {
		log.Fatalf("error creating connection: %v\n", err)
	}

	//create Telegram connection for logs
	tgToken, _ := os.LookupEnv("TELEGRAM_BOT_TOKEN")
	tgBot, err := tgbotapi.NewBotAPI(tgToken)
	if err != nil {
		log.Fatalf("error creating telegram connection: %v\n", err)
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
		graceShutDown("Day is empty. Terminating", tgBot, wac)
	}

	//add custom handlers
	wac.AddHandler(&waHandler{wac, uint64(time.Now().Unix()), msgText, tgBot})

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
func graceShutDown(msg string, tgBot *tgbotapi.BotAPI, wac *whatsapp.Conn) {
	if wac.GetConnected() {
		session, err := wac.Disconnect()
		if err != nil {
			log.Fatalf("error disconnecting: %v\n", err)
		}
		log.Println("wac.Disconnect")
		if err := writeSession(session); err != nil {
			log.Fatalf("error saving session: %v", err)
		}
	}
	tgLog(msg, tgBot)
	fmt.Println(msg)
	fmt.Println("Grace shutdown")
	os.Exit(0)
}

// tgLog - logs message to telegram chat
func tgLog(msg string, tgBot *tgbotapi.BotAPI) {
	tgChatId, _ := os.LookupEnv("TELEGRAM_LOG_CHAT_ID")
	tgChatIdInt, _ := strconv.ParseInt(tgChatId, 10, 64)
	tgMsg := tgbotapi.NewMessage(tgChatIdInt, msg)
	_, err := tgBot.Send(tgMsg)
	if err != nil {
		log.Fatalf("Send telegram error: %v\n", err)
	}
}
