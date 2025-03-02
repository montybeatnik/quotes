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

// NewQuote is a container for the category, author, and the message.
type NewQuote struct {
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

// newQuote adds a new quote.
func (a *App) newQuote(w http.ResponseWriter, r *http.Request) {
	var nq NewQuote
	if err := json.NewDecoder(r.Body).Decode(&nq); err != nil {
		json.NewEncoder(w).Encode(JSONErr{Err: err})
	}
}

func main() {
	// setup DB
	driver := "postgres"
	dsn := "postgres://postgres:postgres@localhost:5432/quotes?sslmode=disable"
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
	store := Store{db: db}
	app := App{mux: http.NewServeMux(), store: store}
	app.mux.HandleFunc("/health", app.healthCheck)
	app.mux.HandleFunc("/category/new", app.newCategory)
	app.mux.HandleFunc("/category", app.getCategories)
	app.mux.HandleFunc("/quote/new", app.newQuote)
	server.Handler = app.mux
	if err := server.ListenAndServe(); err != nil {
		log.Printf("failed to stand up server: %v", err)
	}
}
