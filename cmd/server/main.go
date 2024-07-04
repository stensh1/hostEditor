// Package main contains funcs init() and main(), both just load .env variables, create server.S object and start it.
// There is able to trace an interrupt ctrl+c to stop server.
package main

import (
	"dnsService/pkg/server"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// init loads environment variables from .env file
func init() {
	if err := godotenv.Load("cfg/.env"); err != nil {
		log.Fatal("init(): Cannot read env vars:", err)
	} else {
		log.Println("init(): Env vars have been successfully loaded")
	}
}

// main creates server.S object and starts it w\ interruption tracing
func main() {
	// New server object
	s := server.S{}

	// Channel for tracking interrupts
	var ch = make(chan os.Signal, 1)

	// Trace interrupts ctrl+c
	signal.Notify(ch, os.Interrupt, syscall.SIGTSTP)

	// Starting the server
	go s.Start()

	// Catching keyboard combination and stopping the server
	interrupt := <-ch
	s.LogInfo.Println("Server is shutting down by:", interrupt)
	s.Stop()
}
