package main

import (
	"log"
	"os"

	"github.com/lebensborned/medobs-test/server"
)

func init() {
	os.Setenv("REFRESH_SECRET", "alonewithmyself345r2345") // this should be in an env file
	os.Setenv("ACCESS_SECRET", "imghoulletmedie402340234") // this should be in an env file
}
func main() {
	config := server.NewConfig()
	srv := server.New(config)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}
