package main

import (
	"encoding/json"
	"github.com/Rhymen/go-whatsapp"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

type waHandler struct {
	wac       *whatsapp.Conn
	startTime uint64
	msg       string
	tgBot     *tgbotapi.BotAPI
}

//HandleError needs to be implemented to be a valid WhatsApp handler
func (h *waHandler) HandleError(err error) {
	if e, ok := err.(*whatsapp.ErrConnectionFailed); ok {
		log.Printf("Connection failed, underlying error: %v", e.Err)
		log.Println("Waiting 30sec...")
		<-time.After(30 * time.Second)
		log.Println("Reconnecting...")
		err := h.wac.Restore()
		if err != nil {
			log.Fatalf("Restore failed: %v", err)
		}
	} else {
		if err.Error() == "message type not implemented" {
			return
		}
		log.Fatalf("error occoured: %v\n", err)
	}
}

/*
HandleTextMessage - receives messages
Reply by condition
*/
func (h *waHandler) HandleTextMessage(message whatsapp.TextMessage) {
	chatId, _ := os.LookupEnv("CHAT_ID")
	testChatId, _ := os.LookupEnv("TEST_CHAT_ID")
	senderId, _ := os.LookupEnv("AUTHOR_PHONE")
	search, _ := os.LookupEnv("SEARCH")

	isNewMessage := message.Info.Timestamp >= h.startTime
	isTargetChat := message.Info.RemoteJid != chatId
	isTargetSender := strings.Contains(message.Info.SenderJid, senderId)
	isTargetText := strings.Contains(strings.ToLower(message.Text), search)
	isTargetMessage := isNewMessage && isTargetText && isTargetSender && isTargetChat
	isTestMessage := message.Info.RemoteJid == testChatId && strings.Contains(strings.ToLower(message.Text), "@echo")

	if !isTargetMessage || !isTestMessage {
		return
	}

	msg := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: message.Info.RemoteJid,
		},
		Text: h.msg,
	}

	if _, err := h.wac.Send(msg); err != nil {
		log.Fatalf("error sending message: %v\n", err)
	}

	graceShutDown("✅ Message sent. Terminating", h.tgBot, h.wac)
}

/**
getMessageFromSchedule - reads current schedule from file
return text message fot today to send
*/
func getMessageFromSchedule() (string, error) {
	msg := ""
	byteValue, err := ioutil.ReadFile("./schedule.json")
	if err != nil {
		return msg, err
	}

	type Schedule []struct {
		Day     int    `json:"day"`
		Message string `json:"message"`
	}

	var schedule Schedule
	err = json.Unmarshal(byteValue, &schedule)
	if err != nil {
		return msg, err
	}
	today := int(time.Now().Weekday())

	for _, day := range schedule {
		if day.Day == today {
			return day.Message, nil
		}
	}

	return msg, nil
}
