package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

const isoNameKey = "name"

const distributionKey = "distribution"
const versionKey = "version"
const hostnameKey = "hostname"

const localPathKey = "localpath"
const lastUpdatedKey = "last_updated"

const networksKey = "networks"
const dhcpKey = "dhcp"
const domainKey = "domain"
const macKey = "mac"
const ipNetKey = "ip_net"
const gatewayKey = "gateway"
const dnsKey = "dns"
const noInternetKey = "no_internet"

const passwordKey = "password"
const keyboardKey = "keyboard"
const timezoneKey = "timezone"
const packagesKey = "packages"
const architectureKey = "architecture"
const enableSshPasswordAuthenticationKey = "enable_ssh_authentication_through_password"
const sshKeysKey = "ssh_keys"
const localeKey = "locale"

// orderResourceModel maps the resource schema data.
type isoResourceModel struct {
	Id                       types.String    `tfsdk:"id"`
	LastUpdated              types.String    `tfsdk:"last_updated"`
	LocalPath                types.String    `tfsdk:"localpath"`
	Name                     types.String    `tfsdk:"name"`
	Distribution             types.String    `tfsdk:"distribution"`
	Version                  types.String    `tfsdk:"version"`
	Architecture             types.String    `tfsdk:"architecture"`
	Hostname                 types.String    `tfsdk:"hostname"`
	Locale                   types.String    `tfsdk:"locale"`
	Keyboard                 types.String    `tfsdk:"keyboard"`
	Password                 types.String    `tfsdk:"password"`
	ShhTroughPasswordEnabled types.Bool      `tfsdk:"enable_ssh_authentication_through_password"`
	SshKeys                  []types.String  `tfsdk:"ssh_keys"`
	Timezone                 types.String    `tfsdk:"timezone"`
	Packages                 []types.String  `tfsdk:"packages"`
	Networks                 []networksModel `tfsdk:"networks"`
}

// orderItemCoffeeModel maps coffee order item data.
type networksModel struct {
	Dhcp       types.Bool     `tfsdk:"dhcp"`
	Domain     types.String   `tfsdk:"domain"`
	Mac        types.String   `tfsdk:"mac"`
	IP         types.String   `tfsdk:"ip_net"`
	Gateway    types.String   `tfsdk:"gateway"`
	DNS        []types.String `tfsdk:"dns"`
	NoInternet types.Bool     `tfsdk:"no_internet"`
}

// Schema defines the schema for the resource.
func (r *IsoResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},

			isoNameKey: schema.StringAttribute{
				Required: true,
			},
			distributionKey: schema.StringAttribute{
				Required: true,
			},
			versionKey: schema.StringAttribute{
				Description: "the version of the distribution",
				Required:    true,
			},
			hostnameKey: schema.StringAttribute{
				Required: true,
			},
			networksKey: schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						dhcpKey: schema.BoolAttribute{
							Required: true,
						},
						domainKey: schema.StringAttribute{

							Optional: true,
						},
						macKey: schema.StringAttribute{

							Optional: true,
						},
						ipNetKey: schema.StringAttribute{
							Optional: true,
						},
						gatewayKey: schema.StringAttribute{
							Optional: true,
						},
						dnsKey: schema.ListAttribute{
							ElementType: types.StringType,
							Optional:    true,
						},
						noInternetKey: schema.BoolAttribute{
							Required: true,
						},
					},
				},
			},
			localPathKey: schema.StringAttribute{
				Computed: true,
			},

			// Optional parameters
			localeKey: schema.StringAttribute{
				Optional: true,
			},
			keyboardKey: schema.StringAttribute{
				Optional: true,
			},
			passwordKey: schema.StringAttribute{
				Optional: true,
			},
			enableSshPasswordAuthenticationKey: schema.BoolAttribute{
				Optional: true,
			},
			sshKeysKey: schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},

			timezoneKey: schema.StringAttribute{
				Optional: true,
			},
			architectureKey: schema.StringAttribute{
				Optional: true,
			},
			packagesKey: schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
		},
	}
}
