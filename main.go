package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	_ "github.com/godror/godror"
)

func main() {
	connString := flag.String("c", "", "The connection string to connect to the database")
	inputFile := flag.String("f", "", "The input SQL file to be sent to the server")
	flag.Parse()

	if connString == nil || len(*connString) < 1 {
		fmt.Println("No connection string provided. Exiting...")
		return
	}

	if inputFile == nil || len(*inputFile) < 1 {
		fmt.Println("No input file provided. Exiting...")
		return
	}

	data, err := os.ReadFile(*inputFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	query := string(data)

	db, err := sql.Open("godror", *connString)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	rows, err := db.Query(query)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

}
