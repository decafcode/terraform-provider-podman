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
	ContainerHost     types.String `tfsdk:"container_host"`
	HostKeyAlgorithms types.List   `tfsdk:"host_key_algorithms"`
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

	if !data.HostKeyAlgorithms.IsNull() {
		resp.Diagnostics.Append(
			data.HostKeyAlgorithms.ElementsAs(ctx, &state.HostKeyAlgorithms, false)...,
		)
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
		newContainerResource,
		newImageResource,
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
			"host_key_algorithms": schema.ListAttribute{
				ElementType: types.StringType,
				MarkdownDescription: "An ordered list of public key type names (of the kind found in the second field of an entry in your `~/.ssh/authorized_keys` file) to request from remote SSH servers. If this is not specified then the default sequence of algorithms built into the Go `crypto/ssh` library will be used.\n\n" +
					"  The first key type that the server supports will be used for SSH host key checks and any other host key types will be ignored. This is less secure than OpenSSH, which checks all of a remote host's known keys, but this deficiency is due to what appears to be a limitation in the API of Go's `crypto/ssh` module.\n\n" +
					"  The `crypto/ssh` module seems to start negotiations by requesting host keys based on NIST elliptic curves by default, so you might want to specify `[\"ssh-ed25519\"]` here to force the use of the less-dubious Ed25519 algorithm instead.",
				Optional: true,
			},
		},
	}
}
