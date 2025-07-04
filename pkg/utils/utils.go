package utils

import "log"

// LogAndReturnError logs an error and returns it.
func LogAndReturnError(msg string, err error) error {
	log.Printf("ERROR: %s: %v", msg, err)
	return err
}
