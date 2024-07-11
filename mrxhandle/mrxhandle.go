package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/metarex-media/mrx-demo-handlers/mrxhandle/coordinates"
	"github.com/metarex-media/mrx-demo-handlers/mrxhandle/mrxlog"
	"github.com/metarex-media/mrx-demo-handlers/mrxhandle/mrxlog/utils"
	"github.com/metarex-media/mrx-demo-handlers/mrxhandle/temp"
	"github.com/mmTristan/mrx-tool/decode"
	"github.com/mmTristan/mrx-tool/encode"
)

// include a chain of reliability
// want a trace call of
type errMesage struct {
	Message string
	MrxID   string
	Action  string
	UserID  string
	Parent  *errMesage
}

func (e errMesage) LogKey() string {

	return "MRXPath"
}

// addChild functoin,
// how do I pass it about make it all recursive

func init() {

	/*
		cfg := zap.Config{
			Encoding:         "json",
			Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
			EncoderConfig: zapcore.EncoderConfig{
				MessageKey: "message",

				LevelKey:    "level",
				EncodeLevel: zapcore.CapitalLevelEncoder,

				TimeKey:    "time",
				EncodeTime: zapcore.ISO8601TimeEncoder,

				CallerKey:    "caller",
				EncodeCaller: zapcore.ShortCallerEncoder,
			},
		}

	*/
	// logger2, _ := cfg.Build()
	// loggerw := slog.New(zapslog.NewHandler(zapLogger.Core(), &zapslog.HandlerOptions{AddSource: true}))

}

// KnownFormats uses the MRXID to chooses how to handle any incoming metadata
var knownFormats = map[string]func(MRX *mrxlog.MRXHistory, input []any, API, APISpec, Action string) (any, error){
	coordinates.MRXID:    coordinates.TransformXYZ,
	temp.MRXID:           temp.Transform,
	coordinates.MRXIDVel: coordinates.HandleVelocity,
}

// MRX handle is going to do all the data handling
// look up the register
// get the API
// swagger hub validation
// run the API
// then do some transformation
func main() {

	// set the register up
	fmt.Println(register())

	// have the database open for register calls

	logging := []func(*slog.HandlerOptions){utils.ColourConsole, utils.Console} //, utils.JSON}
	for _, l := range logging {
		l(&slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug})
		// make a loop with the keys
		// this is the current demo approach
		demo()
	}
	utils.JSON(&slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}, "./tmp/")
	demo()

	// quick data generator

	f, _ := os.Create("./demodata/xyzBadDemo.mrx")
	in := make(chan []byte, 100)

	//var inputData []any
	//b, _ := os.ReadFile("./demodata/transformedDataVelocity.mrx.json")
	//err = json.Unmarshal(b, &inputData)

	inputData := make([]map[string]int, 100)
	x, y, z := 0, 0, 0
	for i := 0; i < 100; i++ {
		x = rand.Intn(120) // + i
		y = rand.Intn(100) // + i
		z = rand.Intn(100) // + i
		inputData[i] = map[string]int{"x1": x,
			"y1": y,
			"z1": z}
	}

	for _, dat := range inputData {
		inb, _ := json.MarshalIndent(dat, "", "    ")
		in <- inb
	}

	go func() {
		for {

			if len(in) == 0 {
				close(in)
				break

			}
		}
	}()

	demoConfig := encode.Configuration{Version: "pre alpha",
		Default:          encode.StreamProperties{StreamType: "Example XYZ data", FrameRate: "1/1", NameSpace: "https://metarex.media/reg/MRX.123.456.789.mno"},
		StreamProperties: map[int]encode.StreamProperties{0: {NameSpace: "MRX.123.456.789.mno"}},
	}
	err := encode.EncodeSingleDataStream(f, in, demoConfig)
	fmt.Println(err)
	/*

		DemoData := make([]tempsIn, 50)
		for i := 0; i < 50; i++ {
			temper := float64(rand.Intn(50) - 20)
			DemoData[i] = tempsIn{Temperature: temper, Feels: temper - 5, MinTemp: temper - 10, MaxTemp: temper + 10}
		}

		b2, _ := json.MarshalIndent(DemoData, "", "    ")
		f, _ := os.Create("./demodata/demodataTemp.json")
		f.Write(b2)*/

}

