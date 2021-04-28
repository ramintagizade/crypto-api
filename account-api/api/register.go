package api

import (
	"account-api/rabbit_mq"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
	"unicode"

	"github.com/badoux/checkmail"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	Firstname  string `json:"firstname"`
	Lastname   string `json:"lastname"`
	Ip         string `json:"ip"`
	User_Agent string `json:"user_agent"`
	Role       string `json:"role"`
}

type Message struct {
	Title string `json:"title"`
	Email string `json:"email"`
	Link  string `json:"link"`
	Jwt   string `json:"jwt"`
}

var DATABASE_URI = os.Getenv("postgres")

func getIp(req *http.Request) string {
	ip_address := req.Header.Get("X-Real-Ip")
	if ip_address == "" {
		ip_address = req.Header.Get("X-Forwarded-For")
	}
	if ip_address == "" {
		ip_address = req.RemoteAddr
	}
	return ip_address
}

func isLetter(s string) bool {
	for _, i := range s {
		if !unicode.IsLetter(i) {
			return false
		}
	}
	return true
}

func generatePassword(p string) []byte {
	password := []byte(p)
	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	return hashedPassword
}

func CheckMail(email string) bool {
	err := checkmail.ValidateHost(email)
	return err == nil
}

func registerUser(user User) error {

	db, err := sql.Open("postgres", os.Getenv("postgres"))
	if err != nil {
		log.Println(err)
		return err
	}
	defer db.Close()
	sqlStatement := `INSERT INTO Users (email,password,firstname,lastname,ip,user_agent,role) VALUES ($1,$2,$3,$4,$5,$6,$7)`
	_, err = db.Query(sqlStatement, user.Email, user.Password, user.Firstname,
		user.Lastname, user.Ip, user.User_Agent, user.Role)

	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func createAuth(auth Auth) error {
	db, err := sql.Open("postgres", os.Getenv("postgres"))
	if err != nil {
		log.Println(err)
		return err
	}
	defer db.Close()
	sqlStatement := `INSERT INTO Auths(email,date,attempt,active,jwt) VALUES($1,$2,$3,$4,$5)`
	_, err = db.Query(sqlStatement, auth.Email, auth.Date, auth.Attemt, auth.Active, auth.Jwt)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func Register(w http.ResponseWriter, req *http.Request) {

	if !(req.Method == "POST") {
		w.WriteHeader(404)
		return
	}

	var user User
	err := json.NewDecoder(req.Body).Decode(&user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(404)
		w.Write([]byte(err.Error()))
		return
	}

	if len(user.Password) < 8 || len(user.Password) > 50 {
		w.WriteHeader(404)
		w.Write([]byte("Password length should be between 8 to 50 characters\n"))
		return
	}
	if !isLetter(user.Firstname) {
		w.WriteHeader(404)
		w.Write([]byte("Firstname should consist of characters\n"))
		return
	}
	if !isLetter(user.Lastname) {
		w.WriteHeader(404)
		w.Write([]byte("Lastname should consist of characters\n"))
		return
	}
	if !CheckMail(user.Email) {
		w.WriteHeader(404)
		w.Write([]byte("Email domain is not real"))
		return
	}

	user.Ip = getIp(req)
	user.User_Agent = req.UserAgent()
	user.Role = "User"
	hashedPassword := generatePassword(user.Password)
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(user.Password))
	if err != nil {
		log.Println(err)
	}
	user.Password = string(hashedPassword)

	err = registerUser(user)
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte("Unable to register : " + err.Error()))
		return
	}

	err = CreateWallet(user.Email, "btc")
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte("Unable to create wallet btc: " + err.Error()))
		return
	}
	err = CreateWallet(user.Email, "eth")
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte("Unable to create wallet eth: " + err.Error()))
		return
	}

	jwt, err := CreateJWT(user.Email, os.Getenv("jwtauth"))
	var auth Auth
	auth.Email = user.Email
	auth.Attemt = 0
	auth.Date = time.Now()
	auth.Active = time.Now()
	auth.Jwt = jwt
	err = createAuth(auth)
	if err != nil {
		log.Println(err)
		w.WriteHeader(404)
		w.Write([]byte(err.Error()))
		return
	}
	var mail Mail
	mail.Email = user.Email
	mail.Confirm = false
	mail.Link = GenerateLink()
	err = registerMail(mail)
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte(err.Error()))
		return
	}

	var message Message
	message.Email = mail.Email
	message.Link = "http://" + req.Host + "/mail?confirm=" + mail.Link
	message.Jwt = jwt
	message.Title = "Confirmation mail"
	err = rabbit_mq.SendMessage("register", message)
	if err != nil {
		log.Println(err)
		w.WriteHeader(404)
		w.Write([]byte(err.Error()))
	}
	w.WriteHeader(200)
	w.Write([]byte("Registration complete, jwt=" + jwt))
}
