package api

import (
	"account-api/rabbit_mq"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type Transaction struct {
	Amount     float64   `json:"amount"`
	Commission float64   `json:"commission"`
	Date       time.Time `json:"date"`
	Currency   string    `json:"currency"`
	Sender     string    `json:"sender"`
	Recipient  string    `json:"recipient"`
}

type TransactionList struct {
	Address string `json:"address"`
	Page    int    `json:"page"`
	Date    string `json:"date"`
}

func getSenderEmail(sender string) string {
	db, err := sql.Open("postgres", os.Getenv("postgres"))
	if err != nil {
		log.Println(err)
		return ""
	}
	defer db.Close()
	var email string
	sqlStatement := `SELECT email FROM Wallets WHERE address=$1`
	err = db.QueryRow(sqlStatement, sender).Scan(&email)
	if err != nil {
		return ""
	}
	return email
}

func createTransaction(transaction Transaction) error {
	db, err := sql.Open("postgres", os.Getenv("postgres"))
	if err != nil {
		log.Println(err)
		return err
	}
	defer db.Close()
	sqlStatement := `INSERT INTO Transactions(amount,commission,date,currency,sender,recipient) VALUES($1,$2,$3,$4,$5,$6)`
	_, err = db.Query(sqlStatement, transaction.Amount, transaction.Commission, transaction.Date, transaction.Currency, transaction.Sender, transaction.Recipient)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func Transfer(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		w.WriteHeader(404)
		return
	}
	jwt := req.Header.Get("jwt")
	var transaction Transaction
	json.NewDecoder(req.Body).Decode(&transaction)
	transaction.Currency = strings.ToLower(transaction.Currency)
	currencies := make(map[string]bool)
	currencies["btc"] = true
	currencies["eth"] = true
	if !currencies[transaction.Currency] || transaction.Sender == transaction.Recipient {
		w.WriteHeader(404)
		return
	}
	var email string = getSenderEmail(transaction.Sender)
	if email != "" && CheckAuth(jwt, email, os.Getenv("jwtauth")) {
		confirmed, err := GetMailStatusByEmail(email)
		if err != nil {
			log.Println(err)
			w.WriteHeader(404)
			return
		}
		if confirmed {
			var sender_fee float64 = (transaction.Amount*0.01)*0.01 + transaction.Amount
			sender_balance, err := GetBalance(transaction.Sender)
			if err != nil {
				log.Println(err)
				w.WriteHeader(404)
				w.Write([]byte(err.Error()))
				return
			}
			if sender_balance >= sender_fee {
				recipient_balance, err := GetBalance(transaction.Recipient)
				if err != nil {
					log.Println(err)
					w.WriteHeader(404)
					w.Write([]byte(err.Error()))
					return
				}

				recipient_balance += transaction.Amount
				sender_balance -= sender_fee
				transaction.Commission = 0.01
				transaction.Date = time.Now()
				err = createTransaction(transaction)

				if err != nil {
					log.Println(err)
					w.WriteHeader(404)
					w.Write([]byte("Transaction error : " + err.Error()))
					return
				}
				err = UpdateBalance(transaction.Sender, sender_balance, transaction.Currency)
				if err != nil {
					log.Println(err)
					w.WriteHeader(404)
					return
				}
				err = UpdateBalance(transaction.Recipient, recipient_balance, transaction.Currency)
				if err != nil {
					log.Println(err)
					w.WriteHeader(404)
					return
				}
				var message Transaction = transaction
				rabbit_mq.SendMessage("transaction", message)
				if err != nil {
					log.Println(err)
					w.WriteHeader(404)
					w.Write([]byte(err.Error()))
				}
				w.WriteHeader(200)
				w.Write([]byte("Transaction complete "))
			} else {
				w.WriteHeader(404)
				w.Write([]byte("Insufficient balance"))
			}
		} else {
			w.WriteHeader(404)
			w.Write([]byte("Unconfirmed mail"))
		}
	} else {
		w.WriteHeader(404)
	}
}

func getTransactions(address string, page int) ([]Transaction, error) {
	db, err := sql.Open("postgres", os.Getenv("postgres"))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer db.Close()
	sqlStatement := `SELECT amount,commission,date,currency,sender,recipient FROM Transactions WHERE sender=$1 OR recipient=$1 ORDER BY date asc LIMIT 10 OFFSET $2`
	transactions, err := db.Query(sqlStatement, address, page)
	defer transactions.Close()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var currency, sender, recipient string
	var date time.Time
	var amount, commission float64
	var transaction []Transaction
	for transactions.Next() {
		err := transactions.Scan(&amount, &commission, &date, &currency, &sender, &recipient)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		t := Transaction{
			Amount:     amount,
			Currency:   currency,
			Sender:     sender,
			Recipient:  recipient,
			Date:       date,
			Commission: commission,
		}
		transaction = append(transaction, t)
	}
	return transaction, nil
}

func Transactions(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		w.WriteHeader(404)
		return
	}
	jwt := req.Header.Get("jwt")
	var transactionList TransactionList
	json.NewDecoder(req.Body).Decode(&transactionList)

	var email string = getSenderEmail(transactionList.Address)

	if CheckAuth(jwt, email, os.Getenv("jwtauth")) {
		transactions, err := getTransactions(transactionList.Address, transactionList.Page)
		if err != nil {
			log.Println(err)
			w.WriteHeader(404)
			return
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(transactions)
	} else {
		w.WriteHeader(404)
	}

}

func getTransactionsByDate(query_date string, address string, page int) ([]Transaction, error) {
	db, err := sql.Open("postgres", DATABASE_URI)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer db.Close()
	sqlStatement := `SELECT amount,commission,date ,currency,sender,recipient FROM Transactions  t WHERE (sender=$1 OR recipient=$1) AND t.date::date=$2::date ORDER BY date asc LIMIT 10 OFFSET $3`
	transactions, err := db.Query(sqlStatement, address, query_date, page)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer transactions.Close()
	var currency, sender, recipient string
	var date time.Time
	var amount, commission float64
	var transaction []Transaction
	for transactions.Next() {
		err := transactions.Scan(&amount, &commission, &date, &currency, &sender, &recipient)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		t := Transaction{
			Amount:     amount,
			Currency:   currency,
			Sender:     sender,
			Recipient:  recipient,
			Date:       date,
			Commission: commission,
		}
		transaction = append(transaction, t)
	}
	return transaction, nil
}

func SearchTransactionByDate(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		w.WriteHeader(404)
		return
	}
	var transactionList TransactionList
	json.NewDecoder(req.Body).Decode(&transactionList)
	jwt := req.Header.Get("jwt")
	var email string = getSenderEmail(transactionList.Address)
	if CheckAuth(jwt, email, os.Getenv("jwtauth")) {
		transactions, err := getTransactionsByDate(transactionList.Date, transactionList.Address, transactionList.Page)
		if err != nil {
			log.Println(err)
			w.WriteHeader(404)
			return
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(transactions)
	} else {
		w.WriteHeader(404)
	}
}
