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
	parts := strings.Split(req.ID, "@")
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[0])...)

	if len(parts) > 1 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("container_host"), parts[1])...)
	}
}
