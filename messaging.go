package main

import (
	"encoding/json"
	"fmt"
	"github.com/Rhymen/go-whatsapp"
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
}

//HandleError needs to be implemented to be a valid WhatsApp handler
func (h *waHandler) HandleError(err error) {
	if e, ok := err.(*whatsapp.ErrConnectionFailed); ok {
		log.Printf("Connection failed, underlying error: %v", e.Err)
		log.Println("Waiting 30sec...")
		<-time.After(30 * time.Second)
		log.Println("Reconnecting...")
		tgLogger.Warn("Reconnecting...")
		err := h.wac.Restore()
		if err != nil {
			tgLogger.Error(fmt.Sprintf("Restore failed: %v", err))
			log.Fatalf("Restore failed: %v", err)
		}
	} else {
		if err.Error() == "message type not implemented" {
			return
		}
		tgLogger.Error(fmt.Sprintf("error occoured: %v\n", err))
		log.Fatalf("error occoured: %v\n", err)
	}
}

/*
HandleTextMessage - receives messages
Reply by condition
*/
func (h *waHandler) HandleTextMessage(message whatsapp.TextMessage) {
	chatId := os.Getenv("CHAT_ID")
	testChatId := os.Getenv("TEST_CHAT_ID")
	senderId := os.Getenv("AUTHOR_PHONE")
	search := os.Getenv("SEARCH")

	isNewMessage := message.Info.Timestamp >= h.startTime
	isTargetChat := message.Info.RemoteJid == chatId
	isTargetSender := strings.Contains(message.Info.SenderJid, senderId)
	isTargetText := strings.Contains(strings.ToLower(message.Text), search)
	isTestMessage := isNewMessage && message.Info.RemoteJid == testChatId && strings.Contains(strings.ToLower(message.Text), "@echo")
	isTargetMessage := isNewMessage && isTargetText && isTargetSender && isTargetChat
	doSkipMessage := !isTargetMessage && !isTestMessage

	if isDebug {
		// Additional logs
		if message.Info.RemoteJid == testChatId && isNewMessage {
			fmt.Printf("Mssage text:\n%v\n---\n", message.Text)
			fmt.Printf("time: %v, chatId: %v, senderId: %v\n",
				message.Info.Timestamp, message.Info.RemoteJid, message.Info.SenderJid)
			fmt.Println("============================")
		}
	}

	if doSkipMessage {
		return
	}

	msg := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: message.Info.RemoteJid,
		},
		Text: h.msg,
	}
	if isTestMessage {
		msg = whatsapp.TextMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: message.Info.RemoteJid,
			},
			Text: "Ping is OK!",
		}
	}

	if _, err := h.wac.Send(msg); err != nil {
		tgLogger.Error(fmt.Sprintf("error sending message: %v\n", err))
		log.Fatalf("error sending message: %v\n", err)
	}

	if !isDebug && !isTestMessage {
		graceShutDown("✅ Message sent. Terminating", h.wac)
	}
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

	// My week starts from monday
	if today == 0 {
		today = 7
	}

	for _, day := range schedule {
		if day.Day == today {
			return day.Message, nil
		}
	}

	return msg, nil
}
