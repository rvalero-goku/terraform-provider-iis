package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("IIS_HOST"); v == "" {
		t.Skip("IIS_HOST must be set for acceptance tests")
	}
	if v := os.Getenv("IIS_ACCESS_KEY"); v == "" {
		t.Skip("IIS_ACCESS_KEY must be set for acceptance tests")
	}
}

func TestAccDataSourceIisCertificates_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"iis": func() (*schema.Provider, error) {
				return Provider(), nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIisCertificatesConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.iis_certificates.test", "certificates.#"),
				),
			},
		},
	})
}

const testAccDataSourceIisCertificatesConfig = `
data "iis_certificates" "test" {}
`
