package main

import (
	"log"
	"os/exec"
)

func Espeak(mesg string) {
	// TODO: Just send to stdin
	c := exec.Command("espeak", "-v", "en+f4", "-s", "120", mesg)
	go func() {
		err := c.Run()
		log.Println("espeak done", err)
	}()
}
