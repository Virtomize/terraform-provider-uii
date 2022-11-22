package main

import (
	uiiclient "github.com/Virtomize/uii-go-api"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
	"time"
)

type uiiClientMock struct {
	buildCount *int
}

func (c uiiClientMock) Build(_ string, _ uiiclient.BuildArgs, _ uiiclient.BuildOpts) error {
	if c.buildCount != nil {
		*c.buildCount++
	}
	return nil
}

type constantTimeProvider struct {
	CurrentTime time.Time
}

func (c constantTimeProvider) Now() time.Time {
	return c.CurrentTime
}

func TestClientAddedIsoCanBeRead(t *testing.T) {
	dir, err := os.MkdirTemp("", "uii_unittest")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	isoName := "name"
	distributionName := "debian"
	version := "1"
	hostName := "host"

	uuiMock := uiiClientMock{}
	uut := clientWithStorage{
		VirtomizeClient: uuiMock,
		StorageFolder:   dir,
		TimeProvider:    defaultTimeProvider{},
	}

	createIso, createError := uut.CreateIso(Iso{
		Name:         isoName,
		Distribution: distributionName,
		Version:      version,
		HostName:     hostName,
		Networks:     []Network{},
		Optionals:    BuildOpts{},
	})

	readIso, readError := uut.ReadIso(createIso.Id)

	assert.NoError(t, createError)
	assert.NotNil(t, createIso)
	assert.Equal(t, isoName, createIso.Name)
	assert.Equal(t, distributionName, createIso.Distribution)
	assert.Equal(t, version, createIso.Version)
	assert.Equal(t, hostName, createIso.HostName)
	assert.True(t, createIso.CreationTime.After(time.Now().Add(-1*time.Second)))
	assert.True(t, createIso.CreationTime.Before(time.Now().Add(1*time.Second)))

	assert.NoError(t, readError)
	assert.NotNil(t, readIso)
	assert.Equal(t, isoName, readIso.Name)
	assert.Equal(t, distributionName, readIso.Distribution)
	assert.Equal(t, version, readIso.Version)
	assert.Equal(t, hostName, readIso.HostName)
}

func TestClientUpdatingDistributionCreatesANewIso(t *testing.T) {
	dir, err := os.MkdirTemp("", "uii_unittest")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)
	buildCount := 0

	isoName := "name"
	distributionName := "debian"
	version := "1"
	hostName := "host"

	updatedDistributionName := "redhat"

	uuiMock := uiiClientMock{buildCount: &buildCount}
	uut := clientWithStorage{
		VirtomizeClient: uuiMock,
		StorageFolder:   dir,
		TimeProvider:    defaultTimeProvider{},
	}

	createIso, createError := uut.CreateIso(Iso{
		Name:         isoName,
		Distribution: distributionName,
		Version:      version,
		HostName:     hostName,
		Networks:     []Network{},
		Optionals:    BuildOpts{},
	})

	updateErr := uut.UpdateIso(isoName, Iso{
		Name:         isoName,
		Distribution: updatedDistributionName,
		Version:      version,
		HostName:     hostName,
		Networks:     []Network{},
		Optionals:    BuildOpts{},
	})

	readIso, readError := uut.ReadIso(createIso.Id)

	assert.NoError(t, createError)
	assert.NotNil(t, createIso)

	assert.NoError(t, updateErr)
	assert.Equal(t, 2, buildCount)

	assert.NoError(t, readError)
	assert.NotNil(t, readIso)
	assert.Equal(t, isoName, readIso.Name)
	assert.Equal(t, updatedDistributionName, readIso.Distribution)
	assert.Equal(t, version, readIso.Version)
	assert.Equal(t, hostName, readIso.HostName)
}

func TestClientReadingExpiredIsoRebuildsIt(t *testing.T) {
	dir, err := os.MkdirTemp("", "uii_unittest")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)
	buildCount := 0

	isoName := "name"
	distributionName := "debian"
	version := "1"
	hostName := "host"
	creationTimeOverride := time.Now().Add(5 * 24 * time.Hour)

	uuiMock := uiiClientMock{buildCount: &buildCount}
	uut := clientWithStorage{
		VirtomizeClient: uuiMock,
		StorageFolder:   dir,
		TimeProvider:    defaultTimeProvider{},
	}

	createIso, createError := uut.CreateIso(Iso{
		Name:         isoName,
		Distribution: distributionName,
		Version:      version,
		HostName:     hostName,
		Networks:     []Network{},
		Optionals:    BuildOpts{},
	})

	uut.TimeProvider = constantTimeProvider{CurrentTime: creationTimeOverride}
	readIso, readError := uut.ReadIso(createIso.Id)

	assert.NoError(t, createError)
	assert.NotNil(t, createIso)

	assert.Equal(t, 2, buildCount)

	assert.NoError(t, readError)
	assert.NotNil(t, readIso)
	assert.Equal(t, isoName, readIso.Name)
}
