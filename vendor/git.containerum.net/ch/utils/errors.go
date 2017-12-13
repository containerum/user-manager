package utils

type Error struct {
	Text string `json:"error"`
}

func (e *Error) Error() string {
	return e.Text
}