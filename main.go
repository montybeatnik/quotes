package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"bufio"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
)

// Store houses the DB. It serves as a layer of abstraction with methods
// tailored to the quotes logic.
type Store struct {
	db *sql.DB
}

// AddCategory adds a category for quotes. This allows us to group quotes
// into meaningful buckets.
func (s *Store) AddCategory(name string) error {
	res, err := s.db.Exec(`
INSERT INTO categories (
	name,
	created_at
) VALUES ($1, $2);`, name, time.Now())
	if err != nil {
		return fmt.Errorf("failed to add category: %v", err)
	}
	_ = res // might need this later.
	return nil
}

// Categories retrieves all categories from the DB. As this table grows,
// there may be a need to paginate.
func (s *Store) Categories() ([]Category, error) {
	var cs []Category
	query := `
SELECT 
	id,
	name,
	created_at
FROM categories;`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("get all categories query failed: %v", err)
	}
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.CreatedAt); err != nil {
			log.Printf("failed to scan %v\n", err)
		}
		cs = append(cs, c)
	}
	return cs, nil
}

func (s *Store) AddAuthor(name string) error {
	query := `
INSERT INTO authors (
	name, 
	created_at
) VALUES ($1, $2);`
	_, err := s.db.Exec(query, name, time.Now())
	if err != nil {
		log.Println("new author query failed: %v", err)
	}
	return nil
}

// Authors retrieves all authors from the DB. As this table grows,
// there may be a need to paginate.
func (s *Store) Authors() ([]Author, error) {
	var as []Author
	query := `
SELECT 
	id,
	name,
	created_at
FROM authors;`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("get all authors query failed: %v", err)
	}
	for rows.Next() {
		var a Author 
		if err := rows.Scan(&a.ID, &a.Name, &a.CreatedAt); err != nil {
			log.Printf("failed to scan %v\n", err)
		}
		as = append(as, a)
	}
	return as, nil
}

// Add Quote
func (s *Store) AddQuote(q Quote) error {
	query := `
INSERT INTO messages (
	category_id, 
	author_id,
	message,
	created_at
) VALUES ($1, $2, $3, $4);`
	_, err := s.db.Exec(query, q.Category.ID, q.Author.ID, q.Message, time.Now())
	if err != nil {
		log.Println("new quote query failed: %v", err)
	}
	return nil
}
// App holds the muxer or (router) for hitting different routes.
// It also houses the store. (Way to state the obvious, huh?)
type App struct {
	mux   *http.ServeMux
	store Store
}

// Category facilatates the categorical nature of quotes.
type Category struct {
	ID        int       `json:"id"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// Author is the person attributed to the quote.
type Author struct {
	ID        int       `json:"id"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// Quote is a container for the category, author, and the message.
type Quote struct {
	Category  Category  `json:"category"`
	Message   string    `json:"message,omitempty"`
	Author    Author    `json:"author,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// JSONErr is a layer of abstraction for producing errors in
// the JSON format.
type JSONErr struct {
	Err error  `json:"err,omitempty"`
	Msg string `json:"msg,omitempty"`
}

// healthCheck lets an API caller know if the server is healthy.
func (a *App) healthCheck(w http.ResponseWriter, r *http.Request) {
	err := a.store.db.Ping()
	if err != nil {
		json.NewEncoder(w).Encode(JSONErr{Err: err})
		return
	}
	json.NewEncoder(w).Encode(JSONErr{Err: nil, Msg: "system is healthy"}) // could be a const
}

// newCategory handler to add new categories.
func (a *App) newCategory(w http.ResponseWriter, r *http.Request) {
	var nc Category
	if err := json.NewDecoder(r.Body).Decode(&nc); err != nil {
		json.NewEncoder(w).Encode(JSONErr{Err: err})
	}
	if err := a.store.AddCategory(nc.Name); err != nil {
		log.Println(err)
		json.NewEncoder(w).Encode(JSONErr{Err: err})
		return
	}
}

// getCategory handler to get categories.
func (a *App) getCategories(w http.ResponseWriter, r *http.Request) {
	cs, err := a.store.Categories()
	if err != nil {
		json.NewEncoder(w).Encode(JSONErr{Err: err})
		return
	}
	json.NewEncoder(w).Encode(cs)
}

func (a *App) newAuthor(w http.ResponseWriter, r *http.Request) {
	var author Author
	if err := json.NewDecoder(r.Body).Decode(&author); err != nil {
		json.NewEncoder(w).Encode(JSONErr{Err: err})
	}
	if err := a.store.AddAuthor(author.Name); err != nil {
		json.NewEncoder(w).Encode(JSONErr{Err: err})
	}
}

func (a *App) Authors(w http.ResponseWriter, r *http.Request) {
	authors, err := a.store.Authors()
	if err != nil {
		json.NewEncoder(w).Encode(JSONErr{Err: err})
	}
	json.NewEncoder(w).Encode(authors)
}                                                               

// newQuote adds a new quote.
func (a *App) newQuote(w http.ResponseWriter, r *http.Request) {
	var nq Quote
	if err := json.NewDecoder(r.Body).Decode(&nq); err != nil {
		json.NewEncoder(w).Encode(JSONErr{Err: err})
	}
	if err := a.store.AddQuote(nq); err != nil {
		json.NewEncoder(w).Encode(JSONErr{Err: err})
	}
}

type Config struct {
	Port int
	DSN string
}

func getEnvVars(fn string) Config {
	var cfg Config
	f, err := os.Open(".env")
	if err != nil {
		log.Println("couldn't open file", err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
		cfg.DSN = strings.Replace(strings.Split(scanner.Text(), "=")[1], "'", "", -1)
	}
	fmt.Printf("%q\n", cfg.DSN)
	return cfg
}

func main() {
	godotenv.Load()
	// setup DB
	driver := "postgres"
	log.Println("connecting to db...")
	db, err := sql.Open(driver, os.Getenv("DSN"))
	if err != nil {
		log.Println("failed to connect to DB", err)
		return
	}
	if err := db.Ping(); err != nil {
		log.Println("failed to ping DB", err)
		return
	}
	server := http.Server{Addr: ":8080"}
	store := Store{db: db}
	app := App{mux: http.NewServeMux(), store: store}
	app.mux.HandleFunc("/health", app.healthCheck)
	app.mux.HandleFunc("/category/new", app.newCategory)
	app.mux.HandleFunc("/author/new", app.newAuthor)
	app.mux.HandleFunc("/author", app.Authors)
	app.mux.HandleFunc("/category", app.getCategories)
	app.mux.HandleFunc("/quote/new", app.newQuote)
	server.Handler = app.mux
	if err := server.ListenAndServe(); err != nil {
		log.Printf("failed to stand up server: %v", err)
	}
}
