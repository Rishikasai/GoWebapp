package main

import (
	"bankdb/models"
	"html/template"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username, Password string
	Amount             int
}
type Transactions struct {
	Username, TransactionType, Sender, Receiver string
	Amount                                      int
}

var templates *template.Template
var store = sessions.NewCookieStore([]byte("rishika"))

//var db *sql.DB
var err error

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example

	file, err := os.OpenFile("logrus.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(file)
	} else {
		log.Info("Failed to log to file, using default stderr")
	}

}

var db = models.Init()

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
	r.HandleFunc("/transfer", TransferGetHandler).Methods("GET")
	r.HandleFunc("/transfer", TransferPostHandler).Methods("POST")
	r.HandleFunc("/checkbal", BalanceGetHandler).Methods("GET")
	r.HandleFunc("/transactions", TransactionGetHandler).Methods("GET")
	r.HandleFunc("/index", IndexGetHandler).Methods("GET")
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)

}

func homeGetHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "home.html", nil)
}
func logoutGetHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	a := session.Values["username"]
	session.Values["username"] = nil
	session.Save(r, w)
	http.Redirect(w, r, "/", 302)
	log.WithFields(log.Fields{
		"event": "logout",
		"user":  a,
	}).Info("User loggedout")
}

func loginGetHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "login.html", nil)
}

func loginPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

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
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password2))

	if err != nil {
		templates.ExecuteTemplate(w, "login.html", "invalid login")
		log.WithFields(log.Fields{
			"event": "logIn",
			"user":  username2,
		}).Warn("Login failed. Username and Password doesn't match")
		return
	}
	if err == nil {
		session, _ := store.Get(r, "session")
		session.Values["username"] = username2
		session.Save(r, w)
		log.WithFields(log.Fields{
			"event": "logIn",
			"user":  username2,
		}).Info("User loggedIn")

		http.Redirect(w, r, "/index", 302)

	}
	templates.ExecuteTemplate(w, "login", nil)
}

func registerGetHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "register.html", nil)
}

func registerPostHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")

	Result2, err := db.Query("SELECT * FROM users WHERE username=?", username)
	if Result2.Next() != false {
		w.Write([]byte("User already exists"))
		return
	}

	cost := bcrypt.DefaultCost
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return
	}
	amount := 0

	_, err = db.Exec(`INSERT INTO users (username, password, amount) VALUES (?, ?, ?)`, username, hash, amount)
	if err != nil {
		log.Fatal(err)
	}

	log.WithFields(log.Fields{
		"event": "Register",
		"user":  username,
	}).Info("New User got Registered")

	http.Redirect(w, r, "/login", 301)
}

func DepositGetHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "deposit.html", nil)
}

func DepositPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

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
		amon := amount2
		amount2 = amount2 + user.Amount
		result3, err := db.Prepare("UPDATE users SET amount=? WHERE username=?")
		if err != nil {
			log.Fatal(err)
			http.Redirect(w, r, "/index", 302)
		}
		_, err = db.Exec(`INSERT INTO transactions (username, amount, type, sender, receiver) VALUES (?, ?, ?, ?, ?)`, username2, amon, "credited", "By deposit", username2)
		if err != nil {
			log.Fatal(err)
		}
		result3.Exec(amount2, username2)
		log.WithFields(log.Fields{
			"event":  "Deposit",
			"user":   username2,
			"amount": amon,
		}).Info("Amount Deposited")
	}
	http.Redirect(w, r, "/index", 302)

}

func WithdrawGetHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "withdraw.html", nil)
}
func WithdrawPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

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
			log.WithFields(log.Fields{
				"event":  "Withdraw",
				"user":   username2,
				"amount": amount2,
			}).Warn("Attempted to withdraw more than the balance")
			return
		}
		if user.Amount >= amount2 {
			amo := amount2
			amount2 = user.Amount - amount2
			result3, err := db.Prepare("UPDATE users SET amount=? WHERE username=?")
			if err != nil {
				log.Fatal(err)
				http.Redirect(w, r, "/index", 302)
			}
			_, err = db.Exec(`INSERT INTO transactions (username, amount, type, sender, receiver) VALUES (?, ?, ?, ?, ?)`, username2, amo, "debited", username2, "withdraw")
			if err != nil {
				log.Fatal(err)
			}
			result3.Exec(amount2, username2)
			log.WithFields(log.Fields{
				"event":  "Withdraw",
				"user":   username2,
				"amount": amo,
			}).Info("Amount Withdrawed")
		}
	}
	http.Redirect(w, r, "/index", 302)

}
func BalanceGetHandler(w http.ResponseWriter, r *http.Request) {

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
		log.WithFields(log.Fields{
			"event": "Balance Check",
			"user":  username,
		}).Info("User checked Balance")
		templates.ExecuteTemplate(w, "balance.html", amount4)
	}

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

func TransferGetHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "transfer.html", nil)
}
func TransferPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	receiver := r.PostForm.Get("acc")
	password := r.PostForm.Get("password")
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
		user.Password = password3
		amount2, err := strconv.Atoi(amount)
		if err != nil {
			log.Fatal(err)
		}
		balance := user.Amount - amount2
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

		if err != nil {
			w.Write([]byte("Transfer failed. wrong password"))
			log.WithFields(log.Fields{
				"event": "logIn",
				"user":  username2,
			}).Warn("Transfer failed. Wrong Password")
			return
		}

		if err == nil {
			if user.Amount < amount2 {
				w.Write([]byte("Insufficient amount! Please check balance"))
				log.WithFields(log.Fields{
					"event":  "Withdraw",
					"user":   username2,
					"amount": amount2,
				}).Warn("Attempted to withdraw more than the balance")
				return
			}
			if user.Amount >= amount2 {

				Result2, err := db.Query("SELECT * FROM users WHERE username=?", receiver)
				if Result2.Next() == false {
					w.Write([]byte("User doesn't exists"))
					return
				}
				if err != nil {
					log.Fatal(err)
				}

				user := User{}

				for Result2.Next() {
					var username4, password4 string
					var amount4 int
					err = Result2.Scan(&username4, &password4, &amount4)

					if err != nil {
						panic(err.Error())
					}
					user.Amount = amount4
					user.Password = password4
				}
				recAmount := user.Amount + amount2
				amo := amount2
				result3, err := db.Prepare("UPDATE users SET amount=? WHERE username=?")
				if err != nil {
					log.Fatal(err)
					http.Redirect(w, r, "/index", 302)
				}

				result4, err := db.Prepare("UPDATE users SET amount=? WHERE username=?")
				if err != nil {
					log.Fatal(err)
					http.Redirect(w, r, "/index", 302)
				}

				if err == nil {
					result3.Exec(balance, username2)
					_, err = db.Exec(`INSERT INTO transactions (username, amount, type, sender, receiver) VALUES (?, ?, ?, ?, ?)`, username2, amo, "debited", username2, receiver)
					if err != nil {
						log.Fatal(err)
					}
					result4.Exec(recAmount, receiver)
					_, err = db.Exec(`INSERT INTO transactions (username, amount, type, sender, receiver) VALUES (?, ?, ?, ?, ?)`, receiver, amo, "credited", username2, receiver)
					if err != nil {
						log.Fatal(err)

					}

					log.WithFields(log.Fields{
						"event":    "Transfer",
						"sender":   username2,
						"receiver": receiver,
						"amount":   amo,
					}).Info("Amount Transferred")
				}
			}

			http.Redirect(w, r, "/index", 302)

		}
	}
}

func TransactionGetHandler(w http.ResponseWriter, r *http.Request) {
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
	Result, err := db.Query("SELECT * FROM transactions WHERE username=?", username)
	transaction := Transactions{}
	res := []Transactions{}
	for Result.Next() {
		var username3, transactionType, sender, receiver string
		var amount int
		err = Result.Scan(&username3, &amount, &transactionType, &sender, &receiver)
		if err != nil {
			panic(err.Error())
		}
		transaction.Amount = amount
		transaction.TransactionType = transactionType
		transaction.Sender = sender
		transaction.Receiver = receiver
		res = append(res, transaction)

	}
	log.WithFields(log.Fields{
		"event": "Transactions Check",
		"user":  username,
	}).Info("User checked Transactions")

	templates.ExecuteTemplate(w, "transaction.html", res)

}


/* 
Query:-:MySQL
CREATE DATABASE dbname;
USE dbname;
CREATE TABLE users(
username TEXT,
password BINARY(60),
amount INT);

CREATE TABLE transactions(
username TEXT,
amount INT,
type TEXT,
sender TEXT,
receiver TEXT);
*/

