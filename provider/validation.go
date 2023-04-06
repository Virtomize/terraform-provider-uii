package provider

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	client "github.com/Virtomize/uii-go-api"
	"golang.org/x/text/language"
)

var (
	ErrDistributionRequired        = errors.New("supported distribution required")
	ErrDistributionVersionRequired = errors.New("supported distribution version required")
	ErrInvalidHostname             = errors.New("valid hostname required, allowd characters are a-z,0-9 and -")
	ErrTimeZoneRequired            = errors.New("time zone or empty string required")
	ErrLocaleRequired              = errors.New("valid BCP 47 locale or empty string required, e.g: (\"en-GB\")")
	ErrKeyboardLayoutRequired      = errors.New("valid keyboard layout or empty string required, e.g: (\"en-GB\")")
	ErrCIDRRequired                = errors.New("CIDR required e.g (\"192.168.129.23/17\")")
	ErrNoInternet                  = errors.New("at least one network is required to provide internet access")
	ErrStaticNetworkConfiguration  = errors.New("static network configuration error")
	ErrStaticNetworkGatewaySubnet  = errors.New("static network configuration error: the gateway is not part of your defined subnet, this error occures if e.g. ipnet=192.168.0.20/24 and gateway=192.168.1.1 since your defined subnet does not include the gateway ip")
	ErrStaticNetworkGatewayIP      = errors.New("static network configuration error: invalid gateway ip")
	ErrStaticNetworkNoGateway      = errors.New("static network configuration error: no gateway defined, set gateway parameter to e.g 'gateway=192.168.0.1'")
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

			//nolint: nestif // fine
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
					result = append(result, ErrStaticNetworkNoGateway)
				} else {
					gwIP := net.ParseIP(n.Gateway)
					if gwIP == nil {
						gatewayErr := fmt.Errorf("%w:  %v is invalid", ErrStaticNetworkGatewayIP, n.Gateway)
						result = append(result, gatewayErr)
					}

					_, ipNet, _ := net.ParseCIDR(n.IPNet)
					if !ipNet.Contains(gwIP) {
						result = append(result, ErrStaticNetworkGatewaySubnet)
					}
				}

				if len(n.DNS) > 0 {
					for _, ip := range n.DNS {
						if net.ParseIP(ip) == nil {
							dnsErr := fmt.Errorf("%w: dns ip %v is invalid", ErrStaticNetworkConfiguration, ip)
							result = append(result, dnsErr)
						}
					}
				}
			}
		}

		if !internet {
			result = append(result, ErrNoInternet)
		}
	}

	return result
}

func validateCIDR(value string) error {
	_, _, err := net.ParseCIDR(value)

	if err != nil {
		return fmt.Errorf("%w for %s, error: %w current value: %s",
			ErrCIDRRequired,
			ipNetKey,
			err,
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
		return fmt.Errorf("%w for %s, error: %w current value: %s",
			ErrKeyboardLayoutRequired,
			keyboardKey,
			err,
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
		return fmt.Errorf("%w for %s, error: %w current value: %s",
			ErrLocaleRequired,
			localeKey,
			err,
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
		return fmt.Errorf("%w for %s, error: %w current value: %s",
			ErrTimeZoneRequired,
			timezoneKey,
			err,
			timeZone)
	}

	return nil
}

func validateHostname(hostname string) error {
	reg := regexp.MustCompile(`^([a-zA-Z0-9])+([a-zA-Z0-9\\-])*$`)
	match := reg.MatchString(hostname)

	if !match {
		return fmt.Errorf("%w for %s, current value: %s",
			ErrInvalidHostname,
			hostnameKey,
			hostname)
	}

	return nil
}

func validateDistribution(distribution, version, architecture string, distributions []client.OS) error {
	if len(distributions) == 0 {
		return nil
	}

	foundDistribution := false
	displayNames := []string{}
	for _, d := range distributions {
		displayNames = append(displayNames, d.DisplayName)
		if d.Distribution == distribution {
			foundDistribution = true
			break
		}
	}

	if !foundDistribution {
		return fmt.Errorf("%w for %s, supported are: %s; current value: %s", ErrDistributionRequired, distributionKey, strings.Join(displayNames, ", "), architecture)
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
		return fmt.Errorf("%w for %s, supported are: %s; current value: %s", ErrDistributionVersionRequired, distributionKey, strings.Join(displayNames, ", "), version)
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

	return fmt.Errorf("%w for %s, supported are: %s; current value: %s", ErrDistributionRequired, distributionKey, strings.Join(displayNames, ", "), architecture)
}
