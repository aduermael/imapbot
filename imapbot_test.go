package imapbot

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(t *testing.T) {
	var err error
	logLevel = logLevelDebug

	email := os.Getenv("EMAIL")
	password := os.Getenv("PASSWORD")

	c, err := NewClient("imap.mail.me.com:993", email, string(password))
	if err != nil {
		printFatal(err)
	}
	defer c.Logout()

	inbox := c.GetMailbox("INBOX")
	if inbox != nil {
		mails, err := inbox.Mails()
		if err != nil {
			printFatal(err)
		}
		for _, mail := range mails {
			if mail.Subject == "Votre export des contributions pour le projet comme-convenu-2" {
				fmt.Println("FOUND IT! ->", mail.Subject, mail.uid)
				err := mail.LoadParts("")
				if err != nil {
					printFatal(err)
				}
				mail.PrintParts()
				err = mail.Delete()
				if err != nil {
					printFatal(err)
				}
				break
			}
		}
	}
}
