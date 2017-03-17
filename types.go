package imapbot

import (
	"time"

	"github.com/mxk/go-imap/imap"
)

// Client is an imap client
type Client struct {
	c *imap.Client
}

// Mailbox represents a mail box
type Mailbox struct {
	Name   string
	client *Client
}

// Mail represents an email
type Mail struct {
	Subject string
	Parts   *MailPart
	// the uid is only valid for one client session, don't store it!
	uid     uint32
	size    uint32
	date    time.Time
	client  *Client
	mailbox *Mailbox
}

// MailPart contains what can be found in one part of an email
type MailPart struct {
	ContentType string
	Data        []byte
	Children    []*MailPart
	Attachment  bool
	Filename    string
}
