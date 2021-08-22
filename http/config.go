package http

type Config struct {
	HttpHost  string `json:"http_host"`
	HttpPort  uint   `json:"http_port"`
	PprofPort uint   `json:"pprof_port"`
}
