package server

// Config ...
type Config struct {
	BindAdrr string
	DBURL    string
	DBName   string
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{
		DBURL:    "mongodb://localhost:27017",
		DBName:   "medobs-test",
		BindAdrr: ":8080",
	}
}
