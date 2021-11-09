package config

type Config struct {
	Output     Output
	Connectors []Connector
}

type Connector struct {
	Name, Query, SourceName, DestinationName string
	PollingInterval                          int
	Database                                 Database
}
type Output struct {
	Path string
}

type Database struct {
	Username, Password, ConnectionString string
}
