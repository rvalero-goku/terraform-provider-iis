package iis

import "context"

type WebsiteListItem struct {
	Name            string               `json:"name"`
	ID              string               `json:"id"`
	Status          string               `json:"status"`
	PhysicalPath    string               `json:"physical_path"`
	ApplicationPool ApplicationPoolRef   `json:"application_pool"`
}

type ApplicationPoolRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type WebsiteListResponse struct {
	Websites []WebsiteListItem `json:"websites"`
}

func (client Client) ListWebsites(ctx context.Context) ([]WebsiteListItem, error) {
	var res WebsiteListResponse
	err := getJson(ctx, client, "/api/webserver/websites?fields=*", &res)
	if err != nil {
		return nil, err
	}
	return res.Websites, nil
}
