package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

const _CostOfPassword = 10

func HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var input ToSaveUser
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, `Couldn't read the payload`)
		return
	}

	// create new user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Passcode), _CostOfPassword)
	if err != nil {
		w.WriteHeader(503)
		fmt.Fprintf(w, `Failed to save user`)
		return
	}

	// save user to db
	db := r.Context().Value("_DB").(*sqlx.DB)
	_, err = db.Exec(`insert into "user" (id, name, pinhash) values ($1, $2, $3)`, input.Uid, input.Name, hashedPassword)

	if err != nil {
		log.Fatal(err)
		w.WriteHeader(400)
		fmt.Fprintf(w, `{ "status": "failed", "code": "failed", "message": "Failed to record the database entry" }`)
		return
	}

	w.WriteHeader(201)
	fmt.Fprintf(w, `{ "status": "success", "message": "user successfully created" }`)
}

type ToSaveUser struct {
	Name     string `json:"name"`
	Passcode string `json:"pin"`
	Uid      string `json:"nickname"`
}

func HandlerGetUser(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimSpace(r.URL.Query().Get("username"))

	if username == "" {
		w.WriteHeader(400)
		fmt.Fprintf(w, `{ "code": "bad-input", "message": "Make sure to specify the user", }`)
		return
	}

	db := r.Context().Value("_DB").(*sqlx.DB)
	var user User
	err := db.Get(&user, `select "id", "name" from "user" where id=$1`, username)
	if err != nil {
		w.WriteHeader(404)
		fmt.Fprintf(w, `{ "code": "not-found", "message": "There's no such user. You can create the user", }`)
		return
	}

	resUser, _ := json.Marshal(user)
	fmt.Fprintf(w, "%s", resUser)
}

type User struct {
	Name string `db:"name" json:"name"`
	Uid  string `db:"id" json:"nickname"`
}
