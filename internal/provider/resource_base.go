package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

type resourceBase struct {
	ps *podmanProviderState
}

func (r *resourceBase) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	ps, ok := req.ProviderData.(*podmanProviderState)

	if !ok {
		resp.Diagnostics.AddError("Internal error", "Invalid provider state type")

		return
	}

	r.ps = ps
}

func importState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	pos := strings.Index(req.ID, ",")

	if pos != -1 {
		id := req.ID[:pos]
		container_host := req.ID[(pos + 1):]
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("container_host"), container_host)...)
	} else {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	}
}
