// Package vapp provides a Terraform resource.
package vapp

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	govcdtypes "github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/orange-cloudavenue/terraform-provider-cloudavenue/internal/client"
	"github.com/orange-cloudavenue/terraform-provider-cloudavenue/internal/helpers/boolpm"
	"github.com/orange-cloudavenue/terraform-provider-cloudavenue/internal/provider/common"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &orgNetworkResource{}
	_ resource.ResourceWithConfigure   = &orgNetworkResource{}
	_ resource.ResourceWithImportState = &orgNetworkResource{}
)

// NewOrgNetworkResource is a helper function to simplify the provider implementation.
func NewOrgNetworkResource() resource.Resource {
	return &orgNetworkResource{}
}

// orgNetworkResource is the resource implementation.
type orgNetworkResource struct {
	client *client.CloudAvenue
}

type orgNetworkResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	VAppName           types.String `tfsdk:"vapp_name"`
	VDC                types.String `tfsdk:"vdc"`
	NetworkName        types.String `tfsdk:"network_name"`
	IsFenced           types.Bool   `tfsdk:"is_fenced"`
	RetainIPMacEnabled types.Bool   `tfsdk:"retain_ip_mac_enabled"`
}

// Metadata returns the resource type name.
func (r *orgNetworkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + categoryName + "_org_network"
}

// Schema defines the schema for the resource.
func (r *orgNetworkResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides capability to attach an existing Org VDC Network to a vApp and toggle network features.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the org_network.",
			},
			"vapp_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vdc": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "The name of VDC to use, optional if defined at provider level.",
			},
			"network_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				MarkdownDescription: "Organization network name to which vApp network is connected to.",
			},
			"is_fenced": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolpm.SetDefault(false),
				},
				MarkdownDescription: "Fencing allows identical virtual machines in different vApp networks connect to organization VDC networks that are accessed in this vApp. Default is `false`.",
			},
			"retain_ip_mac_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolpm.SetDefault(false),
				},
				MarkdownDescription: "Specifies whether the network resources such as IP/MAC of router will be retained across deployments. Default is `false`.",
			},
		},
	}
}

