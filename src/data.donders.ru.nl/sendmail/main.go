package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
	"text/template"
)

const (
	EMAIL_NOREPLY = "no-reply@donders.ru.nl"
)

type User struct {
	Email string
	Name  string
}

type ConfigSMTP struct {
	SmtpHost string
	SmtpPort int
	SmtpAuth smtp.Auth
}

type Template struct {
	Subject string
	Body string
}

var opts_userList *string
var opts_fromAddr *string
var opts_smtpHost *string
var opts_smtpPort *int
var opts_smtpUser *string
var opts_smtpPass *string

var config ConfigSMTP

func init() {
	opts_userList = flag.String("l", "", "set path of file containing list of user emails and names.")
	opts_fromAddr = flag.String("f", EMAIL_NOREPLY, "set the sender's email address.")
	opts_smtpHost = flag.String("h", "smtp-auth.ru.nl", "set the network hostname of the SMTP server.")
	opts_smtpPort = flag.Int("p", 25, "set the network port of the SMTP server.")
	opts_smtpUser = flag.String("u", "", "set SMTP username for PLAIN authentication.")
	opts_smtpPass = flag.String("s", "", "set SMTP password for PLAIN authentication.")
	flag.Parse()

	// compose the SMTP configuration struct.  The Authentication is only enabled when
	// the opts_smtpUser is set to non-emtpty string by command-line options.
	config = ConfigSMTP{SmtpHost: *opts_smtpHost, SmtpPort: *opts_smtpPort, SmtpAuth: nil}
	if *opts_smtpUser != "" && *opts_smtpPass != "" {
		config.SmtpAuth = smtp.PlainAuth("", *opts_smtpUser, *opts_smtpPass, *opts_smtpHost)
	}

}

// readUsers reads and constructs User objects from an input file referred by path.
// Each line of the file contains two fields separated by spaces, or tabs.  The two fields are
// 1) email address, 2) username.
func readUsers(path string) ([]User, error) {

	users := []User{}

	fd, err := os.Open(path)

	if err != nil {
		return nil, err
	}
	defer fd.Close()

	liner := bufio.NewScanner(fd)
	for liner.Scan() {
		uinfo := strings.SplitN(liner.Text(), " ", 2)
		users = append(users, User{Email: uinfo[0], Name: uinfo[1]})
	}

	if err := liner.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func readTemplate(path string) (*Template, error) {

	template := Template{Subject: "", Body: ""}

	fd, err := os.Open(path)

	if err != nil {
		return nil, err
	}
	defer fd.Close()

	liner := bufio.NewScanner(fd)
	for liner.Scan() {
		l := liner.Text()
		if strings.HasPrefix(l, "Subject:") {
			template.Subject = strings.TrimSpace(strings.TrimPrefix(l, "Subject:"))
			continue
		}
		template.Body += fmt.Sprintf("%s\n", l)
	}

	if err := liner.Err(); err != nil {
		return nil, err
	}

	return &template, nil
}

// sendMail sends email content via a SMTP server.
func sendMail(config ConfigSMTP, from, to, subject, body string) error {

	// SMTP server address
	addr := fmt.Sprintf("%s:%d", config.SmtpHost, config.SmtpPort)

	// RFC-822 style email message
	msg := []byte("Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	return smtp.SendMail(addr, config.SmtpAuth, from, []string{to}, msg)
}

func main() {

	// reads list of users for sending emails
	users, err := readUsers(*opts_userList)
	if err != nil {
		log.Fatal(err)
	}

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

	// creates template parsers
	pSubj, err := template.New("subj").Parse(string(tmpl.Subject))
	if err != nil {
		log.Fatal(err)
	}

	pBody, err := template.New("body").Parse(string(tmpl.Body))
	if err != nil {
		log.Fatal(err)
	}

	// loop over users to send out an email for each user
	for _, u := range users {

		// derive subject from template
		subj := bytes.NewBuffer([]byte{})
		err := pSubj.Execute(subj, u)
		if err != nil {
			log.Fatal(err)
		}

		// derive body from template
		body := bytes.NewBuffer([]byte{})
		err = pBody.Execute(body, u)
		if err != nil {
			log.Fatal(err)
		}

		// sends email
		err = sendMail(config, *opts_fromAddr, u.Email, string(subj.Bytes()), string(body.Bytes()))
		if err != nil {
			log.Fatal(err)
		}
	}
}
