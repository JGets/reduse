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
	EMAIL_USERNAME = "admin@redu.se"
	EMAIL_SERVER = "smtp.gmail.com"
	EMAIL_PORT = "587"
)

var auth smtp.Auth
var emailServerAddr string
var adminEmailAddrsString string
var adminEmailAddrs []string


func initEmail(adminEmails, password string) error{
	if adminEmails == ""{
		return errors.New("No administrator contact email(s) specified")
	}
	
	if password == ""{
		return errors.New("No password specified")
	}
	
	auth = smtp.PlainAuth("", EMAIL_USERNAME, password, EMAIL_SERVER)
	emailServerAddr = EMAIL_SERVER + ":" + EMAIL_PORT
	adminEmailAddrsString = adminEmails
	adminEmailAddrs = parseAddresses(adminEmails)
	
	return nil
}

func parseAddresses(addresses string) []string{
	return strings.Split(addresses, ",")
}

func sendEmailToAdmins(subject, body string) error{
	emailBody := "To: "+ adminEmailAddrsString + "\r\nSubject: " + subject + "\r\n\r\n" + body
	
	err := smtp.SendMail(emailServerAddr, auth, EMAIL_USERNAME, adminEmailAddrs, []byte(emailBody))
	
	return err
}
