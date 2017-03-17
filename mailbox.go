package imapbot

import (
	"bytes"
	"fmt"
	"mime"
	"net/mail"
	"strings"

	"github.com/mxk/go-imap/imap"
)

// NbMails returns the amount of emails of given mailbox
func (mb *Mailbox) NbMails() uint32 {
	mb.client.c.Select(mb.Name, true)
	return mb.client.c.Mailbox.Messages
}

// Mails returns all emails of given mailbox
func (mb *Mailbox) Mails() ([]*Mail, error) {
	mb.client.c.Select(mb.Name, true)
	mb.client.c.Check()

	set, _ := imap.NewSeqSet("")
	set.Add("1:*")
	// TODO: pagination

	return mb.getMailsHeaderOnly(set, false)
}

func (mb *Mailbox) getMailsHeaderOnly(set *imap.SeqSet, uids bool) ([]*Mail, error) {

	var cmd *imap.Command
	var err error

	if uids {
		cmd, err = mb.client.c.UIDFetch(set, "FLAGS", "INTERNALDATE", "RFC822.SIZE", "BODY[HEADER]", "UID")
		if err != nil {
			return nil, err
		}
	} else {
		cmd, err = mb.client.c.Fetch(set, "FLAGS", "INTERNALDATE", "RFC822.SIZE", "BODY[HEADER]", "UID")
		if err != nil {
			return nil, err
		}
	}

	mails := make([]*Mail, 0)

	for cmd.InProgress() {
		// Wait for the next response (no timeout)
		mb.client.c.Recv(-1)

		// Process command data
		for _, rsp := range cmd.Data {
			header := imap.AsBytes(rsp.MessageInfo().Attrs["BODY[HEADER]"])
			if msg, err := mail.ReadMessage(bytes.NewReader(header)); msg != nil && err == nil {

				subject := msg.Header.Get("Subject")
				words := strings.Split(subject, " ")

				dec := new(mime.WordDecoder)
				for i, word := range words {
					w, err := dec.Decode(word)
					if err == nil {
						words[i] = w
					}
				}
				subject = strings.Join(words, " ")

				mails = append(mails, &Mail{
					Subject: subject,
					uid:     rsp.MessageInfo().UID,
					size:    rsp.MessageInfo().Size,
					date:    rsp.MessageInfo().InternalDate,
					client:  mb.client,
					mailbox: mb,
					Parts:   nil,
				})
			}
			// TODO: handle errors
		}
		cmd.Data = nil

		mb.client.processUnilateralServerData()
	}

	// Check command completion status
	if _, err := cmd.Result(imap.OK); err != nil {
		if err == imap.ErrAborted {
			// Fetch command aborted
			return mails, nil
		}
		// more info in rsp.Info
		return nil, err
	}

	return mails, nil
}

// utils

func report(cmd *imap.Command, err error) *imap.Command {
	var rsp *imap.Response
	if cmd == nil {
		fmt.Printf("--- ??? ---\n%v\n\n", err)
		panic(err)
	} else if err == nil {
		rsp, err = cmd.Result(imap.OK)
	}
	if err != nil {
		fmt.Printf("--- %s ---\n%v\n\n", cmd.Name(true), err)
		panic(err)
	}
	c := cmd.Client()
	fmt.Printf("--- %s ---\n"+
		"%d command response(s), %d unilateral response(s)\n"+
		"%s %s\n\n",
		cmd.Name(true), len(cmd.Data), len(c.Data), rsp.Status, rsp.Info)
	c.Data = nil
	return cmd
}
