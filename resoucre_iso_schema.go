package main

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

const isoNameKey = "name"

const distributionKey = "distribution"
const versionKey = "version"
const hostnameKey = "hostname"

const pathKey = "path"
const lastUpdatedKey = "last_updated"

const networksKey = "networks"
const dhcpKey = "dhcp"
const domainKey = "domain"
const macKey = "mac"
const ipNetKey = "ip_net"
const gatewayKey = "gateway"
const dnsKey = "dns"
const noInternetKey = "no_internet"

const passwordKey = "passwords"
const keyboardKey = "keyboard"
const timezoneKey = "timezone"
const packagesKey = "packages"
const architectureKey = "architecture"
const enableSshPasswordAuthenticationKey = "enable_ssh_authentication_through_password"
const sshKeysKey = "ssh_keys"
const localeKey = "locale"

func resourceIso() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIsoCreate,
		ReadContext:   resourceIsoRead,
		UpdateContext: resourceOrderUpdate,
		DeleteContext: resourceOrderDelete,
		Schema: map[string]*schema.Schema{
			isoNameKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			distributionKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			versionKey: {
				Type:        schema.TypeString,
				Description: "the version of the distribution",
				Required:    true,
			},
			hostnameKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			networksKey: {
				Type:     schema.TypeList,
				MinItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						dhcpKey: {
							Type:     schema.TypeBool,
							Required: true,
						},
						domainKey: {
							Type:     schema.TypeString,
							Optional: true,
						},
						macKey: {
							Type:     schema.TypeString,
							Optional: true,
						},
						ipNetKey: {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validateCIDR,
						},
						gatewayKey: {
							Type:     schema.TypeString,
							Optional: true,
						},
						dnsKey: {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
						},
						noInternetKey: {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			pathKey: {
				Type:     schema.TypeString,
				Computed: true,
			},
			lastUpdatedKey: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			// Optional parameters
			localeKey: {
				Type:     schema.TypeString,
				Optional: true,
			},
			keyboardKey: {
				Type:     schema.TypeString,
				Optional: true,
			},
			passwordKey: {
				Type:     schema.TypeString,
				Optional: true,
			},
			enableSshPasswordAuthenticationKey: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			sshKeysKey: {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},

			timezoneKey: {
				Type:     schema.TypeString,
				Optional: true,
			},
			architectureKey: {
				Type:     schema.TypeString,
				Optional: true,
			},
			packagesKey: {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
		},
	}
}