func demo() {

	sqliteDatabase, err := sql.Open("sqlite3", "register.sql") // Open the created SQLite File
	if err != nil {
		panic(err)
	}
	defer sqliteDatabase.Close()

	target := []string{"./demodata/xyzDemo.mrx", "./demodata/temperatureDemo.mrx", "./demodata/xyzDemoVel.mrx", "./demodata/velDemo.mrx"}
	output := []string{"./demodata/transformedDataXYZ.mrx.json", "./demodata/transformedDataTemp.mrx.json",
		"./demodata/transformedDataVelocity.mrx.json", "./demodata/transformedDataAcceleration.mrx.json"}
	action := []string{"", "", coordinates.Velocity, ""}

	for i, mrxFile := range target {

		slog.Info(fmt.Sprintf("opening %v", mrxFile), "running", "this extra stuff")
		//	slog.Debug("Debug Colour")
		//	slog.Warn("Warn Colour")

		// open the MRX
		f, err := os.Open(mrxFile)
		if err != nil {
			slog.Error(err.Error())
			continue
		}

		// extract the data as a pseudo stream
		dataPoints, err := decode.ExtractStreamData(f)
		if err != nil {
			slog.Error(err.Error())
			continue
		}
		// process each data stream
		// exlcuding the manifest
		for _, data := range dataPoints[:len(dataPoints)-1] {

			metadata := mrxlog.NewMRX(data.MRXID) //MrxID: data.MRXID, Action: "Identify", Extra: map[string]any{"Origin": mrxFile}}
			metadata = metadata.PushChild(*mrxlog.NewMRX("DEV"))
			// 	fmt.Println(metadata)
			metadata.LogInfo(fmt.Sprintf("Searching register for %v", data.MRXID))
			reg, err := extractRegValue(sqliteDatabase, data.MRXID)

			if err != nil {
				slog.Error(err.Error())
				continue
			}

			// create a data array of the data stream
			var inputs []any
			for _, d := range data.Data {
				var goFrom any
				json.Unmarshal(d, &goFrom)
				inputs = append(inputs, goFrom)
			}

			var out any

			// check if we already know the data
			// or it can be transformed to a known type
			// @TODO change to search n layers in the API for
			// a format that is known
			if transform, ok := knownFormats[reg.ID]; ok {
				metadata.LogDebug(fmt.Sprintf("Known format found: %v", data.MRXID))
				// run the local transformations
				out, err = transform(metadata, inputs, reg.Mrx.Services.API, reg.Mrx.Services.Spec, action[i])

			} else if transform, ok := knownFormats[reg.Mrx.Services.ID]; ok {
				metadata.LogDebug(fmt.Sprintf("Known format found via API transformation: %v", data.MRXID))
				//metadata.Action = "API transform"
				out, err = transform(metadata, inputs, reg.Mrx.Services.API, reg.Mrx.Services.Spec, action[i])

			} else {
				metadata.LogWarn(fmt.Sprintf("No known metadata found for %v", data.MRXID))
			}

			if err != nil {
				metadata.LogError(err.Error())
			} else {
				//fmt.Println(metadata)
				fout, _ := os.Create(output[i])
				bout, _ := json.MarshalIndent(out, "", "    ")
				fout.Write(bout)
			}
		}
	}
}

// metarexRegister with only the
// values we are concerned about
type metarexReg struct {
	ID string `json:"metarexId"`
	// more metarex information
	Mrx mrx `json:"mrx"`
}

type mrx struct {
	Services services `json:"services"`
}

type services struct {
	API    string `json:"API"`
	Method string `json:"method"`
	ID     string `json:"metarexId"`
	Spec   string `json:"APISchema"`
}

////////////////////
// Register code for the mock register
// for making and reading

