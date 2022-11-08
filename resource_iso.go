package main

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net"
	"time"
)

func resourceIsoCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*clientWithStorage)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	iso, err := parseIsoFromSchema(d)
	if err != nil {
		return diag.FromErr(err)
	}

	o, err := c.CreateIso(iso)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(o.Id)
	resourceIsoRead(ctx, d, m)

	return diags
}

func parseIsoFromSchema(d *schema.ResourceData) (Iso, error) {
	name := d.Get("name").(string)
	distribution := d.Get(distributionKey).(string)
	version := d.Get(versionKey).(string)
	hostname := d.Get(hostnameKey).(string)

	locale := dataValueOrDefault(d, localeKey, "").(string)
	keyboard := dataValueOrDefault(d, keyboardKey, "").(string)
	password := dataValueOrDefault(d, passwordKey, "").(string)
	shhPasswordAuth := dataValueOrDefault(d, enableSshPasswordAuthenticationKey, false).(bool)
	timezone := dataValueOrDefault(d, timezoneKey, "").(string)
	architecture := dataValueOrDefault(d, architectureKey, "").(string)

	networks, err := parseNetworksFromSchema(d)
	if err != nil {
		return Iso{}, err
	}

	iso := Iso{
		Name:         name,
		Distribution: distribution,
		Version:      version,
		HostName:     hostname,
		Networks:     networks,
		Optionals: BuildOpts{
			Locale:          locale,
			Keyboard:        keyboard,
			Password:        password,
			SSHPasswordAuth: shhPasswordAuth,
			SSHKeys:         nil,
			Timezone:        timezone,
			Arch:            architecture,
			Packages:        nil,
		},
	}
	return iso, nil
}

func parseNetworksFromSchema(n *schema.ResourceData) ([]Network, error) {
	items := n.Get(networksKey).([]interface{})
	var networks []Network
	hasOneNetworkWithInternet := false
	for _, item := range items {
		d := item.(map[string]interface{})

		dhcp := d[dhcpKey].(bool)
		noInternet := d[noInternetKey].(bool)
		hasOneNetworkWithInternet = hasOneNetworkWithInternet || !noInternet

		if dhcp {
			networks = append(networks, Network{
				DHCP:       dhcp,
				NoInternet: noInternet,
			})
		} else {
			domain := valueOrDefault(d, domainKey, "").(string)
			mac := valueOrDefault(d, macKey, "").(string)
			ipNet := valueOrDefault(d, ipNetKey, "").(string)
			gateway := valueOrDefault(d, gatewayKey, "").(string)
			dns := valueToStringListOrDefault(d, dnsKey, []string{})

			net := Network{
				DHCP:       dhcp,
				Domain:     domain,
				MAC:        mac,
				IPNet:      ipNet,
				Gateway:    gateway,
				DNS:        dns,
				NoInternet: noInternet,
			}

			networks = append(networks, net)
		}
	}

	if !hasOneNetworkWithInternet {
		return []Network{}, fmt.Errorf("iso needs at least 1 network configured for internet access")
	}

	return networks, nil
}

func valueOrDefault(dict map[string]interface{}, key string, defaultValue interface{}) interface{} {
	if val, ok := dict[key]; ok {
		return val
	}

	return defaultValue
}

func dataValueOrDefault(d *schema.ResourceData, key string, defaultValue interface{}) interface{} {
	v := d.Get(key)
	if v == nil {
		return defaultValue
	}
	return v
}

func valueToStringListOrDefault(dict map[string]interface{}, key string, defaultValue []string) []string {
	if val, ok := dict[key]; ok {
		var result []string
		list := val.([]interface{})
		for _, v := range list {
			result = append(result, v.(string))
		}

		return result
	}

	return defaultValue
}

func resourceIsoRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*clientWithStorage)

	isoId := d.Id()

	iso, err := c.ReadIso(isoId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set(isoNameKey, iso.Name)
	d.Set(distributionKey, iso.Distribution)
	d.Set(versionKey, iso.Version)
	d.Set(hostnameKey, iso.HostName)
	d.Set(pathKey, iso.LocalPath)
	d.Set(networksKey, flattenNetworks(iso.Networks))
	return diag.Diagnostics{}
}

func flattenNetworks(networks []Network) interface{} {
	if networks != nil {
		ois := make([]interface{}, len(networks), len(networks))

		for i, network := range networks {
			oi := make(map[string]interface{})

			oi[dhcpKey] = network.DHCP
			oi[domainKey] = network.Domain
			oi[macKey] = network.MAC
			oi[ipNetKey] = network.IPNet
			oi[gatewayKey] = network.Gateway
			oi[dnsKey] = network.DNS
			oi[noInternetKey] = network.NoInternet
			ois[i] = oi
		}

		return ois
	}

	return make([]interface{}, 0)
}

func resourceOrderUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*clientWithStorage)

	isoId := d.Id()

	if d.HasChange("name") || d.HasChange("os") {
		iso, err := parseIsoFromSchema(d)
		if err != nil {
			return diag.FromErr(err)
		}

		err = c.UpdateIso(isoId, iso)
		if err != nil {
			return diag.FromErr(err)
		}

		d.Set(lastUpdatedKey, time.Now().Format(time.RFC850))
	}

	return resourceIsoRead(ctx, d, m)
}

func resourceOrderDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*clientWithStorage)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	orderID := d.Id()

	err := c.DeleteIso(orderID)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}

func validateCIDR(v interface{}, path cty.Path) diag.Diagnostics {
	value := v.(string)
	_, _, err := net.ParseCIDR(value)

	var diags diag.Diagnostics
	if err != nil {
		diag := diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "provide CIDR",
			Detail:   fmt.Sprintf("ip_net requires and CIDR (example: 192.168. 129.23/17) %s", value),
		}
		diags = append(diags, diag)
	}
	return diags
}
