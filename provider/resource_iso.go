package provider

import (
	"context"
	"fmt"
	"time"

	client "github.com/Virtomize/uii-go-api"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/tredoe/osutil/user/crypt/sha512_crypt"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &IsoResource{}
)

const (
	errClientInit     = "Client not properly initialised"
	errClientInitDesc = "The UII client is not properly initialised, which lead to an internal error."
)

// NewIsoResource is a helper function to simplify the provider implementation.
func NewIsoResource() resource.Resource {
	return &IsoResource{}
}

// IsoResource is the resource implementation.
type IsoResource struct {
	client *clientWithStorage
}

// Metadata returns the resource type name.
func (r *IsoResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	// this is howe this resource is called in the .tf file ex:
	//		resource "virtomize_iso" "name_of+this_explicit_iso" { ... }
	resp.TypeName = req.ProviderTypeName + "_iso"
}

func (r *IsoResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*clientWithStorage)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *clientWithStorage, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = c
}

// Create creates the resource and sets the initial Terraform state.
func (r *IsoResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan isoResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError(errClientInit, errClientInitDesc)
		return
	}

	distributions, err := r.client.ReadDistributions()
	if err != nil {
		// fallback to allowing everything, to support terraform plan for users that have not created an api key yet
		// not sure about this
		distributions = []client.OS{}
	}

	iso := parseIsoFromResourceModel(plan)

	errors := validateIso(iso, distributions)
	if errors != nil {
		for _, e := range errors {
			resp.Diagnostics.AddError("Error validating iso", e.Error())
		}
		return
	}

	storedIso, err := r.client.CreateIso(iso)
	if err != nil {
		resp.Diagnostics.AddError("Error creating iso", err.Error())
		return
	}

	// set computed values
	plan.ID = types.StringValue(storedIso.ID)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	plan.LocalPath = types.StringValue(storedIso.LocalPath)

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r IsoResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data isoResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// fallback to allowing everything, as no client is present during the plan phase
	var distributions []client.OS

	iso := parseIsoFromResourceModel(data)

	errors := validateIso(iso, distributions)

	for _, e := range errors {
		resp.Diagnostics.AddError(
			"Validation error",
			e.Error(),
		)
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *IsoResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state isoResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError(errClientInit, errClientInitDesc)
		return
	}

	iso, err := r.client.ReadIso(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading ISO from storage",
			"Could not read ISO Id "+state.ID.ValueString()+". Try removing it from state file to force regeneration. Error was: "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	setIsoToModel(iso, &state)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *IsoResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan isoResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError(errClientInit, errClientInitDesc)
		return
	}

	iso := parseIsoFromResourceModel(plan)

	isoID := plan.ID.ValueString()
	err := r.client.UpdateIso(isoID, iso)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Iso",
			"Could not update ISO :"+err.Error(),
		)
		return
	}

	// read updated iso to retrieve recomputed values
	updatedIso, err := r.client.ReadIso(isoID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Iso",
			"Could not read ISO Id "+isoID+": "+err.Error(),
		)
		return
	}

	// set computed values
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	plan.LocalPath = types.StringValue(updatedIso.LocalPath)

	// Update resource state with updated items and timestamp
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *IsoResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state isoResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError(errClientInit, errClientInitDesc)
		return
	}

	err := r.client.DeleteIso(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting ISO",
			"Could not delete iso, unexpected error: "+err.Error(),
		)
		return
	}
}

