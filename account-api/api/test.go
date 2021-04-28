package api

import (
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "admin"
	password = "password"
	dbname   = "crypto"
)

type Users struct {
	Id       int `json:id`
	Email    string
	Password string
}

func CreateTable() {

	// db, err := sql.Open("postgres", "postgres://admin:password@localhost:5432/crypto?sslmode=disable")
	// if err != nil {
	// 	log.Println(err)
	// }

	// defer db.Close()
	// err = db.Ping()

	// if err != nil {
	// 	log.Println(err)
	// }

	// log.Println("Successfully connected ")
	// sqlStatement := `INSERT INTO users (id, email, password)
	// 	VALUES (2 ,'ramintagizade718@gmail.com', 'pass')`
	// log.Println(sqlStatement)
	// result, err := db.Exec(sqlStatement)
	// if err != nil {
	// 	log.Println(err)
	// }
	// log.Println(result)

	// var my_user Users
	// userSql := "SELECT id, email, password FROM users WHERE id = $1"
	// err = db.QueryRow(userSql, 1).Scan(&my_user.Id, &my_user.Email, &my_user.Password)
	// log.Println(my_user.Email)
}
