package main

import (
	"github.com/nemothekid/tinybusters/server"
	"log"
	"os"
	"os/signal"
)

func main() {
	log.Println("[Init] Starting Server...")
	exitChannel := make(chan bool)
	exitFunc := func() {
		exitChannel <- true
	}
	config, err := server.ReadConfig()
	if err != nil {
		log.Println("[Init] Config error. Failed to start server.")
		log.Println(err)
		os.Exit(1)
	}
	config.Quit = exitFunc

	if config.ClientPort != 0 {
		log.Println("[Init] Starting Client...")
		client := server.NewClientServer(config)
		go func() {
			client.Serve()
		}()
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		forceExit := false
		for _ = range c {
			if forceExit {
				os.Exit(2)
			} else {
				go func() {
					exitFunc()
				}()
				forceExit = true
			}
		}
	}()

	<-exitChannel
	log.Println("Bye")
}
