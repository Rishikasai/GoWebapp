package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type User struct {
	Username, Password string
	Amount             int
}

var templates *template.Template
var store = sessions.NewCookieStore([]byte("rishika"))
var db *sql.DB
var err error

func homeGetHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "home.html", nil)
}
func logoutGetHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	session.Values["username"] = nil
	session.Save(r, w)
	http.Redirect(w, r, "/", 302)
}
func loginGetHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "login.html", nil)
}

func loginPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	db, err := sql.Open("mysql", "root:rishika@(127.0.0.1:3306)/dbname")
	username2 := r.PostForm.Get("username")
	password2 := r.PostForm.Get("password")
	Result, err := db.Query("SELECT * FROM users WHERE username=?", username2)
	user := User{}
	for Result.Next() {
		var username3, password3 string
		var amount int
		err = Result.Scan(&username3, &password3, &amount)
		if err != nil {
			panic(err.Error())
		}
		user.Password = password3
	}

	if err != nil {
		panic(err.Error())
	}
	if user.Password != password2 {
		templates.ExecuteTemplate(w, "login.html", "invalid login")
		return
	}

	if user.Password == password2 {
		session, _ := store.Get(r, "session")
		session.Values["username"] = username2
		session.Save(r, w)
		http.Redirect(w, r, "/index", 302)

	}

	templates.ExecuteTemplate(w, "login", nil)
	defer db.Close()
}

func registerGetHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "register.html", nil)
}

func registerPostHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", "root:rishika@(127.0.0.1:3306)/dbname?parseTime=true")
	r.ParseForm()
	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")
	amount := 0
	_, err = db.Exec(`INSERT INTO users (username, password, amount) VALUES (?, ?, ?)`, username, password, amount)
	if err != nil {
		log.Fatal(err)
	}
	http.Redirect(w, r, "/login", 301)
	defer db.Close()
}

func DepositGetHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "deposit.html", nil)
}

func DepositPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	db, err := sql.Open("mysql", "root:rishika@(127.0.0.1:3306)/dbname")
	if err != nil {
		log.Fatal(err)
	}
	session, _ := store.Get(r, "session")
	username2, ok := session.Values["username"].(string)
	if !ok {
		http.Redirect(w, r, "/login", 302)
		return
	}

	amount := r.PostForm.Get("amount")

	Result, err := db.Query("SELECT * FROM users WHERE username=?", username2)
	user := User{}

	for Result.Next() {
		var username3, password3 string
		var amount3 int
		err = Result.Scan(&username3, &password3, &amount3)

		if err != nil {
			panic(err.Error())
		}
		user.Amount = amount3
		amount2, err := strconv.Atoi(amount)
		if err != nil {
			log.Fatal(err)
		}
		amount2 = amount2 + user.Amount
		result3, err := db.Prepare("UPDATE users SET amount=? WHERE username=?")
		if err != nil {
			log.Fatal(err)
			http.Redirect(w, r, "/index", 302)
		}
		result3.Exec(amount2, username2)
	}
	http.Redirect(w, r, "/index", 302)
	defer db.Close()
}

func WithdrawGetHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "withdraw.html", nil)
}
func WithdrawPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	db, err := sql.Open("mysql", "root:rishika@(127.0.0.1:3306)/dbname")
	if err != nil {
		log.Fatal(err)
	}
	session, _ := store.Get(r, "session")
	username2, ok := session.Values["username"].(string)
	if !ok {
		http.Redirect(w, r, "/login", 302)
		return
	}

	amount := r.PostForm.Get("amount")

	Result, err := db.Query("SELECT * FROM users WHERE username=?", username2)
	user := User{}

	for Result.Next() {
		var username3, password3 string
		var amount3 int
		err = Result.Scan(&username3, &password3, &amount3)

		if err != nil {
			panic(err.Error())
		}
		user.Amount = amount3
		amount2, err := strconv.Atoi(amount)
		if err != nil {
			log.Fatal(err)
		}
		if user.Amount < amount2 {
			w.Write([]byte("Insufficient amount! Please check balance"))
			return
		}
		if user.Amount >= amount2 {
			amount2 = user.Amount - amount2
			result3, err := db.Prepare("UPDATE users SET amount=? WHERE username=?")
			if err != nil {
				log.Fatal(err)
				http.Redirect(w, r, "/index", 302)
			}
			result3.Exec(amount2, username2)
		}
	}
	http.Redirect(w, r, "/index", 302)
	defer db.Close()
}
func BalanceGetHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", "root:rishika@(127.0.0.1:3306)/dbname")
	session, _ := store.Get(r, "session")
	untyped, ok := session.Values["username"]
	if !ok {
		http.Redirect(w, r, "/login", 302)
		return
	}
	username, ok := untyped.(string)
	if !ok {
		http.Redirect(w, r, "/login", 302)
		return
	}
	Result, err := db.Query("SELECT * FROM users WHERE username=?", username)
	user := User{}
	for Result.Next() {
		var username3, password3 string
		var amount int
		err = Result.Scan(&username3, &password3, &amount)
		if err != nil {
			panic(err.Error())
		}
		user.Amount = amount
		amount4 := user.Amount
		templates.ExecuteTemplate(w, "balance.html", amount4)
	}
	defer db.Close()
}
func IndexGetHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	username2, ok := session.Values["username"].(string)
	if !ok {
		http.Redirect(w, r, "/login", 302)
		return
	}
	templates.ExecuteTemplate(w, "index.html", username2)
}
func main() {
	templates = template.Must(template.ParseGlob("templates/*.html"))
	r := mux.NewRouter()
	r.HandleFunc("/", homeGetHandler).Methods("GET")
	r.HandleFunc("/login", loginGetHandler).Methods("GET")
	r.HandleFunc("/login", loginPostHandler).Methods("POST")
	r.HandleFunc("/logout", logoutGetHandler).Methods("GET")
	r.HandleFunc("/register", registerGetHandler).Methods("GET")
	r.HandleFunc("/register", registerPostHandler).Methods("POST")
	r.HandleFunc("/deposit", DepositGetHandler).Methods("GET")
	r.HandleFunc("/deposit", DepositPostHandler).Methods("POST")
	r.HandleFunc("/withdraw", WithdrawGetHandler).Methods("GET")
	r.HandleFunc("/withdraw", WithdrawPostHandler).Methods("POST")
	r.HandleFunc("/checkbal", BalanceGetHandler).Methods("GET")
	r.HandleFunc("/index", IndexGetHandler).Methods("GET")
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
