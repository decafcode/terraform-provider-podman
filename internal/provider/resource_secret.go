package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func newSecretResource() resource.Resource {
	return &secretResource{}
}

type secretResource struct {
	resourceBase
}

type secretResourceModel struct {
	ContainerHost types.String `tfsdk:"container_host"`
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Value         types.String `tfsdk:"value"`
	ValueVersion  types.Int32  `tfsdk:"value_version"`
}

func (r *secretResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (r *secretResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Podman Secret resource",

		Attributes: map[string]schema.Attribute{
			"container_host": schema.StringAttribute{
				MarkdownDescription: "URL of the container host where this resource resides",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Secret ID assigned by the container runtime",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Secret name. Must be unique on the container host.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Required: true,
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "Write-only secret contents. This does not get stored in " +
					"Terraform state.",
				Required:  true,
				Sensitive: true,
				WriteOnly: true,
			},
			"value_version": schema.Int32Attribute{
				MarkdownDescription: "An imaginary version number for the write-only secret " +
					"value. This number must be incremented every time the secret value is " +
					"updated, since Terraform does not store the secret value in its state " +
					"and therefore has no way of knowing if the secret value being supplied " +
					"to this resource has changed or not.\n\n" +
					"\tOn imported resources this attribute is set to 1.",
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
				Required: true,
			},
		},
	}
}

func (r *secretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data secretResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.ps.getClient(ctx, data.ContainerHost.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Connection Error", err.Error())

		return
	}

	name := data.Name.ValueString()
	value := data.Value.ValueString()
	out, err := client.SecretCreate(ctx, name, value)

	if err != nil {
		resp.Diagnostics.AddError("Request failed", err.Error())

		return
	}

	data.Id = types.StringValue(out.Id)
	tflog.Trace(ctx, "Secret created", map[string]any{"id": data.Id})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *secretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data secretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	c, err := r.ps.getClient(ctx, data.ContainerHost.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Connection Error", err.Error())

		return
	}

	json, err := c.SecretInspect(ctx, data.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Error fetching secret", err.Error())

		return
	}

	data.Name = types.StringValue(json.Spec.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *secretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Resource is immutable", "Resource is immutable")
}

func (r *secretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data secretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	c, err := r.ps.getClient(ctx, data.ContainerHost.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Connection Error", err.Error())

		return
	}

	err = c.SecretDelete(ctx, data.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Delete failed", err.Error())

		return
	}
}

func (r *secretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importState(ctx, req, resp)
	resp.State.SetAttribute(ctx, path.Root("value_version"), 1)
}
