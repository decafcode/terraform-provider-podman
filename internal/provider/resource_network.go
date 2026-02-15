package provider

import (
	"context"

	"github.com/decafcode/terraform-provider-podman/internal/api"
	"github.com/decafcode/terraform-provider-podman/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func newNetworkResource() resource.Resource {
	return &networkResource{}
}

type networkResource struct {
	resourceBase
}

type networkResourceModel struct {
	ContainerHost types.String `tfsdk:"container_host"`
	DnsEnabled    types.Bool   `tfsdk:"dns_enabled"`
	Id            types.String `tfsdk:"id"`
	Internal      types.Bool   `tfsdk:"internal"`
	Ipv6Enabled   types.Bool   `tfsdk:"ipv6_enabled"`
	Name          types.String `tfsdk:"name"`
}

func (r *networkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (r *networkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Podman network resource",

		Attributes: map[string]schema.Attribute{
			"container_host": schema.StringAttribute{
				MarkdownDescription: "URL of the container host where this resource resides",
				Optional:            true,
			},
			"dns_enabled": schema.BoolAttribute{
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether to enable resolution of private container IPs by container name inside the network. Defaults to false.",
				Optional:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Network ID assigned by the container runtime",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"internal": schema.BoolAttribute{
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Set to true to block all outbound traffic from this network. Containers will not be able to use this network to communicate with any peers outside of this network (incoming connections on published ports are unaffected). Defaults to false",
				Optional:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"ipv6_enabled": schema.BoolAttribute{
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Optional:            true,
				MarkdownDescription: "Enable IPv6 on this network in addition to IPv4. Defaults to false.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Network name. Must be unique on the container host.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
				Required: true,
			},
		},
	}
}

func (r *networkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data networkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	c, err := r.ps.getClient(ctx, data.ContainerHost.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Connection Error", err.Error())

		return
	}

	in := &api.NetworkJson{
		DnsEnabled:  data.DnsEnabled.ValueBool(),
		Internal:    data.Internal.ValueBool(),
		Ipv6Enabled: data.Ipv6Enabled.ValueBool(),
		Name:        data.Name.ValueString(),
	}

	out, err := c.NetworkCreate(ctx, in)

	if err != nil {
		resp.Diagnostics.AddError("Request failed", err.Error())

		return
	}

	data.Id = types.StringValue(out.Id)
	tflog.Trace(ctx, "Network created", map[string]any{"id": data.Id})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *networkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data networkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	c, err := r.ps.getClient(ctx, data.ContainerHost.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Connection Error", err.Error())

		return
	}

	json, err := c.NetworkInspect(ctx, data.Id.ValueString())

	if err != nil {
		status, ok := err.(client.StatusCodeError)

		if ok && status.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Error inspecting network", err.Error())
		}

		return
	}

	data.DnsEnabled = types.BoolValue(json.DnsEnabled)
	data.Internal = types.BoolValue(json.Internal)
	data.Ipv6Enabled = types.BoolValue(json.Ipv6Enabled)
	data.Name = types.StringValue(json.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *networkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Resource is immutable", "Resource is immutable")
}

func (r *networkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data networkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	c, err := r.ps.getClient(ctx, data.ContainerHost.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Connection Error", err.Error())

		return
	}

	err = c.NetworkDelete(ctx, data.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Delete failed", err.Error())

		return
	}
}

func (r *networkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importState(ctx, req, resp)
}
