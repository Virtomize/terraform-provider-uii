package provider

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"testing"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	ProviderName: providerserver.NewProtocol6WithError(NewFromVersion("test")()),
}

func TestSimpleIsoLifeCycle(t *testing.T) {
	testConfiguration := `
provider "virtomize" {
  # apitoken = retrieved from env variables
  # localstorage = use local folder
}

resource "virtomize_iso" "debian_iso" {
    name = "debian_iso"
    distribution = "debian"
    version = "11"
    hostname = "examplehost"
    networks = [{
      dhcp = true
      no_internet = false
    }]
 }`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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

func TestSimpleIsoLifeCycleStaticConfig(t *testing.T) {
	testConfiguration := `
provider "virtomize" {
  # apitoken = retrieved from env variables
  # localstorage = use local folder
}

resource "virtomize_iso" "debian_iso" {
    name = "debian_iso"
    distribution = "debian"
    version = "11"
    hostname = "examplehost"
    networks = [{
      dhcp = false
      domain = "custom_domain"
      mac = "ca:8c:65:0d:e7:58"
      ip_net = "10.0.0.0/24"
      gateway = "10.0.0.1"
      dns = ["1.1.1.1", "8.8.8.8"]
      no_internet = false
    }]
 }`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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

func TestLoopBackIpIsDetectedByValidation(t *testing.T) {
	testConfiguration := `
provider "virtomize" {
}

resource "virtomize_iso" "debian_iso" {
    name = "debian_iso"
    distribution = "debian"
    version = "11"
    hostname = "examplehost"
    networks = [{
      dhcp = false
      domain = "custom_domain"
      mac = "ca:8c:65:0d:e7:58"
      ip_net = "10.0.0.1/24"
      gateway = "10.0.0.1"
      dns = ["1.1.1.1", "8.8.8.8"]
      no_internet = false
    }]
 }`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testConfiguration,
				Check: resource.ComposeTestCheckFunc(
					checkSimpleIsoProperties,
				),
			},
		},
		ErrorCheck: func(err error) error {
			if err == nil {
				return errors.New("expected error but got none")
			}

			// return original error if no match
			return err
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
		return fmt.Errorf("Widget Id is not set")
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
