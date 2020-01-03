package rest

type AddNodeForm struct {
	Host string `json:"host"`
}

type AddLogForm struct {
	RequestID string `json:"request_id"`
	Command   string `json:"command"`
}
