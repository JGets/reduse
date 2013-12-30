package main

import(
	"net/smtp"
	// "log"
	// "os"
	"strings"
	"errors"
)

// var serverAddress string

const(
	EMAIL_SERVER = "smtp.gmail.com"
	EMAIL_PORT = "587"
)

var auth smtp.Auth
var emailServerAddr, adminEmailAddrsString, emailUsername string
var adminEmailAddrs []string


func initEmail(adminEmails, usersername, password string) error{
	if adminEmails == ""{
		return errors.New("No administrator contact email(s) specified")
	}
	
	if usersername == ""{
		return errors.New("No email username specified")
	}
	
	if password == ""{
		return errors.New("No password specified")
	}
	
	auth = smtp.PlainAuth("", usersername, password, EMAIL_SERVER)
	emailServerAddr = EMAIL_SERVER + ":" + EMAIL_PORT
	adminEmailAddrsString = adminEmails
	emailUsername = usersername
	adminEmailAddrs = parseAddresses(adminEmails)
	
	return nil
}

func parseAddresses(addresses string) []string{
	return strings.Split(addresses, ",")
}

func sendEmailToAdmins(subject, body string) error{
	if devMode {
		subject = "(DEV) " + subject
		body = "(DEV)\r\n" + body
	}

	emailBody := "To: "+ adminEmailAddrsString + "\r\nSubject: " + subject + "\r\n\r\n" + body
	
	err := smtp.SendMail(emailServerAddr, auth, emailUsername, adminEmailAddrs, []byte(emailBody))
	
	return err
}

func sendHTMLEmailToAdmins(subject, body string) error {

	if devMode {
		subject = "(DEV) " + subject
		body = "<strong>(DEV)</strong><br/>" + body
	}

	mime := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n";

	emailBody := "To: "+ adminEmailAddrsString + "\r\nSubject: " + subject + "\r\n" + mime + body
	
	err := smtp.SendMail(emailServerAddr, auth, emailUsername, adminEmailAddrs, []byte(emailBody))
	
	return err
}
