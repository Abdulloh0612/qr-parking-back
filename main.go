package main

import (
	"log"

	"qr-parking/server"
)

func main() {
	s, err := server.New()
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}
