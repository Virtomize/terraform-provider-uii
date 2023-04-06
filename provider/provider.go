package provider

import (
	"context"
	"fmt"
	"os"
	"path"

	client "github.com/Virtomize/uii-go-api"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

//nolint: gosec // wrong
const TokenEnvName = "VIRTOMIZE_API_TOKEN"
const StorageEnvName = "VIRTOMIZE_ISO_CACHE"
const ProviderName = "virtomize"

type uiiProviderModel struct {
	APIToken     types.String `tfsdk:"apitoken"`
	LocalStorage types.String `tfsdk:"localstorage"`
}

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &uiiProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New() provider.Provider {
	return &uiiProvider{}
}

func NewFromVersion(version string) func() provider.Provider {
	return func() provider.Provider {
		return &uiiProvider{}
	}
}

// uiiProvider is the provider implementation.
type uiiProvider struct{}

// Metadata returns the provider type name.
func (p *uiiProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = ProviderName
	resp.Version = "0.0.1"
}

// Schema defines the provider-level schema for configuration data.
func (p *uiiProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"apitoken": schema.StringAttribute{
				Optional: true,
				// TODO: make this required, but default to env variable. Check back on
				// https://discuss.hashicorp.com/t/terraform-plugin-framework-required-attribute-and-environment-variables/47505
				Sensitive:           true,
				Description:         fmt.Sprintf("The API token for accessing Virtomize UII. If none is provided, the fallback is to use the environment variable %q.", TokenEnvName),
				MarkdownDescription: fmt.Sprintf("The API token for accessing Virtomize UII. If none is provided, the fallback is to use the environment variable `%s`.", TokenEnvName),
			},

			"localstorage": schema.StringAttribute{
				Optional:    true,
				Description: "The provider will store some data locally to work correctly. Use this parameter to overwrite the default location.",
			},
		},
	}
}

// Configure prepares a Virtomize UII API client for data sources and resources.
func (p *uiiProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config uiiProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// token
	token := config.APIToken.ValueString()
	if config.APIToken.IsUnknown() || config.APIToken.IsNull() {
		token = os.Getenv(TokenEnvName)
	}

	if token == "" {
		resp.Diagnostics.AddError(
			"No token provided to create Virtomize client",
			"Unable to retrieve token for authenticated Virtomize client. When using environment variables use "+TokenEnvName)
		return
	}

	// local storage
	localPath := ""
	if config.LocalStorage.IsUnknown() || config.LocalStorage.IsNull() {
		// not explicitly provided
		envStorageLocation := os.Getenv(StorageEnvName)
		if envStorageLocation != "" {
			localPath = envStorageLocation
		} else {
			localPath = createDefaultStoragePath()
		}
	} else {
		localPath = config.LocalStorage.ValueString()
	}

	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		resp.Diagnostics.AddError(
			"Folder does not exist",
			"Local folder does not exist")
	}

	c, err := client.NewClient(token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Virtomize client",
			"Unable to authenticate user for authenticated Virtomize client")

		return
	}

	client := &clientWithStorage{VirtomizeClient: c, StorageFolder: localPath, TimeProvider: defaultTimeProvider{}}

	// Make the client available during DataSource and Resource
	// type Configure methods.
	resp.ResourceData = client
	resp.DataSourceData = client
}

// DataSources defines the data sources implemented in the provider.
func (p *uiiProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

// Resources defines the resources implemented in the provider.
func (p *uiiProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewIsoResource,
	}
}

func createDefaultStoragePath() string {
	defaultStoragePath := path.Join(os.TempDir(), "uiiterraform")
	_ = os.Mkdir(defaultStoragePath, os.ModePerm)
	return defaultStoragePath
}
