package metronome

// Error is a standard error returned by the DC/OS API.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error converts an API error to a string.
func (err *Error) Error() string {
	return string(err.Code) + " - " + err.Message
}
