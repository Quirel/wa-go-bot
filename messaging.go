package main

import (
	"fmt"
	"github.com/Rhymen/go-whatsapp"
	"log"
	"os"
	"strings"
	"time"
)

type waHandler struct {
	c         *whatsapp.Conn
	startTime uint64
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
	if message.Info.FromMe || message.Info.Timestamp < h.startTime || !strings.Contains(strings.ToLower(message.Text), "@echo") {
		return
	}
	fmt.Printf("%v %v %v %v\n\t%v\n", message.Info.Timestamp, message.Info.Id, message.Info.RemoteJid,
		message.ContextInfo.QuotedMessageID, message.Text)

	msg := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: message.Info.RemoteJid,
		},
		Text: "Echo!",
	}

	if _, err := h.c.Send(msg); err != nil {
		fmt.Fprintf(os.Stderr, "error sending message: %v\n", err)
	}

	fmt.Printf("echoed message to user %v\n", message.Info.RemoteJid)
}
