package config

type Config struct {
	Output   Output
	Metadata Metadata
	Database Database
}

type Output struct {
	FilePath string
}

type Metadata struct {
	Source, Destination string
}

type Database struct {
	Username, Password, ConnectionString, SQLFilePath string
	PollingInterval                                   int
}
