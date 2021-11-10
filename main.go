package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/djaustin/meeseeqls/config"
	"github.com/djaustin/meeseeqls/oracle_worker"
	"github.com/djaustin/meeseeqls/worker"
	_ "github.com/godror/godror"
	"github.com/spf13/viper"
)

var (
	Version = "development"
)

// Result represents the result of a query send to the database

func main() {
	conf, err := initConfig()
	if err != nil {
		log.Fatalln("failed to initialise config", err)
	}

	results := processConnectors(conf.Connectors)
	go writeResults(conf.Output.Path, results)

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-exit
}

// initConfig initiliases a Config object from a config file, assigning default values where appropriate
func initConfig() (config.Config, error) {
	config := config.Config{}
	viper.SetConfigName("meesqls")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/meesqls/")
	viper.AddConfigPath("$HOME/.meesqls")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		return config, fmt.Errorf("error reading configuration: %w", err)
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		return config, fmt.Errorf("error unmarshalling config: %w", err)
	}
	return config, nil
}

func writeResults(path string, results <-chan worker.Result) {
	outFile, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Fatalln(fmt.Sprintf("unable to open output file %q for writing: %v", path, err))
	}
	defer outFile.Close()
	encoder := json.NewEncoder(outFile)

	for msg := range results {
		err = encoder.Encode(msg)
		if err != nil {
			log.Printf("couldn't write to file: %v", err)
			continue
		}
	}
}

func processConnectors(connectors []config.Connector) <-chan worker.Result {
	workers := make([]worker.Worker, len(connectors))
	// Generate workers
	for i, connector := range connectors {
		worker := oracle_worker.New(connector)
		workers[i] = &worker
	}
	// Fan in channels
	results := make(chan worker.Result)
	for _, w := range workers {
		c, err := w.Start()
		if err != nil {
			log.Printf("error starting worker %q: %v", w, err)
		}
		go func() {
			for {
				results <- <-c
			}
		}()
	}
	return results
}