func parseIsoFromResourceModel(d isoResourceModel) Iso {
	name := d.Name.ValueString()
	distribution := d.Distribution.ValueString()
	version := d.Version.ValueString()
	hostname := d.Hostname.ValueString()

	locale := stringOrDefault(d.Locale, "")
	keyboard := stringOrDefault(d.Keyboard, "")
	password := stringOrDefault(d.Password, "")
	shhPasswordAuth := boolOrDefault(d.ShhTroughPasswordEnabled, false)
	timezone := stringOrDefault(d.Timezone, "")
	architecture := stringOrDefault(d.Architecture, "")

	networks := parseNetworksFromSchema(d.Networks)

	iso := Iso{
		Name:         name,
		Distribution: distribution,
		Version:      version,
		HostName:     hostname,
		Networks:     networks,
		Optionals: BuildOpts{
			Locale:          locale,
			Keyboard:        keyboard,
			Password:        hashPassword(password),
			SSHPasswordAuth: shhPasswordAuth,
			SSHKeys:         nil,
			Timezone:        timezone,
			Arch:            architecture,
			Packages:        nil,
		},
	}
	return iso
}

func parseNetworksFromSchema(networksModel []networksModel) []Network {
	var networks []Network
	for _, item := range networksModel {
		dhcp := item.Dhcp.ValueBool()
		noInternet := item.NoInternet.ValueBool()

		if dhcp {
			networks = append(networks, Network{
				DHCP:       dhcp,
				NoInternet: noInternet,
			})
		} else {
			domain := stringOrDefault(item.Domain, "")
			mac := stringOrDefault(item.Mac, "")
			ipNet := stringOrDefault(item.IP, "")
			gateway := stringOrDefault(item.Gateway, "")
			dns := stringListWithValidElements(item.DNS)

			network := Network{
				DHCP:       dhcp,
				Domain:     domain,
				MAC:        mac,
				IPNet:      ipNet,
				Gateway:    gateway,
				DNS:        dns,
				NoInternet: noInternet,
			}

			networks = append(networks, network)
		}
	}

	return networks
}

func stringListWithValidElements(list []types.String) []string {
	result := []string{}

	for _, item := range list {
		if item.IsUnknown() {
			continue
		}

		if item.IsNull() {
			continue
		}

		result = append(result, item.String())
	}

	return result
}

// nolint: unparam // not sure if this is always the case in the future
func stringOrDefault(data types.String, defaultValue string) string {
	if data.IsUnknown() {
		return defaultValue
	}

	if data.IsNull() {
		return defaultValue
	}

	return data.ValueString()
}

func boolOrDefault(data types.Bool, defaultValue bool) bool {
	if data.IsUnknown() {
		return defaultValue
	}

	if data.IsNull() {
		return defaultValue
	}

	return data.ValueBool()
}

func transformNetworksToModel(networks []Network) []networksModel {
	var result []networksModel
	for _, item := range networks {
		if item.DHCP {
			result = append(result,
				networksModel{
					Dhcp:       types.BoolValue(item.DHCP),
					NoInternet: types.BoolValue(item.NoInternet),
					Mac:        types.StringValue(item.MAC),
				},
			)
		} else {
			var dnss []types.String
			for _, dns := range item.DNS {
				dnss = append(dnss, types.StringValue(dns))
			}

			result = append(result,
				networksModel{
					Dhcp:       types.BoolValue(item.DHCP),
					Domain:     types.StringValue(item.Domain),
					Mac:        types.StringValue(item.MAC),
					IP:         types.StringValue(item.IPNet),
					Gateway:    types.StringValue(item.Gateway),
					DNS:        dnss,
					NoInternet: types.BoolValue(item.NoInternet),
				},
			)
		}
	}
	return result
}

func setIsoToModel(iso StoredIso, state *isoResourceModel) {
	state.Name = types.StringValue(iso.Name)
	state.Distribution = types.StringValue(iso.Distribution)
	state.Version = types.StringValue(iso.Version)
	state.Hostname = types.StringValue(iso.HostName)
	state.Networks = transformNetworksToModel(iso.Networks)
	state.LocalPath = types.StringValue(iso.LocalPath)
}

func hashPassword(password string) string {
	if password == "" {
		return ""
	}

	hashFactory := sha512_crypt.New()
	hash, err := hashFactory.Generate([]byte(password), []byte("$6$somesalt"))
	if err != nil {
		panic(err)
	}

	return string(hash)
}
