package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"net/smtp"
	"os"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"
)

const (
	defaultReplyEmail    = "no-reply@donders.ru.nl"
	defaultSMTPHost      = "smtp-auth.ru.nl"
	defaultSMTPPort      = 587
	defaultRecipientList = "recipients.csv"
)

// Recipient is a data structure of email recipient.
type Recipient struct {
	Email string
	Name  string
}

// ConfigSMTP is a data structure of connecting to SMTP server.
type ConfigSMTP struct {
	SMTPHost string
	SMTPPort int
	SMTPAuth smtp.Auth
}

// EmailTemplate is a data structure of email content.
type EmailTemplate struct {
	Subject *template.Template
	Body    *template.Template
}

// Compose applies the given data on the EmailTemplate, and generates the actual
// subject and body strings.
func (t *EmailTemplate) Compose(data interface{}) (string, string, error) {

	// derive subject from template
	subj := bytes.NewBuffer([]byte{})
	err := t.Subject.Execute(subj, data)
	if err != nil {
		return "", "", err
	}

	// derive body from template
	body := bytes.NewBuffer([]byte{})
	err = t.Body.Execute(body, data)
	if err != nil {
		return "", "", err
	}

	// compose
	return subj.String(), body.String(), nil
}

var optsUserList *string
var optsFromAddr *string
var optsSMTPHost *string
var optsSMTPPort *int
var optsSMTPUser *string
var optsSMTPPass *string

var config ConfigSMTP

func usage() {
	fmt.Printf("\nUsage: %s [OPTIONS] <template>\n", os.Args[0])
	fmt.Printf("\nOPTIONS:\n")
	flag.PrintDefaults()
}

func init() {
	optsUserList = flag.String("l", defaultRecipientList, "set `path` of the file containing a list of recipients.")
	optsFromAddr = flag.String("f", defaultReplyEmail, "set the sender's `email` address.")
	optsSMTPHost = flag.String("n", defaultSMTPHost, "set the network `hostname` of the SMTP server.")
	optsSMTPPort = flag.Int("p", defaultSMTPPort, "set the network `port` of the SMTP server.")
	optsSMTPUser = flag.String("u", "", "set SMTP `username` for PLAIN authentication.")
	optsSMTPPass = flag.String("s", "", "set SMTP `password` for PLAIN authentication.")
	flag.Usage = usage
	flag.Parse()

	// compose the SMTP configuration struct.  The Authentication is only enabled when
	// the optsSMTPUser is set to non-emtpty string by command-line options.
	config = ConfigSMTP{SMTPHost: *optsSMTPHost, SMTPPort: *optsSMTPPort, SMTPAuth: nil}
	if *optsSMTPUser != "" && *optsSMTPPass != "" {
		config.SMTPAuth = smtp.PlainAuth("", *optsSMTPUser, *optsSMTPPass, *optsSMTPHost)
	}

	// setup logger
	log.SetOutput(os.Stderr)
}

// readRecipients reads and constructs Recipient objects from the given CSV file.
// Each line of the file contains two fields separated by a comma ','.  The two fields are
// 1) email address, 2) (display) name.
func readRecipients(csvfile string) ([]Recipient, error) {
	recipients := []Recipient{}

	fd, err := os.Open(csvfile)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	reader := csv.NewReader(fd)
	reader.Comment = '#'
	for {
		uinfo, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Error(err)
			break
		}
		recipients = append(recipients, Recipient{Email: uinfo[0], Name: uinfo[1]})
	}
	return recipients, nil
}

// readTemplate reads email content template from the given path of a file.
func readTemplate(path string) (*EmailTemplate, error) {

	fd, err := os.Open(path)

	if err != nil {
		return nil, err
	}
	defer fd.Close()

	subj := ""
	body := ""
	liner := bufio.NewScanner(fd)
	for liner.Scan() {
		l := liner.Text()
		if strings.HasPrefix(l, "Subject:") {
			subj = strings.TrimSpace(strings.TrimPrefix(l, "Subject:"))
			continue
		}
		body += fmt.Sprintf("%s\n", l)
	}

	if err := liner.Err(); err != nil {
		return nil, err
	}

	tmpl := EmailTemplate{Subject: nil, Body: nil}
	tmpl.Subject, err = template.New("subject").Parse(subj)
	if err != nil {
		return nil, err
	}
	tmpl.Body, err = template.New("body").Parse(body)
	if err != nil {
		return nil, err
	}

	return &tmpl, nil
}

// sendMail sends email content via a SMTP server.
func sendMail(config ConfigSMTP, from, to, subject, body string) error {

	// SMTP server address
	addr := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)

	// RFC-822 style email message
	msg := []byte("Subject: " + subject + "\r\n" +
		body + "\r\n")

	return smtp.SendMail(addr, config.SMTPAuth, from, []string{to}, msg)
}

func main() {

	// gets template file path from argument
	args := flag.Args()
	if len(args) != 1 {
		log.Fatal("missing mail content template")
	}

	// reads subject and body templates
	tmpl, err := readTemplate(args[0])
	if err != nil {
		log.Fatal(err)
	}

	// reads list of recipients for sending emails
	recipients, err := readRecipients(*optsUserList)
	if err != nil {
		log.Fatal(err)
	}

	// loop over recipients to send out an email for each user
	for _, u := range recipients {

		subj, body, err := tmpl.Compose(u)
		if err != nil {
			log.Fatal(err)
		}

		// sends email
		err = sendMail(config, *optsFromAddr, u.Email, subj, body)
		if err != nil {
			log.Fatal(err)
		}
		log.Info(fmt.Sprintf("email sent: %s", u.Email))
	}
}
