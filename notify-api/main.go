package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"notify-api/auth"
	"notify-api/mail"
	"os"

	"github.com/go-yaml/yaml"
	"github.com/streadway/amqp"
)

type Message struct {
	Title string `json:"title"`
	Email string `json:"email"`
	Link  string `json:"link"`
	Jwt   string `json:"jwt"`
}

type Setup struct {
	Postgres     string `yaml:"postgres"`
	RabbitMQ     string `yaml:"rabbitmq"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	Database     string `yaml:"database"`
	MailUser     string `yaml:"mailUser"`
	MailPassword string `yaml:"mailPassword"`
}

func readFile() error {
	content, err := ioutil.ReadFile("./setup.yml")
	if err != nil {
		log.Println(err)
		return err
	}
	var info Setup
	if err := yaml.Unmarshal(content, &info); err != nil {
		log.Println(err)
		return err
	}

	os.Setenv("postgres", info.Postgres)
	os.Setenv("rabbitmq", info.RabbitMQ)
	os.Setenv("mailUser", info.MailUser)
	os.Setenv("mailPassword", info.MailPassword)

	return nil
}

func receive() {
	readFile()
	conn, err := amqp.Dial(os.Getenv("rabbitmq"))
	defer conn.Close()
	if err != nil {
		log.Println(err)
		return
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Println(err)
		return
	}
	defer ch.Close()
	qr, err := ch.QueueDeclare(
		"register", false, false, false, false, nil,
	)
	if err != nil {
		log.Println(err)
		return
	}
	msgsRegister, err := ch.Consume(
		qr.Name, "", true, false, false, false, nil,
	)
	if err != nil {
		log.Println(err)
		return
	}
	qt, err := ch.QueueDeclare(
		"transaction", false, false, false, false, nil,
	)
	if err != nil {
		log.Println(err)
		return
	}
	msgsTransaction, err := ch.Consume(
		qt.Name, "", true, false, false, false, nil,
	)
	if err != nil {
		log.Println(err)
		return
	}
	forever := make(chan bool)
	go func() {
		for d := range msgsRegister {
			go func() {
				log.Println("received msg confirmation email ", string(d.Body))
				var message Message
				err = json.Unmarshal([]byte(d.Body), &message)
				if err != nil {
					log.Println(err)
					return
				}
				if ok := auth.CheckAuth(message.Jwt, message.Email, os.Getenv("jwtauth")); ok {
					mail.SendMail(message.Email, message.Link)
				}
			}()
		}
	}()
	go func() {
		for d := range msgsTransaction {
			log.Println("received msg transaction:  ", string(d.Body))
		}
	}()

	<-forever
}

func main() {

	receive()

}
