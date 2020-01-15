package api

type Config struct {
	Enabled bool   `json:"enabled"`
	V3      bool   `json:"v3"`
	Port    string `json:"port"`
	Nodemap struct {
		Enabled bool   `json:"enabled"`
		Mode    string `json:"mode"`
		Geodb   string `json:"mmdb"`
	} `json:"nodemap"`
}
