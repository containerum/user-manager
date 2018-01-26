package routes

import (
	"fmt"

	"gopkg.in/go-playground/validator.v8"
)

//BindError is a type for bind errors
type BindError struct {
	Error string `json:"error"`
}

//ParseBindErorrs parses errors from message content binding
func ParseBindErorrs(in error) []BindError {
	var out []BindError
	for _, v := range in.(validator.ValidationErrors) {
		switch v.Tag {
		case "required":
			out = append(out, BindError{fmt.Sprintf(fieldShouldExist, v.Name)})
		case "email":
			out = append(out, BindError{fmt.Sprintf(fieldShouldBeEmail, v.Name)})
		default:
			out = append(out, BindError{fmt.Sprintf(fieldDefaultProblem, v.Name, v.Tag)})
		}
	}
	return out
}
