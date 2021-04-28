package main

import (
	"account-api/api"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/go-yaml/yaml"
)

type Setup struct {
	Postgres string `yaml:"postgres"`
	RabbitMQ string `yaml:"rabbitmq"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
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

	return nil
}

func main() {

	fmt.Println("Starting the application ... ")
	readFile()
	http.HandleFunc("/register", api.Register)
	http.HandleFunc("/login", api.Login)
	http.HandleFunc("/wallets", api.Wallets)
	http.HandleFunc("/transfer", api.Transfer)
	http.HandleFunc("/transactions", api.Transactions)
	http.HandleFunc("/transactions/date", api.SearchTransactionByDate)
	http.HandleFunc("/mail", api.MailConfirmation)
	http.ListenAndServe(":9123", nil)
}
