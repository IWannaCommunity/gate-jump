package mailer

import (
	"crypto/tls"
	"strconv"
	"time"

	"github.com/IWannaCommunity/gate-jump/src/api/log"
	"github.com/IWannaCommunity/gate-jump/src/api/settings"
	smtp "github.com/go-mail/mail"
)

var mailman *smtp.Dialer

func SMTPInit() error {
	port, _ := strconv.Atoi(settings.Mailer.Port)

	mailman = smtp.NewDialer(settings.Mailer.Host,
		port,
		settings.Mailer.User,
		settings.Mailer.Pass)

	mailman.Timeout = time.Second * 30
	mailman.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	mailman.StartTLSPolicy = smtp.MandatoryStartTLS

	outbox, err := mailman.Dial()
	if err != nil {
		log.Error("Failed dialing the outbox")
		return err
	}
	defer outbox.Close()
	return nil
}
