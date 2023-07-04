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
	ErrInvalidHostname             = errors.New("valid hostname required, allowed characters are -, a-z, and 0-9")
	ErrTimeZoneRequired            = errors.New("time zone or empty string required")
	ErrLocaleRequired              = errors.New("valid BCP 47 locale or empty string required, e.g: (\"en-GB\")")
	ErrKeyboardLayoutRequired      = errors.New("valid keyboard layout or empty string required, e.g: (\"en-GB\")")
	ErrCIDRRequired                = errors.New("CIDR required e.g (\"192.168.129.23/17\")")
	ErrNoInternet                  = errors.New("at least one network is required to provide internet access")
	ErrStaticNetworkConfiguration  = errors.New("static network configuration error")
	ErrStaticNetworkGatewaySubnet  = errors.New("static network configuration error: the gateway is not part of your defined subnet, this error occurs if e.g. ipnet=192.168.0.20/24 and gateway=192.168.1.1 since your defined subnet does not include the gateway ip")
	ErrStaticNetworkGatewayIP      = errors.New("static network configuration error: invalid gateway ip")
	ErrStaticNetworkNoGateway      = errors.New("static network configuration error: no gateway defined, set gateway parameter to e.g 'gateway=192.168.0.1'")
	ErrMissingMac                  = errors.New("missing MAC address needed for multi network configuration")
)

func validateIso(plan isoResourceModel, distributions []client.OS) []error {
	var result []error

	if !plan.Hostname.IsUnknown() {
		hostName := plan.Hostname.ValueString()
		hostNameError := validateHostname(hostName)
		if hostNameError != nil {
			result = append(result, hostNameError)
		}
	}

	if !plan.Distribution.IsUnknown() && !plan.Version.IsUnknown() {
		distribution := plan.Distribution.ValueString()
		version := plan.Version.ValueString()
		architecture := stringOrDefault(plan.Architecture, "")

		err := validateDistribution(distribution, version, architecture, distributions)
		if err != nil {
			result = append(result, err)
		}
	}

	keyboard := stringOrDefault(plan.Keyboard, "")
	langErr := validateKeyboard(keyboard)
	if langErr != nil {
		result = append(result, langErr)
	}

	locale := stringOrDefault(plan.Locale, "")
	localeErr := validateLocale(locale)
	if localeErr != nil {
		result = append(result, localeErr)
	}

	timezone := stringOrDefault(plan.Timezone, "")
	timeErr := validateTimezone(timezone)
	if localeErr != nil {
		result = append(result, timeErr)
	}

	{
		hasInternet := false
		for _, network := range plan.Networks {
			validationErrors := validateNetwork(network, len(plan.Networks) > 1)
			if len(validationErrors) > 0 {
				result = append(result, validationErrors...)
			}
			hasInternet = hasInternet || !network.NoInternet.ValueBool()
		}

		if !hasInternet {
			result = append(result, ErrNoInternet)
		}
	}

	return result
}

func validateNetwork(n networksModel, needsMac bool) []error {
	var errors []error

	mac := stringOrDefault(n.Mac, "")
	_, macErr := net.ParseMAC(mac)
	if !n.Mac.IsUnknown() && mac == "" {
		if macErr != nil {
			errors = append(errors, macErr)
		}
	}

	if needsMac && mac == "" {
		errors = append(errors, ErrMissingMac)
	}

	dhcp := n.Dhcp.ValueBool()
	if dhcp {
		// only validate MAC in DHCP networks
		return errors
	}

	if !n.IP.IsUnknown() {
		ipNet := stringOrDefault(n.IP, "")
		ipError := validateCIDR(ipNet)
		if ipError != nil {
			errors = append(errors, ipError)
		}
	}

	if !n.Gateway.IsUnknown() {
		gateway := stringOrDefault(n.Gateway, "")
		if gateway == "" {
			errors = append(errors, ErrStaticNetworkNoGateway)
		} else {
			gwIP := net.ParseIP(gateway)
			if gwIP == nil {
				gatewayErr := fmt.Errorf("static network configuration - gateway ip %s is invalid", n.Gateway)
				errors = append(errors, gatewayErr)
			}

			// check that gateway is in CIDR of IP
			if !n.IP.IsUnknown() {
				networkIp := stringOrDefault(n.IP, "")
				_, ipNet, _ := net.ParseCIDR(networkIp)
				if !ipNet.Contains(gwIP) {
					errors = append(errors, ErrStaticNetworkGatewaySubnet)
				}
			}
		}
	}

	if len(n.DNS) > 0 {
		for _, ipPlan := range n.DNS {
			if !ipPlan.IsUnknown() {
				ip := ipPlan.ValueString()
				if net.ParseIP(ip) == nil {
					dnsErr := fmt.Errorf("static network configuration - dns ip %s is invalid", ip)
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

	if keyboard == unknownString {
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

	if locale == unknownString {
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

	if timeZone == unknownString {
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
	var displayNames []string
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
