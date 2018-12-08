package main

import (
	"log"
	"os"

	"github.com/arborchat/muscadine/types"
)

func loadHist(history types.Archive, histfile string) {
	if histfile != "" {
		file, err := os.Open(histfile)
		if err != nil {
			log.Println("Error opening history", err)
		} else {
			defer file.Close()
			if err = history.Load(file); err != nil {
				log.Println("Error loading history", err)
			} else {
				log.Println("History loaded from", file.Name())
			}
		}
	}
}
func saveHist(history types.Archive, histfile string) {
	if histfile != "" {
		file, err := os.OpenFile(histfile, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			log.Println("Error opening history", err)
		} else {
			defer file.Close()
			if err = file.Truncate(0); err != nil {
				log.Println("Error truncating history", err)
			}
			if err = history.Persist(file); err != nil {
				log.Println("Error saving history", err)
			} else {
				log.Println("History saved to", file.Name())
			}
		}
	}
}
