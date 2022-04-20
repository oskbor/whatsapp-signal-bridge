package main

import (
	"context"
	"fmt"
	"os"
	osSignal "os/signal"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"github.com/oskbor/bridge/glue"
	"github.com/oskbor/bridge/signal"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func main() {
	// get environment variable
	signalHost := os.Getenv("SIGNAL_HOST")
	signalNumber := os.Getenv("SIGNAL_NUMBER")
	recipient := os.Getenv("SIGNAL_RECIPIENT")
	signalClient, err := signal.NewClient(signal.Host(signalHost), signal.Number(signalNumber))
	if err != nil {
		panic(err)
	}
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container, err := sqlstore.New("sqlite3", "file:./bridge/whatsmeow.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}

	clientLog := waLog.Stdout("Whatsmeow", "WARN", true)
	whatsappClient := whatsmeow.NewClient(deviceStore, clientLog)

	if whatsappClient.Store.ID == nil {
		fmt.Println("No device found, QR flow started")
		// No ID stored, new login
		qrChan, _ := whatsappClient.GetQRChannel(context.Background())
		err = whatsappClient.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				fmt.Println("QR code:", evt.Code)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = whatsappClient.Connect()
		if err != nil {
			panic(err)
		}
	}
	glue.New(whatsappClient, signalClient, glue.SignalRecipient(recipient))

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal, 1)
	osSignal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	whatsappClient.Disconnect()

}
