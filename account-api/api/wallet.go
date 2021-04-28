package api

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

type Wallet struct {
	Email    string  `json:"email"`
	Address  string  `json:"address"`
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
}

func createAddress() string {
	s := make([]byte, 32)
	_, err := rand.Read(s)
	if err != nil {
		log.Println(err)
	}
	hash := hex.EncodeToString(s)
	return hash
}

func CreateWallet(email string, currency string) error {
	address := createAddress()
	db, err := sql.Open("postgres", os.Getenv("postgres"))
	if err != nil {
		log.Println(err)
		return err
	}
	defer db.Close()
	sqlStatement := `INSERT INTO Wallets(address,email,currency,balance) VALUES($1,$2,$3,$4)`
	_, err = db.Query(sqlStatement, address, email, currency, 100.0)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func getWallets(email string) ([]Wallet, error) {
	db, err := sql.Open("postgres", os.Getenv("postgres"))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer db.Close()
	sqlStatement := `SELECT address,currency,balance FROM Wallets WHERE email=$1`
	wallets, err := db.Query(sqlStatement, email)
	defer wallets.Close()
	var address, currency string
	var balance float64
	var wallet []Wallet
	for wallets.Next() {
		err := wallets.Scan(&address, &currency, &balance)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		w := Wallet{
			Address:  address,
			Currency: currency,
			Balance:  balance,
			Email:    email,
		}
		wallet = append(wallet, w)
	}
	return wallet, nil
}

func GetBalance(address string) (float64, error) {
	db, err := sql.Open("postgres", os.Getenv("postgres"))
	if err != nil {
		log.Println(err)
		return 0, err
	}
	defer db.Close()
	sqlStatement := `SELECT balance FROM Wallets WHERE address=$1`
	var balance float64
	err = db.QueryRow(sqlStatement, address).Scan(&balance)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return balance, nil
}

func UpdateBalance(address string, balance float64, currency string) error {
	db, err := sql.Open("postgres", os.Getenv("postgres"))
	if err != nil {
		log.Println(err)
		return err
	}
	defer db.Close()
	var db_currency string
	sqlStatement := `SELECT currency FROM Wallets WHERE address=$1`
	err = db.QueryRow(sqlStatement, address).Scan(&db_currency)
	if err != nil {
		log.Println(err)
		return err
	}
	if currency != db_currency {
		log.Println("different currency ", currency, " ,, ", db_currency)
		return errors.New("Different currency")
	}
	sqlStatement = "UPDATE Wallets SET balance=$1 WHERE address=$2"
	_, err = db.Query(sqlStatement, balance, address)
	if err != nil {
		log.Println("balance error : ", err)
		return err
	}
	return nil
}

func Wallets(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		w.WriteHeader(404)
		return
	}
	var auth Auth
	json.NewDecoder(req.Body).Decode(&auth)
	jwt := req.Header.Get("jwt")
	if CheckAuth(jwt, auth.Email, os.Getenv("jwtauth")) {
		wallets, err := getWallets(auth.Email)
		if err != nil {
			w.WriteHeader(404)
			w.Write([]byte(err.Error()))
			return
		}
		getWallets(auth.Email)
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(wallets)
	} else {
		w.WriteHeader(404)
	}
}
