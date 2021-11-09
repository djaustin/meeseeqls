package oracle_worker

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/djaustin/meeseeqls/config"
	"github.com/djaustin/meeseeqls/worker"
)

type OracleWorker struct {
	DataSource, Query, Name, Source, Destination string
	PollingInterval                              int
	once                                         sync.Once
	db                                           *sql.DB
	resultChan                                   chan<- worker.Result
	exitChan                                     chan struct{}
}

func New(config config.Connector) OracleWorker {
	worker := OracleWorker{
		Name:            config.Name,
		Source:          config.SourceName,
		Destination:     config.DestinationName,
		Query:           config.Query,
		PollingInterval: config.PollingInterval,
		exitChan:        make(chan struct{}),
	}
	worker.DataSource = fmt.Sprintf(
		`user="%s" password="%s" connectString="%s"`,
		config.Database.Username,
		config.Database.Password,
		config.Database.ConnectionString,
	)
	return worker
}

func (o *OracleWorker) Start() (<-chan worker.Result, error) {
	var runErr error
	c := make(chan worker.Result)
	o.resultChan = c
	o.once.Do(func() {
		log.Printf("[%s]\topening a connection...", o.Name)
		db, err := sql.Open("godror", o.DataSource)
		if err != nil {
			runErr = fmt.Errorf("error opening connection: %w", err)
		}
		o.db = db
	})
	go o.run()
	return c, runErr
}

func (o *OracleWorker) run() {
	ticker := time.NewTicker(time.Duration(o.PollingInterval) * time.Second)

	log.Printf("[%s]\tstarting ticker...", o.Name)
	for {
		select {
		case <-o.exitChan:
			log.Printf("[%s]\texiting...", o.Name)
			return
		case <-ticker.C:
			go func() {
				start := time.Now()
				err := o.executeQuery()
				if err != nil {
					log.Printf("[%s]\t%v", o.Name, err)
					return
				}
				end := time.Now()
				duration := end.Sub(start)
				o.resultChan <- worker.Result{
					Start:       start.UnixMilli(),
					End:         end.UnixMilli(),
					Duration:    duration,
					Source:      o.Source,
					Destination: o.Destination,
				}
			}()
		}
	}
}

func (o *OracleWorker) executeQuery() error {
	log.Printf("[%s]\texecuting query...", o.Name)
	rows, err := o.db.Query(o.Query)
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

func (o *OracleWorker) Stop() {
	close(o.exitChan)
}

func (o *OracleWorker) String() string {
	return o.Name
}
