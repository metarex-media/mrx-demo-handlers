package utils

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// generate SQL create the skeleton of the database
func generateSQL(dbName string, overwrite bool) (*sql.DB, error) {

	_, err := os.Open(dbName)

	if !overwrite && err == nil {
		fmt.Printf("Overwriting %s proceed? (y/n) ", dbName)
		input := bufio.NewScanner(os.Stdin)
		input.Scan()
		switch strings.ToLower(input.Text()) {
		case "y", "yes":
		default:
			return nil, fmt.Errorf("database overwrite cancelled, aborting program")
		}
	}

	// @TODO decide to keep this functionality
	os.Remove(dbName) // generate a clean file each time
	// SQLite is a file based database.

	file, err := os.Create(dbName) // Create SQLite file
	if err != nil {
		return nil, err
	}
	file.Close()

	sqliteDatabase, err := sql.Open("sqlite3", dbName) // Open the created SQLite File
	if err != nil {
		return nil, err
	}

	err = createTableNew(sqliteDatabase)
	if err != nil {
		return nil, err
	}

	return sqliteDatabase, nil
}

// createTableNew sets the metarex sql template
func createTableNew(db *sql.DB) error {
	createMetadataTableSQL := `CREATE TABLE log (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"time" TEXT,
		"msg" TEXT,
		"MRXPath" BLOB,
		"level" TEXT,
		"chainID" TEXT,
		"parentID" TEXT,
		"source" BLOB,
		"other" BLOB
	  );` // SQL Statement for Create Table

	statement, err := db.Prepare(createMetadataTableSQL) // Prepare SQL Statement
	if err != nil {
		return err
	}
	_, err = statement.Exec() // Execute SQL Statements

	return err
}

type logContents struct {
	time, msg, level, chainID, parentID string
	mrx, other, source                  []byte
}

// insert MetaData inserts the log as a flat layout
func insertLog(db *sql.DB, lc logContents) error { //Student) {
	// log.Println("Inserting student record ...")

	insertStudentSQL := `INSERT INTO log(time, msg, MRXPath, level, chainID, parentID, source, other) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	statement, err := db.Prepare(insertStudentSQL) // Prepare statement.
	// This is to avoid SQL injections
	if err != nil {
		return err
	}
	_, err = statement.Exec(lc.time, lc.msg, lc.mrx, lc.level, lc.chainID, lc.parentID, lc.source, lc.other)

	return err
}