// register sets up a mock SQL database
// of the metarex register
// each value is stored as an ID then the json value of the register
func register() error {
	reg, err := generateSQL("register.sql", true)

	if err != nil {
		return err
	}

	// add these entries to the register
	registerValues := []string{regXYZ, regXYZKnown, regTemp, regTempKnown, regVelKnown}

	for _, rv := range registerValues {
		var mrxID metarexReg
		json.Unmarshal([]byte(rv), &mrxID)

		err := insertMetaData(reg, mrxID.ID, rv)
		if err != nil {
			return err
		}
	}

	return nil
}

// make a register call to the SQL database
func extractRegValue(db *sql.DB, mrxID string) (metarexReg, error) {

	// get the key (if it is there)
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM metadata WHERE metarexID = \"%v\"", mrxID)) //("SELECT * FROM metadata WHERE key='Excellent' ORDER BY frameId")

	if err != nil {
		return metarexReg{}, err
	}
	defer rows.Close()

	for rows.Next() { //

		var regID int
		var metarexID string // can implement a struct that self scans and loop through but is that really needed
		var md []byte

		err := rows.Scan(&regID, &metarexID, &md)

		if err != nil {

			return metarexReg{}, err
		}
		var regVal metarexReg
		json.Unmarshal(md, &regVal)

		// each key should only have one value
		// so just return the first one
		return regVal, nil
	}

	return metarexReg{}, fmt.Errorf("no register entry with the ID: %v was found", mrxID)
}

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
	createMetadataTableSQL := `CREATE TABLE metadata (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"metarexID" TEXT,
		"entry" BLOB
	  );` // SQL Statement for Create Table
	//"Extra" BLOB

	statement, err := db.Prepare(createMetadataTableSQL) // Prepare SQL Statement
	if err != nil {
		return err
	}
	_, err = statement.Exec() // Execute SQL Statements

	return err
}

// insert MetaData inserts a single row of metaData
func insertMetaData(db *sql.DB, ID, reg string) error { //Student) {
	// log.Println("Inserting student record ...")
	insertStudentSQL := `INSERT INTO metadata(metarexID, entry) VALUES (?, ?)`
	statement, err := db.Prepare(insertStudentSQL) // Prepare statement.
	// This is to avoid SQL injections
	if err != nil {
		return err
	}
	_, err = statement.Exec(ID, []byte(reg))

	return err
}

/////////////////////
// Register values to be put
// into the database

var regXYZ = `{
    "metarexId": "MRX.123.456.789.mno",
    "name": "API Demo",
    "description": "3d data that can be transformed via an API",
    "media-type": "application/json",
    "timing": "clocked",
    "treatAs": "text",
    "mrx": {
        "specification": "",
        "services": {
            "API": "http://localhost:1323/3dTransform",
            "APISchema": "./openAPI.yaml",
            "metarexID": "MRX.123.456.789.pqr",
            "method": "POST"
        }
    }
}`

var regTemp = `{
    "metarexId": "MRX.123.456.789.stu",
    "name": "API Demo",
    "description": "3d data that can be transformed via an API",
    "media-type": "application/json",
    "timing": "clocked",
    "treatAs": "text",
    "mrx": {
        "specification": "",
        "services": {
            "API": "http://localhost:1323/tempTransform",
            "APISchema": "./openAPI.yaml",
            "metarexID": "MRX.123.456.789.vwx",
            "method": "POST"
        }
    }
}`

var regXYZKnown = `{
    "metarexId": "MRX.123.456.789.pqr",
    "name": "API Demo",
    "description": "Known 3d format that the demo can use",
    "media-type": "application/json",
    "timing": "clocked",
    "treatAs": "text",
    "mrx": {
        "specification": ""
    }
}`

var regTempKnown = `{
    "metarexId": "MRX.123.456.789.vwx",
    "name": "API Demo",
    "description": "Known temperature format that the demo can use",
    "media-type": "application/json",
    "timing": "clocked",
    "treatAs": "text",
    "mrx": {
        "specification": ""
    }
}`

var regVelKnown = `{
    "metarexId": "MRX.123.456.789.yza",
    "name": "API Demo",
    "description": "Known 3d format that the demo can ise",
    "media-type": "application/json",
    "timing": "clocked",
    "treatAs": "text",
    "mrx": {
        "specification": ""
    }
}`

// Have a register writer side value that is just for the moment
