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

// GetWebsiteByName retrieves a website by name from the list of all websites
func (client Client) GetWebsiteByName(ctx context.Context, name string) (*Website, error) {
	websites, err := client.ListWebsites(ctx)
	if err != nil {
		return nil, err
	}
	
	for _, site := range websites {
		if site.Name == name {
			// Get full website details by ID
			return client.ReadWebsite(ctx, site.ID)
		}
	}
	
	return nil, nil
}
