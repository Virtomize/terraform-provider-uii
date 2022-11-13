package main

import (
	"context"

	client "github.com/Virtomize/uii-go-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const TOKEN_ENV_NAME = "VIRTOMIZE_API_TOKEN"
const STORAGE_ENV_NAME = "VIRTOMIZE_ISO_CACHE"

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"apitoken": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(TOKEN_ENV_NAME, nil),
			},

			"localstorage": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc(TOKEN_ENV_NAME, nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"virtomize_iso": resourceIso(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	token := d.Get("apitoken").(string)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if token == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "No token provided to create Virtomize client",
			Detail:   "Unable to retrieve token for authenticated Virtomize client. When using environment variables use " + TOKEN_ENV_NAME,
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

	return &clientWithStorage{VirtomizeClient: c, StorageFolder: "C:\\Tools\\Terraform\\Isos\\"}, diags
}
