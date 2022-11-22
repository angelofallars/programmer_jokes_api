package main

import "fmt"

type InputValidationError struct {
	message string
}

func (error *InputValidationError) Error() string {
	return fmt.Sprintf("Input validation error occured: %s", error.message)
}
