package provider

import (
	"context"

	"github.com/decafcode/terraform-provider-podman/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (co *containerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data containerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.Id.ValueString()
	c, err := co.ps.getClient(ctx, data.ContainerHost.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Connection error", err.Error())

		return
	}

	json, err := c.ContainerInspect(ctx, id)

	if err != nil {
		status, ok := err.(client.StatusCodeError)

		if ok && status.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Error inspecting container", err.Error())
		}

		return
	}

	data.Name = types.StringValue(json.Name)
}
