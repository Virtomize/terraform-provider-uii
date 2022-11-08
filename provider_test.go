package main

import (
	client "github.com/Virtomize/uii-go-api"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderBuild(t *testing.T) {
	uuiClient, err := client.NewClient("token")
	assert.NoError(t, err)

	uut := clientWithStorage{
		VirtomizeClient: uuiClient,
		StorageFolder:   "C:/Tools/Terraform/Isos/",
	}
	iso, err := uut.CreateIso(Iso{Name: "debian", Distribution: "Debian"})
	assert.NoError(t, err)
	assert.Equal(t, iso.Name, "debian")
}

func TestProviderBuildDelete(t *testing.T) {
	uuiClient, err := client.NewClient("token")
	assert.NoError(t, err)

	uut := clientWithStorage{
		VirtomizeClient: uuiClient,
		StorageFolder:   "C:/Tools/Terraform/Isos/",
	}
	iso, err := uut.CreateIso(Iso{Name: "debian", Distribution: "Debian"})
	assert.NoError(t, err)
	assert.Equal(t, iso.Name, "debian")

	err = uut.DeleteIso(iso.Id)
	assert.NoError(t, err)
}
