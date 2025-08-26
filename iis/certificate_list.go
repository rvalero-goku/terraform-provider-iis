package iis

import "context"

type CertificateListResponse struct {
	Certificates []Certificate `json:"certificates"`
}

func (client Client) ListCertificates(ctx context.Context) ([]Certificate, error) {
	var res CertificateListResponse
	err := getJson(ctx, client, "/api/certificates", &res)
	if err != nil {
		return nil, err
	}
	return res.Certificates, nil
}
