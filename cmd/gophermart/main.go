package main

import (
	"github.com/1Asi1/gophermart/internal/server"
)

func main() {
	s := server.New()
	s.Run()
}
