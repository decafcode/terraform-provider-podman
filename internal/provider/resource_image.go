package provider

import (
	"context"

	"github.com/decafcode/terraform-provider-podman/internal/api"
	"github.com/decafcode/terraform-provider-podman/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func newImageResource() resource.Resource {
	return &imageResource{}
}

type imageResource struct {
	resourceBase
}

type imageResourceAuthModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

type imageResourceModel struct {
	Auth          types.Object `tfsdk:"auth"`
	ContainerHost types.String `tfsdk:"container_host"`
	Id            types.String `tfsdk:"id"`
	Policy        types.String `tfsdk:"policy"`
	Preserve      types.Bool   `tfsdk:"preserve"`
	PullNumber    types.Int32  `tfsdk:"pull_number"`
	Reference     types.String `tfsdk:"reference"`
}

func (r *imageResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image"
}

func (r *imageResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Podman Image resource",
		Attributes: map[string]schema.Attribute{
			"auth": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						Required:  true,
						Sensitive: true,
						WriteOnly: true,
					},
					"password": schema.StringAttribute{
						Required:  true,
						Sensitive: true,
						WriteOnly: true,
					},
				},
				MarkdownDescription: "Optional credentials to use when pulling from this image's repository",
				Optional:            true,
				WriteOnly:           true,
			},
			"container_host": schema.StringAttribute{
				MarkdownDescription: "URL of the container host where this resource resides",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Image ID calculated by the container runtime",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy": schema.StringAttribute{
				MarkdownDescription: `The circumstances under which this image should be pulled. Options are: "always" (default), "missing", "newer", "never".`,
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("always", "missing", "never", "newer"),
				},
			},
			"preserve": schema.BoolAttribute{
				MarkdownDescription: "When true, do not delete the underlying image on the Podman host after this resource is deleted by Terraform. The storage taken up by this image will need to be released manually using e.g. the Podman CLI at a later time, but this behavior may be useful when you are doing some local testing and don't want to run into Docker Hub's stringent free tier rate limits.\n\n" +
					"  Note that this behavior also applies when this resource is deleted and recreated due to a new image tag being pulled. The image corresponding to the old tag will be preserved in this situation.",
				Optional: true,
			},
			"pull_number": schema.Int32Attribute{
				MarkdownDescription: "Increment this number to force an immediate pull of this container, provided that this image resource's `policy` allows it.\n\n " +
					"  This feature might be useful for pulling a rapidly-changing `latest` tag corresponding to a CI build artifact, but in general it is recommended that you reference images by stable tag and/or digest instead if possible.",
				Optional: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
			},
			"reference": schema.StringAttribute{
				MarkdownDescription: "Reference of the image to pull. Stable image tags (which the image publisher promises not to change) should be used here, and the use of image digests is strongly recommended when pulling from registries that you do not control, in order to protect against supply chain attacks.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Required: true,
			},
		},
	}
}

func (r *imageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data imageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	c, err := r.ps.getClient(ctx, data.ContainerHost.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Connection Error", err.Error())

		return
	}

	in := api.ImagePullQuery{
		Policy:    data.Policy.ValueString(),
		Reference: data.Reference.ValueString(),
	}

	var auth *api.RegistryAuth
	var authAttr types.Object

	// Write-only attributes are only available in the Config, not the Plan.
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("auth"), &authAttr)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !authAttr.IsNull() {
		var authData imageResourceAuthModel
		resp.Diagnostics.Append(authAttr.As(ctx, &authData, basetypes.ObjectAsOptions{})...)

		if resp.Diagnostics.HasError() {
			return
		}

		auth = &api.RegistryAuth{
			Username: authData.Username.ValueString(),
			Password: authData.Password.ValueString(),
		}
	}

	ch, err := c.ImagePull(ctx, in, auth)

	if err != nil {
		resp.Diagnostics.AddError("Request failed", err.Error())

		return
	}

	var id string
	var ok bool

	for event := range ch {
		err, match := event.(error)

		if match {
			resp.Diagnostics.AddError("Communication error", err.Error())

			return
		}

		errEvent, match := event.(api.ImagePullErrorEvent)

		if match {
			resp.Diagnostics.AddError("Server returned an error", errEvent.Error)

			return
		}

		msgEvent, match := event.(api.ImagePullStreamEvent)

		if match {
			tflog.Trace(ctx, "Progress message", map[string]any{"msg": msgEvent.Stream})
		}

		idEvent, match := event.(api.ImagePullImagesEvent)

		if match {
			tflog.Trace(ctx, "Image pulled", map[string]any{"id": idEvent.Id})
			id = idEvent.Id
			ok = true
		}
	}

	if !ok {
		resp.Diagnostics.AddError("Protocol error", "No image ID was received from the container host")

		return
	}

	data.Id = types.StringValue(id)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *imageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data imageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Preserve.ValueBool() {
		tflog.Trace(
			ctx,
			"Preserving underlying image",
			map[string]any{"id": data.Id.ValueString()})

		return
	}

	c, err := r.ps.getClient(ctx, data.ContainerHost.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Connection error", err.Error())

		return
	}

	err = c.ImageDelete(ctx, data.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Delete failed", err.Error())

		return
	}
}

func (r *imageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data imageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	c, err := r.ps.getClient(ctx, data.ContainerHost.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Connection Error", err.Error())

		return
	}

	_, err = c.ImageInspect(ctx, data.Id.ValueString())

	if err != nil {
		status, ok := err.(client.StatusCodeError)

		if ok && status.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Error inspecting image", err.Error())
		}

		return
	}
}

func (r *imageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var newData imageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &newData)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newData)...)
}
