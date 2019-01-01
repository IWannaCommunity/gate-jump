package mailer

import (
	"time"

	"github.com/IWannaCommunity/gate-jump/src/api/log"
	smtp "github.com/go-mail/mail"
)

var Outbox chan *smtp.Message

func Daemon() {
	Outbox = make(chan *smtp.Message)
	sender := *new(smtp.SendCloser)
	err := *new(error)
	open := false

	for {
		select {
		case msg, ok := <-Outbox:
			if !ok {
				return
			}
			if !open {
				if sender, err = mailman.Dial(); err != nil {
					log.Fatal("Could not communicate with local SMTP host, %v", err)
				}
				open = true
			}
			if err := smtp.Send(sender, msg); err != nil {
				log.Error("Could not send email, %v", err)
			}

		case <-time.After(30 * time.Second):
			if open {
				if err := sender.Close(); err != nil {
					log.Fatal("Could not close connection with local SMTP host, %v", err)
				}
				open = false
			}
		}
	}
}
