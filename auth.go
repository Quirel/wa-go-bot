package main

import (
	"encoding/gob"
	"fmt"
	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/Rhymen/go-whatsapp"
	"os"
)

const sessionFilePath = "./whatsappSession.gob"

/*
readSession - reads previously saved session from file
*/
func readSession() (whatsapp.Session, error) {
	session := whatsapp.Session{}
	file, err := os.Open(sessionFilePath)
	if err != nil {
		return session, err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&session)
	if err != nil {
		return session, err
	}
	return session, nil
}

/*
writeSession - writes successful session to file
*/
func writeSession(session whatsapp.Session) error {
	file, err := os.Create(sessionFilePath)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(session)
	if err != nil {
		return err
	}
	return nil
}

/*
login - restores session or create new by QR-code scan
*/
func login(wac *whatsapp.Conn) error {
	//load saved session
	session, err := readSession()
	if err == nil {
		//restore session
		session, err = wac.RestoreWithSession(session)
		if err != nil {
			return fmt.Errorf("restoring failed: %v\n", err)
		}
	} else {
		//no saved session -> regular login
		qr := make(chan string)
		go func() {
			terminal := qrcodeTerminal.New()
			terminal.Get(<-qr).Print()
		}()
		session, err = wac.Login(qr)
		if err != nil {
			return fmt.Errorf("error during login: %v\n", err)
		}
	}

	//save session
	err = writeSession(session)
	if err != nil {
		return fmt.Errorf("error saving session: %v\n", err)
	}

	fmt.Println("login successful")
	fmt.Println("============================")
	return nil
}
