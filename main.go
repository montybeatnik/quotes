package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

type App struct {
	mux *http.ServeMux
	db  *sql.DB
}

type NewQuote struct {
	Message   string    `json:"message,omitempty"`
	Author    string    `json:"author,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type JSONErr struct {
	Err error  `json:"err,omitempty"`
	Msg string `json:"msg,omitempty"`
}

func (a *App) helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "hello, world")
}

func (a *App) newQuote(w http.ResponseWriter, r *http.Request) {
	var nq NewQuote
	if err := json.NewDecoder(r.Body).Decode(&nq); err != nil {
		json.NewEncoder(w).Encode(JSONErr{Err: err})
	}
	log.Printf("%+v", nq)
}

func main() {
	// setup DB
	driver := "postgres"
	dsn := "postgres://postgres:postgres@localhost:5432/quotes?sslmode=disable" // "postgres://YourUserName:YourPassword@YourHostname:5432/YourDatabaseName"
	log.Println("connecting to db...")
	db, err := sql.Open(driver, dsn)
	if err != nil {
		log.Println("failed to connect to DB", err)
		return
	}
	if err := db.Ping(); err != nil {
		log.Println("failed to ping DB", err)
		return
	}
	server := http.Server{Addr: ":8080"}
	app := App{mux: http.NewServeMux(), db: db}
	app.mux.HandleFunc("/", app.helloWorld)
	app.mux.HandleFunc("/new", app.newQuote)
	server.Handler = app.mux
	if err := server.ListenAndServe(); err != nil {
		log.Printf("failed to stand up server: %v", err)
	}
}
