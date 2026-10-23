package main

import (
	"log"
	"os"
)

func getEnvVariable(s string) string {
	variable := os.Getenv(s)
	if variable == "" {
		log.Fatalf("%s must be set", s)
	}

	return variable
}
