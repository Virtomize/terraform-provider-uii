package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const isoNameKey = "name"

const distributionKey = "distribution"
const versionKey = "version"
const hostnameKey = "hostname"

const localPathKey = "localpath"

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

//nolint: gosec // wrong
const enableSSHPasswordAuthenticationKey = "enable_ssh_authentication_through_password"
const sshKeysKey = "ssh_keys"
const localeKey = "locale"

// orderResourceModel maps the resource schema data.
type isoResourceModel struct {
	ID                       types.String    `tfsdk:"id"`
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
	SSHKeys                  []types.String  `tfsdk:"ssh_keys"`
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
//nolint: funlen // nope
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
				Required:            true,
				Description:         "The distribution, for example \"debian\"",
				MarkdownDescription: "The distribution, for example `debian`",
			},
			versionKey: schema.StringAttribute{
				Description:         "The version of the distribution, for example \"11\"",
				MarkdownDescription: "The version of the distribution, for example `11`",
				Required:            true,
			},
			hostnameKey: schema.StringAttribute{
				Required:    true,
				Description: "The host name to be configured during the installation",
			},
			networksKey: schema.ListNestedAttribute{
				Required:    true,
				Description: "A list of networks that should be configured. Must contain at least one network with internet access.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						dhcpKey: schema.BoolAttribute{
							Required:    true,
							Description: "Specifies if the network should automatically retrieve its IP through and DHCP server.",
						},
						domainKey: schema.StringAttribute{
							Optional:    true,
							Description: "The domain used for this network. Necessary only for non DHCP networks.",
						},
						macKey: schema.StringAttribute{
							Optional:    true,
							Description: "The mac address of the network card this network configuration should be applied to. Only necessary if more then one card is present.",
						},
						ipNetKey: schema.StringAttribute{
							Optional:            true,
							Description:         "The CIDR for this network, for example \"198.51.100.0/22\", Necessary only for non DHCP networks.",
							MarkdownDescription: "The CIDR for this network, for example `198.51.100.0/22`, Necessary only for non DHCP networks.",
						},
						gatewayKey: schema.StringAttribute{
							Optional:    true,
							Description: "The gateway used for this network. Necessary only for non DHCP networks.",
						},
						dnsKey: schema.ListAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Description: "A list of DNS server for this network. Necessary only for non DHCP networks.",
						},
						noInternetKey: schema.BoolAttribute{
							Required:    true,
							Description: "A flag indicating if this network does not have access to internet. At least one network needs internet access.",
						},
					},
				},
			},
			localPathKey: schema.StringAttribute{
				Computed:    true,
				Description: "The path where the ISO is temporary cached after its creation.",
			},

			// Optional parameters
			localeKey: schema.StringAttribute{
				Optional:            true,
				Description:         "The locale used for the OS. For example \"en-en\". Defaults to English.",
				MarkdownDescription: "The locale used for the OS. For example `en-en`. Defaults to English.",
			},
			keyboardKey: schema.StringAttribute{
				Optional:            true,
				Description:         "The keyboard layout used for the OS. For example \"en-en\". Defaults to English.",
				MarkdownDescription: "The keyboard layout used for the OS. For example `en-en`. Defaults to English.",
			},
			passwordKey: schema.StringAttribute{
				Optional:            true,
				Description:         "A password to be set the \"root\" user. The default password if this parameter is not set is \"virtomize\".",
				MarkdownDescription: "A password to be set the `root` user. The default password if this parameter is not set is `virtomize`.",
			},
			enableSSHPasswordAuthenticationKey: schema.BoolAttribute{
				Optional:    true,
				Description: "If true, login into the OS through SSH will be enabled.",
			},
			sshKeysKey: schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "A list of SSH keys to be installed for use with the SSH login.",
			},

			timezoneKey: schema.StringAttribute{
				Optional:    true,
				Description: "The timezone to be used by the OS.",
			},
			architectureKey: schema.StringAttribute{
				Optional:            true,
				Description:         "The architecture variant of the OS that should be installed. \"32\" or \"64\". Defaults to 64.",
				MarkdownDescription: "The architecture variant of the OS that should be installed. `32` or `64`. Defaults to `64`.",
			},
			packagesKey: schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "A list of additional packages that should be installed in addition to the necessary ones.",
			},
		},
	}
}
