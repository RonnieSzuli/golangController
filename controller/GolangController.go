package controller

import (
	"net/http"
	"encoding/json"
	"io/ioutil"
	"strings"
	"strconv"
	"database/sql"
    "fmt"
	_ "github.com/lib/pq"

	"github.com/gorilla/mux"
)

const (
    host     = "localhost"
    port     = 5432
    user     = "postgres"
    password = ""
    dbname   = "postgres"
)

type Student struct {
	Id      int64 `json:"Id"`
	Name 	string `json:"Name"`
}

var Students []Student

func ReturnAllStudents(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(Students)
	db, err := OpenConnection();
	returnAllStmt := `select * from students`
	_, e := db.Exec(returnAllStmt)
	HandleTransaction(db, err, e)
}

func ReturnSingleStudent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["id"]
	for _, student := range Students {
		if strconv.FormatInt(int64(student.Id), 10) == key {
			json.NewEncoder(w).Encode(student)
			db, err := OpenConnection();
			returnStmt := `select * from students where id = $1`
			_, e := db.Exec(returnStmt, student.Id)
			HandleTransaction(db, err, e)
		}
	}
}

func CreateNewStudent(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var student Student
	json.Unmarshal(reqBody, &student)
	Students = append(Students, student)

	json.NewEncoder(w).Encode(student)
	db, err := OpenConnection();
	insertStmt := `insert into students ("id", "name") values($1, $2)`
	_, e := db.Exec(insertStmt, student.Id, student.Name)
	HandleTransaction(db, err, e)
}

func DeleteStudent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	for index, student := range Students {
		if strconv.FormatInt(int64(student.Id), 10) == id {
			Students = append(Students[:index], Students[index+1:]...)
			db, err := OpenConnection();
			deleteStmt := `delete from students where id = $1`
			_, e := db.Exec(deleteStmt, student.Id)
			HandleTransaction(db, err, e)
		}
	}
}

func UpdateStudent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 0, 64)
	CheckError(err)
	reqBody, _ := ioutil.ReadAll(r.Body)
	for index, student := range Students {
		if student.Id == id {
			var newStudents Student
			json.Unmarshal(reqBody, &newStudents)
			json.NewEncoder(w).Encode(newStudents)
			var oldStudent Student
			oldStudent.Id = id
			oldStudent.Name = student.Name
			if len(strings.TrimSpace(newStudents.Name)) > 0 {
				oldStudent.Name = newStudents.Name
			}
			Students = append(Students[:index], oldStudent)
			db, err := OpenConnection();
			updateStmt := `update students set name = $1 where id = $2`
			_, e := db.Exec(updateStmt, oldStudent.Name, oldStudent.Id)
			HandleTransaction(db, err, e)
			break
		}
	}
}

func CheckError(err error) {
    if err != nil {
        panic(err)
    }
}

func OpenConnection() (*sql.DB, error) {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlconn)
	return db, err
}

func HandleTransaction(db *sql.DB, err error, e error) {
    CheckError(err)
	defer db.Close()
	CheckError(e)
}