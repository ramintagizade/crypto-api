package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"
)

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
