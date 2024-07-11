// package utils contains the utilities for mrx logging
package utils

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/lmittmann/tint"
)

// set colour output straight to console
func ColourConsole(Opts *slog.HandlerOptions) {

	// set global logger with custom options
	// assign colours based on Operating system
	colourStart(Opts, false)

}

// set default
// used for writing straight to command line
func Console(Opts *slog.HandlerOptions) {
	colourStart(Opts, true)
}

func colourStart(Opts *slog.HandlerOptions, noColour bool) {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level:       Opts.Level,
			TimeFormat:  time.RFC3339,
			ReplaceAttr: Opts.ReplaceAttr,
			AddSource:   Opts.AddSource,
			NoColor:     noColour,
		}),
	))
}

/*
Need a database name as an option
*/
func JSON(Opts *slog.HandlerOptions, folder string) {

	var mu sync.Mutex
	format := "2006-01-02T15:04:05Z07"
	logName := time.Now().Format(format)
	// generate a new SQL
	path, _ := filepath.Abs(folder + logName + ".db")
	db, _ := generateSQL(path, true)

	sw := sqlWriter{mu: &mu, db: db}
	slog.SetDefault(slog.New(slog.NewJSONHandler(&sw, Opts)))

	/*

		SQL can only do writing to SQL databases.


		Take the text handler split by spaces and work along?

	*/
}

type sqlWriter struct {
	mu *sync.Mutex
	db *sql.DB
	// database
}

const (
	dbTime     = "time"
	dbMSG      = "msg"
	dbMrxPath  = "MRXPath"
	dbLevel    = "level"
	dbSource   = "source"
	dbChainID  = "chainID"
	dbParentID = "parentID"
)

// Write the json output, filter into
// strings and []bytes.
// not functional with any other style of input
func (s sqlWriter) Write(p []byte) (n int, err error) {

	// get the JSON input here and convert it
	// into a flat map
	var in map[string]any
	json.Unmarshal(p, &in)

	md := logContents{msg: in[dbMSG].(string), level: in[dbLevel].(string),
		time: in[dbTime].(string)}
	// convert the path to bytes
	if in[dbMrxPath] != nil {
		b, _ := json.Marshal(in[dbMrxPath])
		md.mrx = b
	}

	if in[dbChainID] != nil {
		md.chainID = in[dbChainID].(string)
	}

	if in[dbParentID] != nil {
		md.parentID = in[dbParentID].(string)
	}

	if in[dbSource] != nil {
		b, _ := json.Marshal(in[dbSource])
		md.source = b
	}
	// delete all the keys we know
	delete(in, dbMSG)
	delete(in, dbLevel)
	delete(in, dbTime)
	delete(in, dbMrxPath)
	delete(in, dbChainID)
	delete(in, dbParentID)
	delete(in, dbSource)
	// then check for any extra log messages
	if len(in) != 0 {
		b, _ := json.Marshal(in)
		md.other = b
	}

	// lock before writing to the database
	s.mu.Lock()
	defer s.mu.Unlock()

	err = insertLog(s.db, md)
	if err != nil {

		return 0, err
	}

	return len(p), nil
}

/*
{
    "time": "2024-02-13T09:30:11.396575798Z",
    "level": "INFO",
    "source": {
        "function": "github.com/metarex-media/mrx-demo-handlers/mrxhandle/mrxlog.(*MRXHistory).LogInfo",
        "file": "/workspace/mrx-api-demo/mrxhandle/mrxlog/mrxlog.go",
        "line": 168
    },
    "msg": "transofrming to accleration",
    "MRXPath": {
        "MrxID": "MRX.123.456.789.yza",
        "Action": "",
        "Parent": {
            "MrxID": "MRX.123.456.789.yza",
            "Action": "",
            "Origin": ""
        },
        "Origin": ""
    }
}

*/
