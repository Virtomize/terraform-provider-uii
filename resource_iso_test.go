package main

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"testing"
)

func TestSimpleIsoLifeCycle(t *testing.T) {
	testAccProvider := Provider()
	testAccProviders := map[string]*schema.Provider{
		"virtomize": testAccProvider,
	}

	testConfiguration := `
resource "virtomize_iso" "debian_iso" {
    name = "debian_iso"
    distribution = "debian"
    version = "11"
    hostname = "examplehost"
    networks {
      dhcp = true
      no_internet = false
    }
 }`

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testConfiguration,
				Check: resource.ComposeTestCheckFunc(
					checkSimpleIsoProperties,
				),
			},
		},
	})
}

func checkSimpleIsoProperties(state *terraform.State) error {
	resource_name := "virtomize_iso.debian_iso"
	rs, ok := state.RootModule().Resources[resource_name]
	if !ok {
		return fmt.Errorf("Not found: %s", resource_name)
	}

	if rs.Primary.ID == "" {
		return fmt.Errorf("Widget ID is not set")
	}

	attributes := rs.Primary.Attributes

	err := verifyAttribute(attributes, "name", "debian_iso")
	if err != nil {
		return err
	}

	return nil
}

func verifyAttribute(attributes map[string]string, key string, expectedValue string) error {
	value, ok := attributes[key]
	if !ok {
		return fmt.Errorf("could not find attribute \"%s\"", key)
	}

	if value != expectedValue {
		return fmt.Errorf("resource attribute \"%s\" was \"%s\" but expected \"%s\"", key, value, expectedValue)
	}

	return nil
}

// testAccPreCheck validates the necessary test API keys exist
// in the testing environment
func testAccPreCheck(t *testing.T) {
	if v := os.Getenv(TokenEnvName); v == "" {
		t.Fatalf("%s must be set for acceptance tests", TokenEnvName)
	}
}
