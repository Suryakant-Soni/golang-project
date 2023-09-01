package jsonapi

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"errors"

	"mailinglist/mdb"
)

func setJsonHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

// will convert the JSON body received in request to go struct
func fromJson[T any](body io.Reader, target T) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	json.Unmarshal(buf.Bytes(), &target)
}

// this is returning the json only with in the response http
func returnJson[T any](w http.ResponseWriter, withData func() (T, error)) {
	setJsonHeader(w)
	data, serverErr := withData()
	if serverErr != nil {
		w.WriteHeader(500)
		serverErrJson, err := json.Marshal(&serverErr)
		if err != nil {
			fmt.Println("Error", err)
			return
		}
		w.Write(serverErrJson)
		return
	}
	dataJson, err := json.Marshal(&data)
	if err != nil {
		fmt.Println("Error", err)
		return
	}
	w.Write(dataJson)
}

// this functions will use internaly returnjson function and it is returning json to the response
func returnErr(w http.ResponseWriter, err error, code int) {
	returnJson(w, func() (interface{}, error) {
		errorMessage := struct {
			Err string
		}{
			Err: err.Error(),
		}
		w.WriteHeader(code)
		return errorMessage, nil
	})
}
func GetEmailBatch(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			return
		}
		queryOptions := mdb.GetEmailBatchQueryParams{}
		fromJson(req.Body, &queryOptions)

		if queryOptions.Count <= 0 || queryOptions.Page <= 0{
			returnErr(w, errors.New("page and count fields are required and mandatory"),400)
			return
		}
		returnJson(w,func()(interface{},error){
			log.Printf("Json get email batch: %v \n", queryOptions)
			return mdb.GetEmailBatch(db,queryOptions)
		})
	})
}

func GetEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			return
		}
		entry := mdb.EmailEntry{}
		fromJson(req.Body, &entry)

		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON createEmail: %v \n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

func DeleteEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "PUT" {
			return
		}
		entry := mdb.EmailEntry{}
		// need to get data from body which is json and convert in and put it inside address of &entry
		fromJson(req.Body, &entry)

		if err := mdb.DeleteEmail(db, entry.Email); err != nil {
			returnErr(w, err, 400)
			return
		}
		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON updateEmail: %v \n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}
func UpdateEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "PUT" {
			return
		}
		entry := mdb.EmailEntry{}
		fromJson(req.Body, &entry)

		if err := mdb.UpdateEmail(db, entry); err != nil {
			returnErr(w, err, 400)
			return
		}
		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON updateEmail: %v \n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}
func CreateEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			return
		}
		entry := mdb.EmailEntry{}
		fromJson(req.Body, &entry)

		if err := mdb.CreateEmail(db, entry.Email); err != nil {
			returnErr(w, err, 400)
			return
		}
		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON createEmail: %v \n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

func Serve(db *sql.DB, bind string) {
	http.Handle("/email/create", CreateEmail(db))
	http.Handle("/email/get", GetEmail(db))
	http.Handle("/email/get_batch", GetEmailBatch(db))
	http.Handle("/email/update", UpdateEmail(db))
	http.Handle("/email/delete", DeleteEmail(db))
	log.Printf("JSON API server listening on %v \n", bind)
	err := http.ListenAndServe(bind, nil)
	if err != nil{
		fmt.Println("Error",err)
		log.Fatalf("JSON server error: %v",err)
	}
}
