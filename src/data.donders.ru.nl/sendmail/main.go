package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
	"strings"
	"text/template"
)

const (
	EMAIL_NOREPLY = "no-reply@donders.ru.nl"
	EMAIL_SUPPORT = "datasupport@donders.ru.nl"
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

var opts_userlist *string
var opts_smtpHost *string
var opts_smtpPort *int
var opts_smtpUser *string
var opts_smtpPass *string

var config ConfigSMTP

func init() {
	opts_userlist = flag.String("l", "", "set path of file containing list of user emails and names.")
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

// sendMail sends email content via a SMTP server.
func sendMail(config ConfigSMTP, from, to, subject, body string) error {

	// SMTP server address
	addr := fmt.Sprintf("%s:%d", config.SmtpHost, config.SmtpPort)

	// RFC-822 style email message
	msg := []byte("To: " + EMAIL_SUPPORT + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	return smtp.SendMail(addr, config.SmtpAuth, from, []string{to}, msg)
}

func main() {

	users, err := readUsers(*opts_userlist)
	if err != nil {
		log.Fatal(err)
	}

	// gets template file path from argument
	args := flag.Args()
	if len(args) != 1 {
		log.Fatal("missing mail content template")
	}

	// reads template
	mailbody, err := ioutil.ReadFile(args[0])
	if err != nil {
		log.Fatal(err)
	}

	// creates template parser
	tmpl, err := template.New("mailbody").Parse(string(mailbody))
	if err != nil {
		log.Fatal(err)
	}

	// loop over users to send out an email for each user
	for _, u := range users {
		w := bytes.NewBuffer([]byte{})
		err := tmpl.Execute(w, u)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", string(w.Bytes()))
		// sends email
		//sendMail(config, EMAIL_NOREPLY, u.Email, "", string(buf))
	}
}
