package mailrelay

import (
	// "fmt"
	"appengine"
	amail "appengine/mail"
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/mail"
	"path"
	"strings"
)

func init() {
	http.HandleFunc("/_ah/mail/", incomingMail)
}

func incomingMail(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	// var b bytes.Buffer
	// if _, err := b.ReadFrom(r.Body); err != nil {
	// c.Errorf("Error reading body: %v", err)
	// return
	// }
	// c.Infof("Received mail: %v", b.String())

	_, address := path.Split(r.URL.Path)

	var userFileContent []byte
	userFileContent, _ = ioutil.ReadFile("users.txt")

	var userFileString string
	userFileString = bytes.NewBuffer(userFileContent).String()
	users := strings.Split(userFileString, "\n")
	// c.Infof("Users %v", users)
	var validUser bool = false
	var destinationAddress string

	for iter := 0; iter < len(users); iter++ {
		if items := strings.Split(users[iter], ":"); items[0]+"@kemailtranslator.appspotmail.com" == address {
			validUser = true
			destinationAddress = items[1]
		}
	}

	if validUser == false {
		c.Errorf("Invalid user: %v", address)
		return
	}

	msg, err := mail.ReadMessage(r.Body)
	if err != nil {
		c.Errorf("ERROR PROCESSING BODY: %v", err)
	}

	var body bytes.Buffer

	contentType := msg.Header.Get("Content-Type")
	items := strings.Split(contentType, ";")

	if items[0] == "text/plain" {
		body.ReadFrom(msg.Body)
	} else {
		bsplit := strings.Split(items[1], "=")
		boundary := bsplit[1]

		reader := multipart.NewReader(msg.Body, boundary)
		part, _ := reader.NextPart()

		body.ReadFrom(part)
	}

	subject := msg.Header.Get("Subject")

	attachment := amail.Attachment{
		Name: subject + ".txt",
		Data: body.Bytes(),
	}

	email := &amail.Message{
		Sender:      "angelo@dinardi.name",
		To:          []string{destinationAddress},
		Subject:     subject,
		Body:        "attached",
		Attachments: []amail.Attachment{attachment},
	}

	if err := amail.Send(c, email); err != nil {
		c.Errorf("Alas, my user, the email failed to sendeth: %v", err)
	}
}
