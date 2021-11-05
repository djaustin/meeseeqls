package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/godror/godror"
)

// Result represents the result of a query send to the database
type Result struct {
	Start       int64         `json:"queryStart"`
	End         int64         `json:"queryEnd"`
	Duration    time.Duration `json:"duration"`
	Source      string        `json:"fromSite"`
	Destination string        `json:"toSite"`
}

func validateStringArg(arg *string, msg string) bool {
	if arg == nil || len(*arg) < 1 {
		fmt.Println(msg)
		return false
	}
	return true
}

func main() {

	username := flag.String("u", "", "The username of the user used to connect to the database")
	password := flag.String("p", "", "The password of the user used to connect to the database")
	connString := flag.String("c", "", "The connection string to connect to the database")
	inputFile := flag.String("f", "", "The input SQL file to be sent to the server")
	outFilePath := flag.String("o", "out.csv", "The file path of the output CSV file")
	interval := flag.Int("i", 60, "The interval (in seconds) at which to query the database")
	source := flag.String("s", "", "The site the application is being run from")
	destination := flag.String("d", "", "The site the query is being sent to")

	flag.Parse()

	// Make sure that any required flags have been provided to avoid errors later on
	if !(validateStringArg(username, "No username provided") && validateStringArg(password, "No password provided") && validateStringArg(inputFile, "No SQL file path provided") && validateStringArg(connString, "No connection string provided") && validateStringArg(source, "No source location provided") && validateStringArg(destination, "No destination location provided")) {
		flag.PrintDefaults()
		return
	}

	logWithTimestamp(fmt.Sprintf("Reading input file %s...", *inputFile))
	data, err := os.ReadFile(*inputFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	query := string(data)
	logWithTimestamp("Input file read successfully")

	dataSource := fmt.Sprintf(`user="%s" password="%s" connectString="%s"`, *username, *password, *connString)
	logWithTimestamp(fmt.Sprintf("Opening a connection to %s", *connString))
	db, err := sql.Open("godror", dataSource)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()
	logWithTimestamp("Database connection opened successfully")

	logWithTimestamp(fmt.Sprintf("Opening file '%s' for results output", *outFilePath))
	// Opening the file in append mode so that we can add to the tail of the file without overwriting the content
	outFile, err := os.OpenFile(*outFilePath, os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer outFile.Close()
	logWithTimestamp("Output file opened successfully")

	encoder := json.NewEncoder(outFile)

	tick := time.NewTicker(time.Duration(*interval) * time.Second)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	logWithTimestamp("Starting ticker...")
	for {
		select {
		case <-sc:
			logWithTimestamp("Exiting...")
			return
		case <-tick.C:
			go func() {
				start := time.Now()

				executeQuery(db, query)

				end := time.Now()
				duration := end.Sub(start)
				result := Result{
					Start:       start.UnixMilli(),
					End:         end.UnixMilli(),
					Duration:    duration,
					Source:      *source,
					Destination: *destination,
				}
				logWithTimestamp("Query complete, writing to file...")

				encoder.Encode(result)
				logWithTimestamp("File write complete")
			}()

		}
	}
}

func executeQuery(db *sql.DB, query string) {
	logWithTimestamp("Executing query...")
	rows, err := db.Query(query)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	var row string
	for rows.Next() {
		rows.Scan(&row)
	}

}

func logWithTimestamp(msg string) {
	fmt.Printf("%s\t%s\n", time.Now().Format(time.RFC3339), msg)

}
