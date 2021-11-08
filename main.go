package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/djaustin/meeseeqls/config"
	_ "github.com/godror/godror"
	"github.com/spf13/viper"
)

var (
	Version = "development"
)

// Result represents the result of a query send to the database
type Result struct {
	Start       int64         `json:"queryStart"`
	End         int64         `json:"queryEnd"`
	Duration    time.Duration `json:"duration"`
	Source      string        `json:"fromSite"`
	Destination string        `json:"toSite"`
}

func main() {
	showVersion := flag.Bool("version", false, "Displays version information about the application")
	flag.Parse()
	if *showVersion {
		fmt.Printf("Version:\t%v\n", Version)
		return
	}
	conf, err := initConfig()
	if err != nil {
		log.Fatalf("error initialising application config: %v", err)
	}

	log.Printf("reading input file %s...", conf.Database.SQLFilePath)
	data, err := os.ReadFile(conf.Database.SQLFilePath)
	if err != nil {
		log.Fatalln(err)
	}
	query := string(data)
	log.Print("input file read successfully")

	dataSource := fmt.Sprintf(`user="%s" password="%s" connectString="%s"`, conf.Database.Username, conf.Database.Password, conf.Database.ConnectionString)
	log.Printf("opening a connection to %s", conf.Database.ConnectionString)
	db, err := sql.Open("godror", dataSource)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()
	log.Print("database connection opened successfully")

	log.Printf("opening file '%s' for results output", conf.Output.FilePath)
	// Opening the file in append mode so that we can add to the tail of the file without overwriting the content
	outFile, err := os.OpenFile(conf.Output.FilePath, os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer outFile.Close()
	log.Print("output file opened successfully")

	encoder := json.NewEncoder(outFile)

	tick := time.NewTicker(time.Duration(conf.Database.PollingInterval) * time.Second)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	log.Print("starting ticker...")
	for {
		select {
		case <-sc:
			log.Print("exiting...")
			return
		case <-tick.C:
			go func() {
				start := time.Now()

				err := executeQuery(db, query)
				if err != nil {
					log.Print(err)
					return
				}

				end := time.Now()
				duration := end.Sub(start)
				result := Result{
					Start:       start.UnixMilli(),
					End:         end.UnixMilli(),
					Duration:    duration,
					Source:      conf.Metadata.Source,
					Destination: conf.Metadata.Destination,
				}
				log.Print("query complete, writing to file...")

				encoder.Encode(result)
				log.Print("file write complete")
			}()

		}
	}
}

// initConfig initiliases a Config object from a config file, assigning default values where appropriate
func initConfig() (config.Config, error) {
	config := config.Config{}
	viper.SetConfigName("meeseeqls.yml")
	viper.SetConfigType("yaml")
	viper.SetDefault("database.pollingInterval", 60)
	viper.SetDefault("database.sqlFilePath", "./query.sql")
	viper.SetDefault("output.filePath", "./meeseeqls.csv")
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("error finding working directory: %v", err)
	} else {
		viper.AddConfigPath(wd)
	}
	err = viper.ReadInConfig()
	if err != nil {
		return config, fmt.Errorf("error reading configuration: %w", err)
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		return config, fmt.Errorf("error unmarshalling config: %w", err)
	}
	return config, nil
}

func executeQuery(db *sql.DB, query string) error {
	log.Print("executing query...")
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	var row string
	for rows.Next() {
		rows.Scan(&row)
	}
	return nil
}
