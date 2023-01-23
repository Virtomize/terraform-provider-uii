package provider

import (
	"fmt"
	client "github.com/Virtomize/uii-go-api"
	"golang.org/x/text/language"
	"net"
	"regexp"
	"strings"
)

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
			strings.Join(displayNames, ","),
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
			strings.Join(displayNames, ","),
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
		strings.Join(displayNames, ","),
		architecture)
}
