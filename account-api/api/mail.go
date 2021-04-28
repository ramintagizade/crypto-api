package api

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

type Mail struct {
	Email   string `json:"email"`
	Confirm bool   `json:"confirm"`
	Link    string `json:"link"`
}

func GenerateLink() string {
	s := make([]byte, 64)
	_, err := rand.Read(s)
	if err != nil {
		log.Println(err)
	}
	hash := hex.EncodeToString(s)
	return hash
}

func GetMailStatusByLink(link string) (bool, error) {
	db, err := sql.Open("postgres", os.Getenv("postgres"))
	if err != nil {
		log.Println(err)
		return false, err
	}
	defer db.Close()
	sqlStatement := `SELECT confirmed FROM Mail WHERE link=$1`
	var confirm bool
	err = db.QueryRow(sqlStatement, link).Scan(&confirm)
	if err != nil {
		log.Println(err)
		return false, err
	}
	return confirm, nil
}

func GetMailStatusByEmail(email string) (bool, error) {
	db, err := sql.Open("postgres", os.Getenv("postgres"))
	if err != nil {
		log.Println(err)
		return false, err
	}
	defer db.Close()
	sqlStatement := `SELECT confirmed FROM Mail WHERE email=$1`
	var confirm bool
	err = db.QueryRow(sqlStatement, email).Scan(&confirm)
	if err != nil {
		log.Println(err)
		return false, err
	}
	return confirm, nil
}

func registerMail(mail Mail) error {
	db, err := sql.Open("postgres", os.Getenv("postgres"))
	if err != nil {
		log.Println(err)
		return err
	}
	defer db.Close()
	sqlStatement := `INSERT INTO Mail (email,confirmed,link) VALUES ($1,$2,$3)`
	_, err = db.Query(sqlStatement, mail.Email, mail.Confirm, mail.Link)
	if err != nil {
		return err
	}
	return nil
}

func updateMailStatus(link string) error {
	db, err := sql.Open("postgres", os.Getenv("postgres"))
	if err != nil {
		log.Println(err)
		return err
	}
	defer db.Close()
	sqlStatement := `UPDATE Mail SET confirmed=$1 WHERE link=$2`
	_, err = db.Query(sqlStatement, true, link)
	if err != nil {
		return err
	}
	return nil
}

func MailConfirmation(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		return
	}
	var link string = req.URL.Query().Get("confirm")
	ok, err := GetMailStatusByLink(link)
	if err == nil && !ok {
		err = updateMailStatus(link)
		if err != nil {
			w.WriteHeader(404)
			w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(200)
		w.Write([]byte("Mail confirmed"))
	} else {
		if ok {
			w.WriteHeader(404)
			w.Write([]byte("Already confirmed"))
		} else {
			w.WriteHeader(404)
		}
	}
}
