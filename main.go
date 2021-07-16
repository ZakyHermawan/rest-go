package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/darahayes/go-boom"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)


type Book struct {
	Id int `json:"id"`
	Harga int `json:"harga"`
	Judul string `json:"judul"`
	Pengarang string `json:"pengarang"`
	Penerbit string `json:"penerbit"`
}


type tmpBook struct {
	Harga int `json:"harga"`
	Judul string `json:"judul"`
	Pengarang string `json:"pengarang"`
	Penerbit string `json:"penerbit"`
}


func homePage(w http.ResponseWriter, _* http.Request) {
	fmt.Println("Endpoint hit: Homepage")
	_, err := fmt.Fprintf(w, "Welcome to Homepage")
	checkErr(err)

}

func returnAll(w http.ResponseWriter, _* http.Request) {
	fmt.Println("Endpoint hit: All Book")
	err := json.NewEncoder(w).Encode(Books)
	checkErr(err)
}

func returnSingle(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint hit: Return single book")
	vars := mux.Vars(r)

	for _, book := range Books {
		id, _ := strconv.Atoi(vars["id"])
		if book.Id == id {
			err := json.NewEncoder(w).Encode(book)
			checkErr(err)
		}
	}
}

func createNew(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint hit: Create new")
	reqBody, _ := ioutil.ReadAll(r.Body)

	var tmp tmpBook
	err := json.Unmarshal(reqBody, &tmp)
	checkErr(err)
	judul := tmp.Judul

	for _, val := range Books {
		if val.Judul == judul {
			err = errors.New("buku sudah ada")
			boom.NotAcceptable(w, err)
			return
		}
	}

	book := Book{lastId+1, tmp.Harga, tmp.Judul, tmp.Pengarang, tmp.Penerbit}
	Books = append(Books, book)

	insertNewData(&book)
	lastId++

	err = json.NewEncoder(w).Encode(Books)
	checkErr(err)
}

func deleteOne(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint hit: Delete one")
	vars := mux.Vars(r)

	for index, book := range Books {
		id, _ := strconv.Atoi(vars["id"])
		if book.Id == id {
			Books = append(Books[:index], Books[index+1:]...)
			deleteData(&book)
		}
	}

	err:= json.NewEncoder(w).Encode(Books)
	checkErr(err)
}

func updateOne(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint hit: Update one")
	reqBody, _ := ioutil.ReadAll(r.Body)

	var newBook Book
	err := json.Unmarshal(reqBody, &newBook)
	checkErr(err)

	judul := newBook.Judul

	for index, book := range Books {
		if book.Judul == judul {
			oldId := Books[index].Id
			Books[index] = newBook
			Books[index].Id = oldId
			updateData(&newBook)
		}
	}

	err = json.NewEncoder(w).Encode(Books)
	checkErr(err)

}


func handler() {
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/book", returnAll).Methods("GET")
	myRouter.HandleFunc("/book", createNew).Methods("POST")
	myRouter.HandleFunc("/book", updateOne).Methods("PUT")
	myRouter.HandleFunc("/book/{id}", returnSingle)
	myRouter.HandleFunc("/book/delete/{id}", deleteOne)

	log.Fatal(http.ListenAndServe(":8000", myRouter))
}


func checkErr(e error) {
	if e != nil {
		fmt.Println(e)
		panic(e)
	}
}

func insertNewData(book *Book) {
	db, err := sql.Open("sqlite3", "database.db")
	checkErr(err)


	stmt, err := db.Prepare(`INSERT INTO buku
		(id, harga, judul, pengarang, penerbit)
		VALUES (?, ?, ?, ?, ?)`,
	)
	checkErr(err)

	_, err = stmt.Exec(lastId+1, book.Harga, book.Judul, book.Pengarang, book.Penerbit)
	checkErr(err)
}

func updateData(newBook *Book) {
	db, err := sql.Open("sqlite3", "database.db")
	checkErr(err)
	
	stmt, err := db.Prepare(`UPDATE buku
		SET harga = ?,
		    pengarang = ?,
		    penerbit = ?
		WHERE judul = ?`,
	)
	checkErr(err)

	_, err = stmt.Exec(newBook.Harga, newBook.Pengarang, newBook.Penerbit, newBook.Judul)
	checkErr(err)
	err = stmt.Close()
	checkErr(err)

}

func deleteData(book *Book) {
	db, err := sql.Open("sqlite3", "database.db")
	checkErr(err)

	stmt, err := db.Prepare("DELETE FROM buku WHERE id=?")
	checkErr(err)

	_, err = stmt.Exec(book.Id)
	checkErr(err)
}

var (
	Books []Book
	lastId, banyakData int
)


func main() {
	fmt.Println("Mulai")
	db, err := sql.Open("sqlite3", "database.db")

	checkErr(err)

	rows, err := db.Query("SELECT COUNT (*) FROM buku")
	checkErr(err)
	rows.Next()

	err = rows.Scan(&banyakData)
	checkErr(err)
	err = rows.Close()
	checkErr(err)

	rows, err = db.Query("SELECT MAX(ROWID) FROM buku")
	checkErr(err)
	rows.Next()

	err = rows.Scan(&lastId)
	checkErr(err)
	err = rows.Close()
	checkErr(err)

	Books = make([]Book, banyakData)

	rows, err = db.Query("SELECT * FROM buku")
	checkErr(err)

	var (
		id, harga, counter int
		judul, pengarang, penerbit string
	)


	for rows.Next() {
		err := rows.Scan(&id, &harga, &judul, &pengarang, &penerbit)
		checkErr(err)
		Books[counter].Id = id
		Books[counter].Harga = harga
		Books[counter].Judul = judul
		Books[counter].Pengarang = pengarang
		Books[counter].Penerbit = penerbit
		fmt.Println(Books[counter])
		counter++
	}

	err = rows.Close()
	checkErr(err)
	
	handler()
}