func (r *orgNetworkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.CloudAvenue)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.CloudAvenue, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *orgNetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var (
		plan *orgNetworkResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgNetworkRef, errInitOrg := plan.initOrgNetworkQuery(ctx, r.client, true)
	if errInitOrg != nil {
		resp.Diagnostics.AddError(errInitOrg.Summary, errInitOrg.Detail)
		return
	}

	if orgNetworkRef.VAppLocked {
		defer orgNetworkRef.VAppUnlockF()
	}

	orgNetworkName := plan.NetworkName.ValueString()
	orgNetwork, err := orgNetworkRef.VDC.GetOrgVdcNetworkByNameOrId(orgNetworkName, true)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving org network", err.Error())
		return
	}

	retainIPMac := plan.RetainIPMacEnabled.ValueBool()
	isFenced := plan.IsFenced.ValueBool()

	vappNetworkSettings := &govcd.VappNetworkSettings{RetainIpMacEnabled: &retainIPMac}

	vAppNetworkConfig, err := orgNetworkRef.VApp.AddOrgNetwork(vappNetworkSettings, orgNetwork.OrgVDCNetwork, isFenced)
	if err != nil {
		resp.Diagnostics.AddError("Error creating vApp network", err.Error())
		return
	}

	vAppNetwork := govcdtypes.VAppNetworkConfiguration{}
	for _, networkConfig := range vAppNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == orgNetwork.OrgVDCNetwork.Name {
			vAppNetwork = networkConfig
		}
	}

	if vAppNetwork == (govcdtypes.VAppNetworkConfiguration{}) {
		resp.Diagnostics.AddError("Error creating vApp network", "vApp network not found in vApp network config")
		return
	}

	networkID, err := govcd.GetUuidFromHref(vAppNetwork.Link.HREF, false)
	if err != nil {
		resp.Diagnostics.AddError("Error creating vApp network uuid", err.Error())
		return
	}

	id := common.NormalizeID("urn:vcloud:network:", networkID)

	plan = &orgNetworkResourceModel{
		ID:                 types.StringValue(id),
		VAppName:           plan.VAppName,
		VDC:                plan.VDC,
		NetworkName:        plan.NetworkName,
		IsFenced:           plan.IsFenced,
		RetainIPMacEnabled: plan.RetainIPMacEnabled,
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *orgNetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *orgNetworkResourceModel

	// Get current state
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgNetworkRef, errInitOrg := state.initOrgNetworkQuery(ctx, r.client, false)
	if errInitOrg != nil {
		if errInitOrg.Summary == ErrVAppNotFound {
			resp.State.RemoveResource(ctx)
		}
		resp.Diagnostics.AddError(errInitOrg.Summary, errInitOrg.Detail)
		return
	}

	if orgNetworkRef.VAppLocked {
		defer orgNetworkRef.VAppUnlockF()
	}

	vAppNetworkConfig, err := orgNetworkRef.VApp.GetNetworkConfig()
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving vApp network config", err.Error())
		return
	}

	vAppNetwork, networkID, errFindNetwork := state.findOrgNetwork(vAppNetworkConfig)
	if errFindNetwork != nil {
		resp.Diagnostics.AddError(errFindNetwork.Summary, errFindNetwork.Detail)
		return
	}

	if vAppNetwork == (&govcdtypes.VAppNetworkConfiguration{}) {
		resp.State.RemoveResource(ctx)
		return
	}

	id := common.NormalizeID("urn:vcloud:network:", *networkID)
	isFenced := false
	if vAppNetwork.Configuration.FenceMode == govcdtypes.FenceModeNAT {
		isFenced = true
	}

	plan := &orgNetworkResourceModel{
		ID:                 types.StringValue(id),
		VAppName:           state.VAppName,
		VDC:                state.VDC,
		NetworkName:        state.NetworkName,
		IsFenced:           types.BoolValue(isFenced),
		RetainIPMacEnabled: types.BoolValue(*vAppNetwork.Configuration.RetainNetInfoAcrossDeployments),
	}

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *orgNetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state *orgNetworkResourceModel

	// Get current state
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgNetworkRef, errInitOrg := plan.initOrgNetworkQuery(ctx, r.client, true)
	if errInitOrg != nil {
		resp.Diagnostics.AddError(errInitOrg.Summary, errInitOrg.Detail)
		return
	}

	if orgNetworkRef.VAppLocked {
		defer orgNetworkRef.VAppUnlockF()
	}

	vAppNetworkConfig, err := orgNetworkRef.VApp.GetNetworkConfig()
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving vApp network config", err.Error())
		return
	}

	vAppNetwork, _, errFindNetwork := plan.findOrgNetwork(vAppNetworkConfig)
	if errFindNetwork != nil {
		resp.Diagnostics.AddError(errFindNetwork.Summary, errFindNetwork.Detail)
		return
	}

	if vAppNetwork == (&govcdtypes.VAppNetworkConfiguration{}) {
		resp.State.RemoveResource(ctx)
		return
	}

	isFenced := false
	if vAppNetwork.Configuration.FenceMode == govcdtypes.FenceModeNAT {
		isFenced = true
	}

	if plan.IsFenced.ValueBool() != isFenced || plan.RetainIPMacEnabled.ValueBool() != *vAppNetwork.Configuration.RetainNetInfoAcrossDeployments {
		tflog.Debug(ctx, "updating vApp network")
		retainIP := plan.RetainIPMacEnabled.ValueBool()
		vappNetworkSettings := &govcd.VappNetworkSettings{
			ID:                 state.ID.ValueString(),
			RetainIpMacEnabled: &retainIP,
		}
		_, err = orgNetworkRef.VApp.UpdateOrgNetwork(vappNetworkSettings, plan.IsFenced.ValueBool())
		if err != nil {
			resp.Diagnostics.AddError("Error updating vApp network", err.Error())
			return
		}
	}

	plan = &orgNetworkResourceModel{
		ID:                 state.ID,
		VAppName:           state.VAppName,
		VDC:                state.VDC,
		NetworkName:        state.NetworkName,
		IsFenced:           plan.IsFenced,
		RetainIPMacEnabled: plan.RetainIPMacEnabled,
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *orgNetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *orgNetworkResourceModel

	// Get current state
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgNetworkRef, errInitOrg := state.initOrgNetworkQuery(ctx, r.client, true)
	if errInitOrg != nil {
		resp.Diagnostics.AddError(errInitOrg.Summary, errInitOrg.Detail)
		return
	}

	if orgNetworkRef.VAppLocked {
		defer orgNetworkRef.VAppUnlockF()
	}
	_, err := orgNetworkRef.VApp.RemoveNetwork(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting vApp network", err.Error())
		return
	}
}

func (r *orgNetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state *orgNetworkResourceModel
	resourceURI := strings.Split(req.ID, ".")

	if len(resourceURI) != 3 && len(resourceURI) != 2 {
		resp.Diagnostics.AddError("Error importing org_network", "Wrong resource URI format. Expected vdc.vapp.org_network_name or vapp.org_network_name")
		return
	}

	state = &orgNetworkResourceModel{
		VAppName:    types.StringValue(resourceURI[0]),
		NetworkName: types.StringValue(resourceURI[1]),
	}

	if len(resourceURI) == 4 {
		state = &orgNetworkResourceModel{
			VDC:         types.StringValue(resourceURI[0]),
			VAppName:    types.StringValue(resourceURI[1]),
			NetworkName: types.StringValue(resourceURI[2]),
		}
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}