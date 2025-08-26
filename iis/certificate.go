package iis

type Certificate struct {
	Alias      string `json:"alias"`
	ID         string `json:"id"`
	IssuedBy   string `json:"issued_by"`
	Subject    string `json:"subject"`
	Thumbprint string `json:"thumbprint"`
}
