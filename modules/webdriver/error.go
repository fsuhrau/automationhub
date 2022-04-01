package webdriver

type WDAError struct {
	Value struct {
		Error     string `json:"error"`
		Message   string `json:"message"`
		Traceback string `json:"traceback"`
	} `json:"value"`
	SessionId string `json:"sessionId"`
}
