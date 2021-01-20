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
	c         *whatsapp.Conn
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
		err := h.c.Restore()
		if err != nil {
			log.Fatalf("Restore failed: %v", err)
		}
	} else {
		if err.Error() == "message type not implemented" {
			return
		}
		log.Printf("error occoured: %v\n", err)
	}
}

/*
HandleTextMessage - receives messages
Reply by condition
*/
func (h *waHandler) HandleTextMessage(message whatsapp.TextMessage) {
	chatId, _ := os.LookupEnv("CHAT_ID")
	search, _ := os.LookupEnv("SEARCH")

	if message.Info.Timestamp >= h.startTime {
		fmt.Printf("senderJid: %v, id: %v, remoteJid: %v, \nmessage:\t%v\n", message.Info.SenderJid, message.Info.Id, message.Info.RemoteJid,
			message.Text)
	}

	// TODO: add sender check
	if message.Info.Timestamp < h.startTime || !strings.Contains(strings.ToLower(message.Text), search) || message.Info.RemoteJid != chatId {
		return
	}

	msg := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: message.Info.RemoteJid,
		},
		Text: h.msg,
	}

	if _, err := h.c.Send(msg); err != nil {
		fmt.Fprintf(os.Stderr, "error sending message: %v\n", err)
	}

	fmt.Printf("message sent to user %v\n", message.Info.RemoteJid)
}

/**
getMessageFromSchedule - reads current schedule from file
return text message fot today to send
*/
func getMessageFromSchedule() (string, error) {
	msg := ""
	byteValue, err := ioutil.ReadFile("./schedule.json")
	if err != nil {
		log.Fatalf("schedule error: %v", err)
	}

	type Schedule []struct {
		Day     int    `json:"day"`
		Message string `json:"message"`
	}

	var schedule Schedule
	json.Unmarshal(byteValue, &schedule)
	today := int(time.Now().Weekday())

	for _, day := range schedule {
		if day.Day == today {
			return day.Message, nil
		}
	}

	return msg, nil
}
