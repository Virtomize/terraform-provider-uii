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
		hasInternet := false
		for _, n := range iso.Networks {
			validationErrors := validateNetwork(n)
			if len(validationErrors) > 0 {
				result = append(result, validationErrors...)
			}
			hasInternet = hasInternet || n.NoInternet
		}

		if !hasInternet {
			result = append(result, fmt.Errorf("ISO needs at least 1 configured with internet access"))
		}
	}

	return result
}

func validateNetwork(n Network) []error {
	var errors []error
	if !n.DHCP {
		ipError := validateCIDR(n.IPNet)
		if ipError != nil {
			errors = append(errors, ipError)
		}

		if n.MAC != "" {
			_, macErr := net.ParseMAC(n.MAC)
			if macErr != nil {
				errors = append(errors, macErr)
			}
		}

		if n.Gateway == "" {
			gatewayErr := fmt.Errorf("static network configuration - no gateway defined set gateway parameter e.g 'gateway=192.168.0.1'")
			errors = append(errors, gatewayErr)
		} else {
			gwIP := net.ParseIP(n.Gateway)
			if gwIP == nil {
				gatewayErr := fmt.Errorf("static network configuration - gateway ip %v is invalid", n.Gateway)
				errors = append(errors, gatewayErr)
			}

			_, ipNet, _ := net.ParseCIDR(n.IPNet)
			if !ipNet.Contains(gwIP) {
				gatewayErr := fmt.Errorf("static network configuration - the gateway is not part of your defined subnet, this error occures if e.g. ipnet=192.168.0.20/24 and gateway=192.168.1.1 since your defined subnet does not include the gateway ip")
				errors = append(errors, gatewayErr)
			}
		}

		if len(n.DNS) > 0 {
			for _, ip := range n.DNS {
				if net.ParseIP(ip) == nil {
					dnsErr := fmt.Errorf("static network configuration - dns ip %v is invalid", ip)
					errors = append(errors, dnsErr)
				}
			}
		}
	}
	return errors
}

func validateCIDR(value string) error {
	_, _, err := net.ParseCIDR(value)

	if err != nil {
		//nolint: errorlint // can't have two errors
		return fmt.Errorf("%w for %s, error: %s current value: %s",
			ErrCIDRRequired,
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
		//nolint: errorlint // can't have two errors
		return fmt.Errorf("%w for %s, error: %s current value: %s",
			ErrKeyboardLayoutRequired,
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
		//nolint: errorlint // can't have two errors
		return fmt.Errorf("%w for %s, error: %s current value: %s",
			ErrLocaleRequired,
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
		//nolint: errorlint // can't have two errors
		return fmt.Errorf("%w for %s, error: %s current value: %s",
			ErrTimeZoneRequired,
			timezoneKey,
			err.Error(),
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
