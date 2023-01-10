package main

import (
	"context"
	"uii-terraform-framework-provider/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Provider documentation generation.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name virtomize-uii

func main() {
	providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		// TODO: Update this string with the published name of your provider.
		Address: "virtomize.com/uii/virtomize",
	})
}
