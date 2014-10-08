// +build debug

package main

import (
	"log"
	"os"
)

const logPath = "/var/log/pi3g-usbpatcher"

func init() {
	version += "+debug"
	// start logging
	logfile, err := os.OpenFile(logPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		debug(err)
		debug("Falling back to stdout logging")
	} else {
		log.SetOutput(logfile)
	}
	defer logfile.Close()
}

func debug(args ...interface{}) {
	log.Println(args...)
}

func debugf(fmt string, args ...interface{}) {
	log.Printf(fmt, args...)
}
