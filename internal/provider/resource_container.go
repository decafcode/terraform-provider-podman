package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func newContainerResource() resource.Resource {
	return &containerResource{}
}

type containerResource struct {
	resourceBase
}

type containerResourceMountModel struct {
	Target  types.String `tfsdk:"target"`
	Options types.List   `tfsdk:"options"`
	Source  types.String `tfsdk:"source"`
	Type    types.String `tfsdk:"type"`
}

type containerResourceNamespaceModel struct {
	Mode    types.String `tfsdk:"mode"`
	Options types.List   `tfsdk:"options"`
}

type containerResourceNetworkModel struct {
	Id types.String `tfsdk:"id"`
}

type containerResourcePortMappingModel struct {
	ContainerPort types.Int32         `tfsdk:"container_port"`
	HostIP        iptypes.IPv4Address `tfsdk:"host_ip"`
	HostPort      types.Int32         `tfsdk:"host_port"`
	Protocols     types.List          `tfsdk:"protocols"`
}

type containerResourceSecretModel struct {
	Gid    types.Int32  `tfsdk:"gid"`
	Mode   types.Int32  `tfsdk:"mode"`
	Path   types.String `tfsdk:"path"`
	Secret types.String `tfsdk:"secret"`
	Uid    types.Int32  `tfsdk:"uid"`
}

type containerResourceUserModel struct {
	Group types.String `tfsdk:"group"`
	User  types.String `tfsdk:"user"`
}

type containerResourceUploadModel struct {
	Base64      types.Bool   `tfsdk:"base64"`
	Content     types.String `tfsdk:"content"`
	ContentHash types.String `tfsdk:"content_hash"`
	Gid         types.Int32  `tfsdk:"gid"`
	Mode        types.Int32  `tfsdk:"mode"`
	Path        types.String `tfsdk:"path"`
	Uid         types.Int32  `tfsdk:"uid"`
}

type containerResourceModel struct {
	Command          types.List   `tfsdk:"command"`
	ContainerHost    types.String `tfsdk:"container_host"`
	Entrypoint       types.List   `tfsdk:"entrypoint"`
	Env              types.Map    `tfsdk:"env"`
	Id               types.String `tfsdk:"id"`
	Image            types.String `tfsdk:"image"`
	Labels           types.Map    `tfsdk:"labels"`
	Mounts           types.List   `tfsdk:"mounts"`
	Name             types.String `tfsdk:"name"`
	Networks         types.List   `tfsdk:"networks"`
	PortMappings     types.List   `tfsdk:"port_mappings"`
	RestartPolicy    types.String `tfsdk:"restart_policy"`
	Secrets          types.List   `tfsdk:"secrets"`
	SecretEnv        types.Map    `tfsdk:"secret_env"`
	SelinuxOptions   types.List   `tfsdk:"selinux_options"`
	StartImmediately types.Bool   `tfsdk:"start_immediately"`
	Uploads          types.List   `tfsdk:"uploads"`
	User             types.Object `tfsdk:"user"`
	UserNamespace    types.Object `tfsdk:"user_namespace"`
}

type containerResourceUploadKey struct {
	contentHash string
	gid         int32
	mode        int32
	path        string
	uid         int32
}

func (*containerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container"
}

func (r *containerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importState(ctx, req, resp)
}

func (u *containerResourceUploadModel) key() containerResourceUploadKey {
	return containerResourceUploadKey{
		contentHash: u.ContentHash.ValueString(),
		gid:         u.Gid.ValueInt32(),
		mode:        u.Mode.ValueInt32(),
		path:        u.Path.ValueString(),
		uid:         u.Uid.ValueInt32(),
	}
}
