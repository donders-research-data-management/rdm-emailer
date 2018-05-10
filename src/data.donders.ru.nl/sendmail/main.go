package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"text/template"
	log "github.com/sirupsen/logrus"
)

const (
	EMAIL_NOREPLY = "no-reply@donders.ru.nl"
)

// Recipient is a data structure of email recipient.
type Recipient struct {
	Email string
	Name  string
}

// ConfigSMTP is a data structure of connecting to SMTP server.
type ConfigSMTP struct {
	SmtpHost string
	SmtpPort int
	SmtpAuth smtp.Auth
}

// Template is a data structure of email content.
type EmailTemplate struct {
	Subject *template.Template
	Body *template.Template
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

var opts_userList *string
var opts_fromAddr *string
var opts_smtpHost *string
var opts_smtpPort *int
var opts_smtpUser *string
var opts_smtpPass *string

var config ConfigSMTP

func usage() {
	fmt.Printf("\nUsage: %s [OPTIONS] <template>\n", os.Args[0])
	fmt.Printf("\nOPTIONS:\n")
	flag.PrintDefaults()
}

func init() {
	opts_userList = flag.String("l", "", "set `path` of the file containing a list of recipients.")
	opts_fromAddr = flag.String("f", EMAIL_NOREPLY, "set the sender's `email` address.")
	opts_smtpHost = flag.String("n", "smtp-auth.ru.nl", "set the network `hostname` of the SMTP server.")
	opts_smtpPort = flag.Int("p", 25, "set the network `port` of the SMTP server.")
	opts_smtpUser = flag.String("u", "", "set SMTP `username` for PLAIN authentication.")
	opts_smtpPass = flag.String("s", "", "set SMTP `password` for PLAIN authentication.")
	flag.Usage = usage
	flag.Parse()

	// compose the SMTP configuration struct.  The Authentication is only enabled when
	// the opts_smtpUser is set to non-emtpty string by command-line options.
	config = ConfigSMTP{SmtpHost: *opts_smtpHost, SmtpPort: *opts_smtpPort, SmtpAuth: nil}
	if *opts_smtpUser != "" && *opts_smtpPass != "" {
		config.SmtpAuth = smtp.PlainAuth("", *opts_smtpUser, *opts_smtpPass, *opts_smtpHost)
	}

	// setup logger
	log.SetOutput(os.Stderr)
}

// readRecipients reads and constructs Recipient objects from the given path of a file.
// Each line of the file contains two fields separated by spaces, or tabs.  The two fields are
// 1) email address, 2) username.
func readRecipients(path string) ([]Recipient, error) {

	recipients := []Recipient{}

	fd, err := os.Open(path)

	if err != nil {
		return nil, err
	}
	defer fd.Close()

	liner := bufio.NewScanner(fd)
	for liner.Scan() {
		l := strings.TrimSpace(liner.Text())

		// ignore the empty line or the line with '#' as the first charactor
		if len(l) == 0 || []rune(l)[0] == '#' {
			continue
		}

		uinfo := strings.SplitN(liner.Text(), " ", 2)
		// ignore the line with less than 2 fileds separated by space
		if len(uinfo) < 2 {
			log.Warn(fmt.Sprint("invalid recipient: %s", l))
			continue
		}
		recipients = append(recipients, Recipient{Email: uinfo[0], Name: uinfo[1]})
	}

	if err := liner.Err(); err != nil {
		return nil, err
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
	addr := fmt.Sprintf("%s:%d", config.SmtpHost, config.SmtpPort)

	// RFC-822 style email message
	msg := []byte("Subject: " + subject + "\r\n" +
		body + "\r\n")

	return smtp.SendMail(addr, config.SmtpAuth, from, []string{to}, msg)
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
	recipients, err := readRecipients(*opts_userList)
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
		err = sendMail(config, *opts_fromAddr, u.Email, subj, body)
		if err != nil {
			log.Fatal(err)
		}
		log.Info(fmt.Sprintf("email sent: %s", u.Email))
	}
}
