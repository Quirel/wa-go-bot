package main

import (
	"fmt"
	"github.com/Rhymen/go-whatsapp"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"os"
	"strconv"
	"time"
)

// graceShutDown - terminates script without error
func graceShutDown(msg string, tgBot *tgbotapi.BotAPI, wac *whatsapp.Conn) {
	if wac.GetConnected() {
		session, err := wac.Disconnect()
		if err != nil {
			log.Fatalf("error disconnecting: %v\n", err)
		}
		fmt.Println("wac.Disconnect")
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
	tgChatId := os.Getenv("TELEGRAM_LOG_CHAT_ID")
	tgChatIdInt, _ := strconv.ParseInt(tgChatId, 10, 64)
	tgMsg := tgbotapi.NewMessage(tgChatIdInt, "Wa-Go-Bot: "+msg)
	_, err := tgBot.Send(tgMsg)
	if err != nil {
		log.Fatalf("Send telegram error: %v\n", err)
	}
}

// ping - verifies phone connectivity
func ping(wac *whatsapp.Conn, tgBot *tgbotapi.BotAPI) {
	isDebug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	pingTime, _ := strconv.Atoi(os.Getenv("PING_TIME"))

	isPinged := true

	// ping every pingTime
	for range time.Tick(time.Duration(pingTime) * time.Second) {
		pong, err := wac.AdminTest()

		if !pong || err != nil {
			tgLog(fmt.Sprintf("⚠️ error pinging: %v\n", err), tgBot)
			if isDebug {
				fmt.Printf("⚠️ error pinging: %v\n", err)
			}
			isPinged = false
			//log.Fatalf("⚠️ error pinging in: %v\n", err)
		} else if !isPinged {
			tgLog(fmt.Sprintf("⚠️✅ Ping is OK"), tgBot)
			if isDebug {
				fmt.Printf("⚠️✅ Ping is OK")
			}
			isPinged = true
		}
	}
}
