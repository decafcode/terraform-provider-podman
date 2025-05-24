package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (co *containerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

	err = c.ContainerStop(ctx, id)

	if err != nil {
		resp.Diagnostics.AddError("Error stopping container", err.Error())

		return
	}

	err = c.ContainerDelete(ctx, id)

	if err != nil {
		resp.Diagnostics.AddError("Error deleting container", err.Error())

		return
	}
}
