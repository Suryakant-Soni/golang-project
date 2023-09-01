package mdb

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/mattn/go-sqlite3"
)

type EmailEntry struct {
	Id          int64
	Email       string
	ConfirmedAt *time.Time
	OptOut      bool
}

func TryCreate(db *sql.DB) {
	_, err := db.Exec(`
	CREATE TABLE emails (
		id 			 INTEGER_PRIMARY_KEY,
		email 		 TEXT_UNIQUE,
		confirmed_at INTEGER,
		opt_out 	 INTEGER
	);
	`)
	if err != nil {
		if sqlError, ok := err.(sqlite3.Error); ok {
			if sqlError.Code != 1 {
				log.Fatal(sqlError)
			}
		} else {
			log.Fatal(err)
		}
	}
}
func emailEntryFromRow(row *sql.Rows) (*EmailEntry, error) {
	var id int64
	var email string
	var ConfirmedAt int64
	var optOut bool

	err := row.Scan(&id, &email, &ConfirmedAt, &optOut)
	if err != nil {
		fmt.Println("Error", err)
		return nil, err
	}
	t := time.Unix(ConfirmedAt, 0)
	return &EmailEntry{Id: id, Email: email, ConfirmedAt: &t, OptOut: optOut}, nil

}
func CreateEmail(db *sql.DB, email string) error {
	_, err := db.Exec(`
	INSERT INTO emails(email, confirmed_at,opt_out
		) VALUES(?,0,false)`, email)
	if err != nil {
		fmt.Println("Error", err)
		return err
	}
	return nil
}
func GetEmail(db *sql.DB, email string) (*EmailEntry, error) {
	rows, err := db.Query(`
SELECT id, email, confirmed_at,opt_out FROM emails
WHERE email = ?
`, email)
	if err != nil {
		fmt.Println("Error", err)
		return nil, err
	}
	// when using db.query db keeps the coannection open which has to be shut as below
	defer rows.Close()

	for rows.Next() {
		return emailEntryFromRow(rows)
	}
	return nil, nil
}

func UpdateEmail(db *sql.DB, entry EmailEntry) error {
	// as db saves time in unix here
	t := entry.ConfirmedAt.Unix()

	_, err := db.Exec(`
	INSERT INTO emails(email, confirmed_at, opt_out)
	VALUES(?,?,?)
	ON CONFLICT(email) DO UPDATE SET
	confirmed_at=?,
	opt_out=?`, entry.Email, t, entry.OptOut, t, entry.OptOut)
	if err != nil {
		fmt.Println("Error", err)
		return err
	}
	return nil
}

func DeleteEmail(db *sql.DB, email string) error {
	_, err := db.Exec(`
UPDATE emails
SET opt_out=true
WHERE email=?`, email)
	if err != nil {
		fmt.Println("Error", err)
		return err
	}
	return nil
}

type GetEmailBatchQueryParams struct {
	Page  int
	Count int
}

func GetEmailBatch(db *sql.DB, params GetEmailBatchQueryParams) ([]EmailEntry, error) {
	var empty []EmailEntry

	rows, err := db.Query(`
SELECT id, email, confirmed_at,opt_out
FROM emails
WHERE opt_out=false
ORDER BY id ASC
LIMIT ? OFFSET ?
`, params.Count, (params.Page-1)*params.Count)
	if err != nil {
		fmt.Println("Error", err)
		return empty, err
	}
	defer rows.Close()
	emails := make([]EmailEntry, 0, params.Count)
	for rows.Next() {
		email, err := emailEntryFromRow(rows)
		if err != nil {
			fmt.Println("Error", err)
			return nil, err
		}
		emails = append(emails, *email)
	}
	return emails, nil
}
