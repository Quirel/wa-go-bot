package main

import (
	"fmt"
	"github.com/Rhymen/go-whatsapp"
	tlogger "github.com/quirel/telegram-logger"
	"log"
	"os"
	"strconv"
	"time"
)

var tgLogger *tlogger.TgLogger

//var once sync.Once

// createTgLoggerInstance creates or returns instance of telegram-logger
func createTgLoggerInstance() *tlogger.TgLogger {
	//once.Do(func() {
	tgToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	tgChatId, _ := strconv.ParseInt(os.Getenv("TELEGRAM_LOG_CHAT_ID"), 10, 64)
	log.Println("Create new tg-logger")
	var err error
	level := "info"
	if isDebug {
		level = "info"
	}
	tgLogger, err = tlogger.NewLogger(level, tgToken, []int64{tgChatId})
	if err != nil {
		log.Fatal("Problem with creating telegram-logger")
	}
	tgLogger.SetName("WA Volley Bot")
	//})
	return tgLogger
}

// graceShutDown terminates script without error
func graceShutDown(msg string, wac *whatsapp.Conn) {
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
	tgLogger.Info(msg)
	fmt.Println(msg)
	fmt.Println("Grace shutdown")
	os.Exit(0)
}

// ping verifies phone connectivity
func ping(wac *whatsapp.Conn) {
	isDebug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	pingTime, _ := strconv.Atoi(os.Getenv("PING_TIME"))

	isPinged := true

	// ping every pingTime
	for range time.Tick(time.Duration(pingTime) * time.Second) {
		pong, err := wac.AdminTest()

		if !pong || err != nil {
			tgLogger.Error(fmt.Sprintf("Error pinging: %v\n", err))
			if isDebug {
				fmt.Printf("⚠️ error pinging: %v\n", err)
			}
			isPinged = false
			//log.Fatalf("⚠️ error pinging in: %v\n", err)
		} else if !isPinged {
			tgLogger.Warn("✅ Ping is OK")
			if isDebug {
				fmt.Printf("⚠️✅ Ping is OK")
			}
			isPinged = true
		}
	}
}
