package main

import (
	"github.com/1Asi1/gophermart/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	s := server.New()
	s.Run()
}
