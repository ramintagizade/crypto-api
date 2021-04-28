package mail

import (
	"log"
	"net/smtp"
	"os"
)

func SendMail(receiver string, link string) {
	from := os.Getenv("mailUser")
	password := os.Getenv("mailPassword")
	to := []string{
		receiver,
	}
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	message := []byte("Confirm your email : " + link)
	auth := smtp.PlainAuth("", from, password, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Email sent")
}
