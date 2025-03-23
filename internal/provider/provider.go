package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type podmanProvider struct {
	version string
	env     PodmanProviderEnv
}

type podmanProviderModel struct {
	ContainerHost types.String `tfsdk:"container_host"`
}

func New(version string, env *PodmanProviderEnv) func() provider.Provider {
	return func() provider.Provider {
		return &podmanProvider{
			version: version,
			env:     *env,
		}
	}
}

func (p *podmanProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data podmanProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	state, err := newProviderState(&p.env)

	if err != nil {
		resp.Diagnostics.AddError("Error initializing provider", err.Error())
	}

	if p.env.ContainerHost != "" &&
		!data.ContainerHost.IsNull() &&
		data.ContainerHost.ValueString() != p.env.ContainerHost {

		resp.Diagnostics.AddWarning("Configuration conflict",
			"Environment variable CONTAINER_HOST is set, provider attribute container_host "+
				"is also set, and their values differ. The provider attribute will take "+
				"precedence over the environment variable.")
	}

	state.DefaultHost = p.env.ContainerHost

	if !data.ContainerHost.IsNull() {
		state.DefaultHost = data.ContainerHost.ValueString()
	}

	resp.ResourceData = state
}

func (p *podmanProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *podmanProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "podman"
	resp.Version = p.version
}

func (p *podmanProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newNetworkResource,
		newSecretResource,
	}
}

func (p *podmanProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"container_host": schema.StringAttribute{
				MarkdownDescription: "Default container host URL. Must be specified if resources do not specify a container_host attribute.",
				Optional:            true,
			},
		},
	}
}
