package provider

import (
	"fmt"
	client "github.com/Virtomize/uii-go-api"
	"golang.org/x/text/language"
	"net"
	"regexp"
	"strings"
	"time"
)

func validateIso(iso Iso, distributions []client.OS) []error {
	var result []error

	{
		hostNameError := validateHostname(iso.HostName)
		if hostNameError != nil {
			result = append(result, hostNameError)
		}
	}

	{
		err := validateDistribution(iso.Distribution, iso.Version, iso.Optionals.Arch, distributions)
		if err != nil {
			result = append(result, err)
		}
	}

	langErr := validateKeyboard(iso.Optionals.Keyboard)
	if langErr != nil {
		result = append(result, langErr)
	}

	localeErr := validateLocale(iso.Optionals.Locale)
	if localeErr != nil {
		result = append(result, localeErr)
	}

	timeErr := validateTimezone(iso.Optionals.Timezone)
	if localeErr != nil {
		result = append(result, timeErr)
	}

	{
		internet := false
		for _, n := range iso.Networks {
			internet = internet || !n.NoInternet

			if !n.DHCP {
				ipError := validateCIDR(n.IPNet)
				if ipError != nil {
					result = append(result, ipError)
				}

				if n.MAC != "" {
					_, macErr := net.ParseMAC(n.MAC)
					if macErr != nil {
						result = append(result, macErr)
					}
				}

				if n.Gateway == "" {
					gatewayErr := fmt.Errorf("static network configuration - no gateway defined set gateway parameter e.g 'gateway=192.168.0.1'")
					result = append(result, gatewayErr)
				} else {
					gwIP := net.ParseIP(n.Gateway)
					if gwIP == nil {
						gatewayErr := fmt.Errorf("static network configuration - gateway ip %v is invalid", n.Gateway)
						result = append(result, gatewayErr)
					}

					_, ipNet, _ := net.ParseCIDR(n.IPNet)
					if !ipNet.Contains(gwIP) {
						gatewayErr := fmt.Errorf("static network configuration - the gateway is not part of your defined subnet, this error occures if e.g. ipnet=192.168.0.20/24 and gateway=192.168.1.1 since your defined subnet does not include the gateway ip")
						result = append(result, gatewayErr)
					}
				}

				if len(n.DNS) > 0 {
					for _, ip := range n.DNS {
						if net.ParseIP(ip) == nil {
							dnsErr := fmt.Errorf("static network configuration - dns ip %v is invalid", ip)
							result = append(result, dnsErr)
						}
					}
				}
			}
		}

		if !internet {
			result = append(result, fmt.Errorf("ISO needs at least 1 configured with internet access"))
		}
	}

	return result
}

func validateCIDR(value string) error {
	_, _, err := net.ParseCIDR(value)

	if err != nil {
		return fmt.Errorf("%s requires an CIDR (example: 192.168. 129.23/17), error: %s current value: %s",
			ipNetKey,
			err.Error(),
			value)
	}
	return nil
}

func validateKeyboard(keyboard string) error {
	if keyboard == "" {
		return nil
	}

	_, err := language.Parse(keyboard)
	if err != nil {
		return fmt.Errorf("%s requires a valid keyboard layout or empty string (\"en-en\"), error: %s current value: %s",
			keyboardKey,
			err.Error(),
			keyboard)
	}

	return nil
}

func validateLocale(locale string) error {
	if locale == "" {
		return nil
	}

	_, err := language.Parse(locale)
	if err != nil {
		return fmt.Errorf("%s requires a valid local or empty string (\"en-en\"), error: %s current value: %s",
			localeKey,
			err.Error(),
			locale)
	}

	return nil
}

func validateTimezone(timeZone string) error {
	if timeZone == "" {
		return nil
	}

	_, err := time.LoadLocation(timeZone)
	if err != nil {
		return fmt.Errorf("%s requires a time zone or empty string, error: %s current value: %s",
			timezoneKey,
			err.Error(),
			timeZone)
	}

	return nil
}

func validateHostname(hostname string) error {
	reg := regexp.MustCompile("^([a-zA-Z0-9])+([a-zA-Z0-9\\-])*$")
	match := reg.MatchString(hostname)

	if !match {
		return fmt.Errorf("%s requires a valid hostname, allowed characters are a-z,0-9 and -, current value: %s",
			hostnameKey,
			hostname)
	}

	return nil
}

func validateDistribution(distribution string, version string, architecture string, distributions []client.OS) error {
	if len(distributions) == 0 {
		return nil
	}

	foundDistribution := false
	var displayNames []string
	for _, d := range distributions {
		displayNames = append(displayNames, d.DisplayName)
		if d.Distribution == distribution {
			foundDistribution = true
			break
		}
	}

	if !foundDistribution {
		return fmt.Errorf("%s requires a supported distribution, supported are: %s; current value: %s",
			distributionKey,
			strings.Join(displayNames, ", "),
			distribution)
	}

	foundVersion := false
	for _, d := range distributions {
		if d.Distribution == distribution {
			displayNames = append(displayNames, d.DisplayName)
			if d.Version == version {
				foundVersion = true
				break
			}
		}
	}

	if !foundVersion {
		return fmt.Errorf("%s requires a supported distribution version, supported are: %s; current value: %s",
			distributionKey,
			strings.Join(displayNames, ", "),
			version)
	}

	if architecture == "" {
		// architecture is optional
		return nil
	}

	for _, d := range distributions {
		if d.Distribution == distribution && d.Version == version && d.Architecture == architecture {
			return nil
		}
	}

	return fmt.Errorf("%s requires a supported distribution, supported are: %s; current value: %s",
		distributionKey,
		strings.Join(displayNames, ", "),
		architecture)
}
