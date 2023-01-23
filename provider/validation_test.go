package provider

import (
	client "github.com/Virtomize/uii-go-api"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHostNameValidation(t *testing.T) {
	assert.NoError(t, validateHostname("host"))
	assert.NoError(t, validateHostname("1ho3st2"))
	assert.NoError(t, validateHostname("host-example1"))

	assert.Error(t, validateHostname("-host"))
	assert.Error(t, validateHostname("host?"))
}

func TestDistributionValidation(t *testing.T) {
	debian10 := client.OS{Architecture: "64", DisplayName: "Debian 10 x64", Distribution: "debian", Version: "10"}
	debian11 := client.OS{Architecture: "64", DisplayName: "Debian 11 x64", Distribution: "debian", Version: "11"}

	assert.NoError(t, validateDistribution("debian", "10", "64", []client.OS{debian10}))
	assert.NoError(t, validateDistribution("debian", "10", "", []client.OS{debian10}))
	assert.NoError(t, validateDistribution("debian", "10", "64", []client.OS{debian10, debian11}))
	assert.NoError(t, validateDistribution("debian", "10", "64", []client.OS{}))

	assert.Error(t, validateDistribution("debian", "10", "64", []client.OS{debian11}))
	assert.Error(t, validateDistribution("debian", "11", "64", []client.OS{debian10}))
	assert.Error(t, validateDistribution("debian", "10", "8", []client.OS{debian10}))
}
