package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (co *containerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var oldData containerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &oldData)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var newData containerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &newData)...)

	if resp.Diagnostics.HasError() {
		return
	}

	c, err := co.ps.getClient(ctx, oldData.ContainerHost.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Connection Error", err.Error())

		return
	}

	id := oldData.Id.ValueString()
	oldName := oldData.Name.ValueString()
	newName := newData.Name.ValueString()

	if oldName != newName {
		err = c.ContainerRename(ctx, id, newName)

		if err != nil {
			resp.Diagnostics.AddError("Error renaming container", err.Error())

			return
		}

		tflog.Trace(ctx, "Renamed container", map[string]any{
			"id":  id,
			"old": oldName,
			"new": newName,
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newData)...)
}
