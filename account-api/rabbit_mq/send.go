package rabbit_mq

import (
	"encoding/json"
	"log"
	"os"

	"github.com/streadway/amqp"
)

func SendMessage(name string, message interface{}) error {
	conn, err := amqp.Dial(os.Getenv("rabbitmq"))
	defer conn.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	ch, err := conn.Channel()
	defer ch.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	q, err := ch.QueueDeclare(
		name, false, false, false, false, nil,
	)
	if err != nil {
		log.Println(err)
		return err
	}
	out, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
		return err
	}
	err = ch.Publish(
		"", q.Name, false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(out),
		})
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Sent : ", message)
	return nil
}
