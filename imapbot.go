package imapbot

import (
	"fmt"
	"strings"
	"time"

	"github.com/mxk/go-imap/imap"
)

// NewClient returns a new imap Client
func NewClient(addr, email, password string) (client *Client, err error) {
	client = &Client{}
	client.c, err = dial("imap.mail.me.com:993")
	if err != nil {
		return nil, err
	}
	print(client.c.Data[0].Info)
	client.c.Data = nil
	// Enable encryption, if supported by the server
	if client.c.Caps["STARTTLS"] {
		client.c.StartTLS(nil)
	}
	// Authenticate
	if client.c.State() == imap.Login {
		_, err = client.c.Login(email, password)
		if err != nil {
			return nil, err
		}
	}
	return client, nil
}

// Logout logs out
func (c *Client) Logout() {
	c.c.Logout(30 * time.Second)
}

// Mailboxes returns all mailboxes
func (c *Client) Mailboxes() []*Mailbox {
	cmd, _ := imap.Wait(c.c.List("", "%"))
	mailboxes := make([]*Mailbox, len(cmd.Data))
	for i, rsp := range cmd.Data {
		info := rsp.MailboxInfo()
		mailboxes[i] = &Mailbox{Name: info.Name, client: c}
	}
	// Check for new unilateral server data responses
	for _, rsp := range c.c.Data {
		fmt.Println("Server data:", rsp)
	}
	c.c.Data = nil
	return mailboxes
}

// GetMailbox returns a mailbox with specified name
func (c *Client) GetMailbox(name string) *Mailbox {
	cmd, _ := imap.Wait(c.c.List("", "INBOX"))
	for _, rsp := range cmd.Data {
		info := rsp.MailboxInfo()
		return &Mailbox{Name: info.Name, client: c}
	}
	// Check for new unilateral server data responses
	for _, rsp := range c.c.Data {
		fmt.Println("Server data:", rsp)
	}
	c.c.Data = nil
	return nil
}

func dial(addr string) (c *imap.Client, err error) {
	if strings.HasSuffix(addr, ":993") {
		c, err = imap.DialTLS(addr, nil)
	} else {
		c, err = imap.Dial(addr)
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) processUnilateralServerData() {
	for _, rsp := range c.c.Data {
		fmt.Println("Server data:", rsp)
	}
	c.c.Data = nil
}
