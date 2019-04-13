package main

import (
	"io"
	"log"
)

func CloseOrFatal(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func CloseOrLog(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		log.Fatal(err)
	}
}
