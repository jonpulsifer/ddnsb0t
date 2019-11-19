package ddns

// Request is the expected struct we want as a request
type Request struct {
	IPAddress string `json:"IPAddress"`
	DNSName   string `json:"DNSName"`
	APIToken  string `json:"APIToken"`
}

// Response is the response we expect to return
type Response struct {
	Status    string `json:"status"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions,omitempty"`
}
