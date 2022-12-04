package main

import (
	client "github.com/Virtomize/uii-go-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func getToken() string {
	return os.Getenv("UIITOKEN")
}

func createProvider() *schema.Provider {
	provider := Provider()
	config := terraform.ResourceConfig{Config: map[string]interface{}{}, Raw: map[string]interface{}{}}
	config.Config["apitoken"] = getToken()
	provider.Configure(nil, &config)

	return provider
}

func TestProvider(t *testing.T) {
	provider := createProvider()

	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderBuild(t *testing.T) {
	uuiClient, err := client.NewClient(getToken())
	assert.NoError(t, err)

	storage := createDefaultStoragePath()
	defer os.RemoveAll(storage)

	uut := clientWithStorage{
		VirtomizeClient: uuiClient,
		StorageFolder:   storage,
		TimeProvider:    defaultTimeProvider{},
	}
	iso, err := uut.CreateIso(Iso{Name: "debian_iso", Distribution: "debian", HostName: "host", Version: "11", Networks: []Network{{
		DHCP:       true,
		NoInternet: false,
	}}})

	assert.NoError(t, err)
	assert.Equal(t, iso.Name, "debian_iso")
}

func TestProviderBuildDelete(t *testing.T) {
	uuiClient, err := client.NewClient(getToken())
	assert.NoError(t, err)

	storage := createDefaultStoragePath()
	defer os.RemoveAll(storage)

	uut := clientWithStorage{
		VirtomizeClient: uuiClient,
		StorageFolder:   storage,
		TimeProvider:    defaultTimeProvider{},
	}
	iso, err := uut.CreateIso(Iso{Name: "debian_iso", Distribution: "debian", HostName: "host", Version: "11", Networks: []Network{{
		DHCP:       true,
		NoInternet: false,
	}}})
	assert.NoError(t, err)
	assert.Equal(t, iso.Name, "debian_iso")

	err = uut.DeleteIso(iso.Id)
	assert.NoError(t, err)
}
