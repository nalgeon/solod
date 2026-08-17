package os_test

// errText returns the message of err. A nil error gives "nil".
func errText(err error) string {
	if err == nil {
		return "nil"
	}
	return err.Error()
}
