package main

import (
	"context"
	"log"
	"terraform-provider-uii/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Provider documentation generation.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name virtomize-uii

func main() {
	err := providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/Virtomize/uii",
	})
	if err != nil {
		log.Fatal(err)
	}
}
