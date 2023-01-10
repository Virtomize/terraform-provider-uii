package main

import (
	"context"
	"uii-terraform-framework-provider/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		// TODO: Update this string with the published name of your provider.
		Address: "virtomize.com/uii/virtomize",
	})
}
