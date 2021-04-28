package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"database/sql"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	Email    string    `json:"email"`
	Password string    `json:"password"`
	Date     time.Time `json:"date"`
	Attemt   int       `json:"attempt"`
	Active   time.Time `json:"active"` //time.Now().Add(3 * time.Minute)
	Jwt      string    `json:"jwt"`
}

func CreateJWT(email string, signingKey string) (string, error) {
	jwtStr := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour * 168)
	claims["email"] = email
	jwtStr.Claims = claims
	return jwtStr.SignedString([]byte(signingKey))
}

func checkJwt(jwt string) (string, error) {
	db, err := sql.Open("postgres", os.Getenv("postgres"))
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer db.Close()
	var email string
	sqlStatement := `SELECT email FROM Auths WHERE jwt=$1`
	err = db.QueryRow(sqlStatement, jwt).Scan(&email)
	if err != nil {
		log.Println(err)
		return "", err
	}
	return email, nil
}

func CheckAuth(jwtString string, email string, signingKey string) bool {
	db_email, err := checkJwt(jwtString)
	if err != nil || db_email != email {
		return false
	}
	claims := jwt.MapClaims{}
	claims["email"] = email
	claims["exp"] = time.Now().Add(time.Hour * 168)
	parsedJWT, err := jwt.ParseWithClaims(jwtString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Signing method error: %v", token.Header["alg"])
		}
		if clms, ok := token.Claims.(jwt.MapClaims); ok && clms["email"] == email {
			return []byte(signingKey), nil
		}
		return errors.New("Invalid jwt"), nil
	})
	return err == nil && parsedJWT.Valid
}

func getAuthAttempts(email string) int {
	db, err := sql.Open("postgres", os.Getenv("postgres"))
	if err != nil {
		log.Println(err)
		return -1
	}
	defer db.Close()
	var attempt int
	sqlStatement := `SELECT attempt FROM Auths WHERE email=$1`
	err = db.QueryRow(sqlStatement, email).Scan(&attempt)
	if err != nil {
		log.Println(err)
		return -1
	}
	return attempt
}

func updateAuthAttempts(auth Auth) error {
	db, err := sql.Open("postgres", os.Getenv("postgres"))
	if err != nil {
		log.Println(err)
		return err
	}
	defer db.Close()
	var sqlStatement string
	if (auth.Active != time.Time{}) {
		sqlStatement = `UPDATE Auths SET active=$1, attempt=$2 WHERE email=$3`
		_, err = db.Query(sqlStatement, auth.Active, auth.Attemt, auth.Email)
	} else {
		sqlStatement = `UPDATE Auths SET attempt=$1 WHERE email=$2`
		_, err = db.Query(sqlStatement, auth.Attemt, auth.Email)
	}
	if err != nil {
		return err
	}
	return nil
}

func updateAuth(auth Auth) error {
	db, err := sql.Open("postgres", os.Getenv("postgres"))
	if err != nil {
		log.Println(err)
		return err
	}
	defer db.Close()
	sqlStatement := `UPDATE Auths SET date=$2,attempt=$3,active=$4,jwt=$5 WHERE email=$1`
	_, err = db.Query(sqlStatement, auth.Email, auth.Date, auth.Attemt, auth.Active, auth.Jwt)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func getPassword(email string) string {
	db, err := sql.Open("postgres", os.Getenv("postgres"))
	if err != nil {
		log.Println(err)
		return err.Error()
	}
	defer db.Close()
	sqlStatement := `SELECT password FROM Users WHERE email=$1`
	var password string
	err = db.QueryRow(sqlStatement, email).Scan(&password)
	if err != nil {
		return ""
	}
	return password
}

func isDueTime(cur_time time.Time, active time.Time) bool {
	_, seconds := cur_time.Zone()
	location, _ := time.LoadLocation("UTC")
	cur_time = cur_time.In(location).Add(time.Second * time.Duration(seconds))
	active = active.In(location)
	return active.Unix() < cur_time.Unix()
}

func checkLoginTime(email string) error {
	db, err := sql.Open("postgres", os.Getenv("postgres"))
	if err != nil {
		log.Println(err)
		return err
	}
	defer db.Close()
	var active time.Time
	sqlStatement := `SELECT active FROM Auths WHERE email=$1`
	err = db.QueryRow(sqlStatement, email).Scan(&active)
	if err != nil {
		log.Println(err)
		return err
	}
	cur_time := time.Now()
	check := isDueTime(cur_time, active)
	if !check {
		return errors.New("Login not possible wait for a few minutes")
	}
	return nil
}

func Login(w http.ResponseWriter, req *http.Request) {
	if !(req.Method == "POST") {
		w.WriteHeader(404)
		return
	}
	var auth Auth
	err := json.NewDecoder(req.Body).Decode(&auth)
	if err != nil {
		log.Println(err)
		w.WriteHeader(404)
		return
	}
	if !CheckMail(auth.Email) {
		w.WriteHeader(404)
		w.Write([]byte("Email domain is not real"))
		return
	}

	hashedPassword := getPassword(auth.Email)
	if hashedPassword == "" {
		log.Println("Email not registered")
		w.WriteHeader(404)
		w.Write([]byte("Email not registered"))
		return
	}
	if err := checkLoginTime(auth.Email); err != nil {
		w.WriteHeader(404)
		w.Write([]byte(err.Error()))
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(auth.Password))
	if err != nil {
		log.Println(err)
		w.WriteHeader(404)
		w.Write([]byte("Password is wrong"))
		attempts := getAuthAttempts(auth.Email)
		if attempts == -1 {
			return
		}
		if attempts >= 2 {
			auth.Attemt = 0
			auth.Active = time.Now().Add(time.Minute * 3)
		} else {
			auth.Attemt = attempts + 1
		}
		err = updateAuthAttempts(auth)
		if err != nil {
			log.Println(err)
			return
		}

		return
	}

	jwt, err := CreateJWT(auth.Email, os.Getenv("jwtauth"))
	auth.Attemt = 0
	auth.Jwt = jwt
	auth.Date = time.Now()
	auth.Active = time.Now()
	err = updateAuth(auth)
	if err != nil {
		log.Println(err)
		w.WriteHeader(404)
		w.Write([]byte("Auth error : " + err.Error()))
		return
	}
	w.WriteHeader(200)
	w.Write([]byte("new login, jwt=" + jwt))
}
