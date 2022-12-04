package main

import (
	"context"
	"os"
	"path"

	client "github.com/Virtomize/uii-go-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const TokenEnvName = "VIRTOMIZE_API_TOKEN"
const StorageEnvName = "VIRTOMIZE_ISO_CACHE"

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"apitoken": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc(TokenEnvName, nil),
			},

			"localstorage": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(StorageEnvName, nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"virtomize_iso": resourceIso(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	token := d.Get("apitoken").(string)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if token == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "No token provided to create Virtomize client",
			Detail:   "Unable to retrieve token for authenticated Virtomize client. When using environment variables use " + TokenEnvName,
		})

		return nil, diags
	}

	c, err := client.NewClient(token)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Virtomize client",
			Detail:   "Unable to authenticate user for authenticated Virtomize client",
		})

		return nil, diags
	}

	defaultStoragePath := ""

	localStorageOverride := d.Get("localstorage")
	if localStorageOverride != nil {
		defaultStoragePath = localStorageOverride.(string)
	} else {
		defaultStoragePath = createDefaultStoragePath()
	}

	return &clientWithStorage{VirtomizeClient: c, StorageFolder: defaultStoragePath, TimeProvider: defaultTimeProvider{}}, diags
}

func createDefaultStoragePath() string {
	defaultStoragePath := path.Join(os.TempDir(), "uiiterraform")
	_ = os.Mkdir(defaultStoragePath, os.ModePerm)
	return defaultStoragePath
}
