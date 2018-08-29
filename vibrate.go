package main

import (
	"log"
	"os/exec"
)

func Vibrate() {
	c := exec.Command("/bin/sh", "-c", "dbus-send --system --print-reply --dest=com.nokia.mce /com/nokia/mce/request com.nokia.mce.request.req_start_manual_vibration int32:255 int32:1000")

	go func() {
		err := c.Run()
		log.Println("dbus vibrate", err)
	}()
}
