package imapbot

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/mail"
	"strconv"
	"strings"

	"github.com/mxk/go-imap/imap"
)

func (m *Mail) Delete() error {
	if m.uid == 0 {
		return errors.New("uid == 0")
	}
	m.client.c.Select(m.mailbox.Name, false)
	m.client.c.Check()
	set, _ := imap.NewSeqSet("")
	set.AddNum(m.uid)
	report(m.client.c.UIDStore(set, "+FLAGS.SILENT", imap.NewFlagSet(`\Deleted`)))
	report(m.client.c.Expunge(set))
	return nil
}

// DownloadAll downloads all email content
func (m *Mail) DownloadAll() error {
	return errors.New("work in progress")
}

// LoadParts registers all different parts but only loads ones with
// specified content type
func (m *Mail) LoadParts(contentType string) (err error) {

	m.client.c.Select(m.mailbox.Name, true)
	m.client.c.Check()

	set, _ := imap.NewSeqSet("")
	set.AddNum(m.uid)

	var cmd *imap.Command
	cmd, err = m.client.c.UIDFetch(set, "FLAGS", "INTERNALDATE", "RFC822.SIZE", "BODY[]")
	if err != nil {
		return err
	}

	for cmd.InProgress() {
		// Wait for the next response (no timeout)
		m.client.c.Recv(-1)

		// Process command data
		for _, rsp := range cmd.Data {

			mailBytes := imap.AsBytes(rsp.MessageInfo().Attrs["BODY[]"])

			if msg, _ := mail.ReadMessage(bytes.NewReader(mailBytes)); msg != nil {
				m.Parts, err = loadParts(msg, "text/csv")
			}
		}
		cmd.Data = nil

		m.client.processUnilateralServerData()
	}

	// Check command completion status
	if rsp, err := cmd.Result(imap.OK); err != nil {
		if err == imap.ErrAborted {
			fmt.Println("Fetch command aborted")
		} else {
			fmt.Println("Fetch error:", rsp.Info)
		}
	}

	return err
}

func loadParts(i interface{}, contentType string) (*MailPart, error) {

	var part *MailPart = nil

	if msg, ok := i.(*mail.Message); ok {
		mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
		if err != nil {
			return nil, err
		}

		mediaDisposition, paramsDisposition, _ := mime.ParseMediaType(msg.Header.Get("Content-Disposition"))

		if strings.HasPrefix(mediaType, "multipart/") {
			mr := multipart.NewReader(msg.Body, params["boundary"])

			part = &MailPart{ContentType: mediaType, Children: make([]*MailPart, 0), Data: nil, Attachment: false, Filename: ""}
			for {
				var p *multipart.Part
				p, err = mr.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					return nil, err
				}
				var child *MailPart
				child, err = loadParts(p, contentType)
				if err != nil {
					return nil, err
				}
				part.Children = append(part.Children, child)
			}
		} else {
			attachment := mediaDisposition == "attachment"
			filename := ""
			if attachment {
				filename = paramsDisposition["filename"]
			}
			part = &MailPart{ContentType: mediaType, Children: nil, Data: nil, Attachment: attachment, Filename: filename}
			var err error
			// load only if requested
			if contentType == "" || contentType == mediaType {
				part.Data, err = ioutil.ReadAll(msg.Body)
			}
			if err != nil {
				return nil, err
			}
		}

	} else if p, ok := i.(*multipart.Part); ok {
		mediaType, params, err := mime.ParseMediaType(p.Header.Get("Content-Type"))
		if err != nil {
			return nil, err
		}
		mediaDisposition, paramsDisposition, _ := mime.ParseMediaType(p.Header.Get("Content-Disposition"))

		if strings.HasPrefix(mediaType, "multipart/") {
			mr := multipart.NewReader(p, params["boundary"])
			part = &MailPart{ContentType: mediaType, Children: make([]*MailPart, 0), Data: nil}
			for {
				var p *multipart.Part
				p, err = mr.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					return nil, err
				}

				var child *MailPart
				child, err = loadParts(p, contentType)
				if err != nil {
					return nil, err
				}
				part.Children = append(part.Children, child)
			}
		} else {
			attachment := mediaDisposition == "attachment"
			filename := ""
			if attachment {
				filename = paramsDisposition["filename"]
			}
			part = &MailPart{ContentType: mediaType, Children: nil, Data: nil, Attachment: attachment, Filename: filename}
			var err error
			// load only if requested
			if contentType == "" || contentType == mediaType {
				part.Data, err = ioutil.ReadAll(p)
			}
			if err != nil {
				return nil, err
			}
		}
	} else {
		return nil, errors.New("unknown interface")
	}

	return part, nil

}

// PrintParts can be used for debug to print loaded parts
func (m *Mail) PrintParts() {
	printParts(m.Parts, 0)
}

func printParts(part *MailPart, level int) {
	if part == nil {
		return
	}
	prefix := ""
	for i := 0; i < level; i++ {
		prefix += "  "
	}
	data := "nil"
	if part.Data != nil {
		data = strconv.Itoa(len(part.Data)) + " bytes"
	}
	fmt.Println(prefix+part.ContentType, "data:", data, "attachment:", part.Attachment, part.Filename)

	for _, child := range part.Children {
		printParts(child, level+1)
	}
}
