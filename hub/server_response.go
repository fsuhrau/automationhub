package hub

type ServerResponse struct {
	SessionID string      `json:"sessionId"`
	State     string      `json:"state"`
	Message   string      `json:"message"`
	HCode     int64       `json:"hcode"`
	Status    int64       `json:"status"`
	Value     interface{} `json:"value"`
	Payload   interface{} `json:"payload"`
}
