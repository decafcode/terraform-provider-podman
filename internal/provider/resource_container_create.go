package provider

import (
	"archive/tar"
	"context"
	"fmt"
	"strings"

	"github.com/decafcode/terraform-provider-podman/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (co *containerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data containerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	in := api.ContainerCreateJson{
		Command:       make([]string, 0),
		Env:           make(map[string]string, 0),
		Image:         data.Image.ValueString(),
		Name:          data.Name.ValueString(),
		Networks:      make(map[string]api.ContainerCreateNetworkJson, 0),
		RestartPolicy: data.RestartPolicy.ValueString(),
		Secrets:       make([]api.ContainerCreateSecretJson, 0),
		SecretEnv:     make(map[string]string, 0),
		SelinuxOpts:   make([]string, 0),
		Netns: api.ContainerCreateNamespaceJson{
			// This is hardcoded for now
			NSMode: "bridge",
		},
	}

	resp.Diagnostics.Append(data.Command.ElementsAs(ctx, &in.Command, false)...)
	resp.Diagnostics.Append(data.Env.ElementsAs(ctx, &in.Env, false)...)
	resp.Diagnostics.Append(data.Labels.ElementsAs(ctx, &in.Labels, false)...)
	resp.Diagnostics.Append(writeMounts(ctx, &data.Mounts, &in.Mounts)...)
	resp.Diagnostics.Append(writeNetworks(ctx, &data.Networks, &in.Networks)...)
	resp.Diagnostics.Append(writePortMappings(ctx, &data.PortMappings, &in.PortMappings)...)
	resp.Diagnostics.Append(writeSecrets(ctx, &data.Secrets, &in.Secrets)...)
	resp.Diagnostics.Append(data.SecretEnv.ElementsAs(ctx, &in.SecretEnv, false)...)
	resp.Diagnostics.Append(data.SelinuxOptions.ElementsAs(ctx, &in.SelinuxOpts, false)...)
	resp.Diagnostics.Append(writeUser(ctx, &data.User, &in.User)...)
	resp.Diagnostics.Append(writeNamespace(ctx, &data.UserNamespace, &in.Userns)...)

	if resp.Diagnostics.HasError() {
		return
	}

	c, err := co.ps.getClient(ctx, data.ContainerHost.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Connection Error", err.Error())

		return
	}

	out, err := c.ContainerCreate(ctx, &in)

	if err != nil {
		resp.Diagnostics.AddError("Container create failed", err.Error())

		return
	}

	tflog.Trace(ctx, "Container created", map[string]any{"id": out.Id})

	for i := range out.Warnings {
		resp.Diagnostics.AddWarning("Container creation returned warning", out.Warnings[i])
	}

	data.Id = types.StringValue(out.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	uploads := make([]containerResourceUploadModel, 0)
	resp.Diagnostics.Append(data.Uploads.ElementsAs(ctx, &uploads, false)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if len(uploads) > 0 {
		err = c.ContainerArchive(ctx, out.Id, func(arc *tar.Writer) error {
			for i := range uploads {
				var content types.String
				d := req.Config.GetAttribute(
					ctx,
					path.Root("uploads").AtListIndex(i).AtName("content"),
					&content)

				if d.HasError() {
					resp.Diagnostics.Append(d...)

					return fmt.Errorf("error extracting upload content from Terraform config")
				}

				err := writeUpload(ctx, arc, &uploads[i], content.ValueString())

				if err != nil {
					return err
				}
			}

			return nil
		})

		if err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("uploads"), "File upload failed", err.Error())

			return
		}
	}

	if data.StartImmediately.ValueBool() {
		err = c.ContainerStart(ctx, out.Id)

		if err != nil {
			resp.Diagnostics.AddError("Container start failed", err.Error())
		}
	}
}

func writeMounts(ctx context.Context, in *types.List, out *[]api.ContainerCreateMountJson) diag.Diagnostics {
	var result diag.Diagnostics

	models := make([]containerResourceMountModel, 0)
	result.Append(in.ElementsAs(ctx, &models, false)...)

	if result.HasError() {
		return result
	}

	for i := range models {
		inItem := models[i]

		var options []string
		result.Append(inItem.Options.ElementsAs(ctx, &options, false)...)

		if result.HasError() {
			return result
		}

		*out = append(*out, api.ContainerCreateMountJson{
			Destination: inItem.Target.ValueString(),
			Options:     options,
			Source:      inItem.Source.ValueString(),
			Type:        inItem.Type.ValueString(),
		})
	}

	return result
}

func writeNamespace(ctx context.Context, in *types.Object, out *api.ContainerCreateNamespaceJson) diag.Diagnostics {
	var result diag.Diagnostics

	if in.IsNull() {
		return result
	}

	var model containerResourceNamespaceModel
	result.Append(in.As(ctx, &model, basetypes.ObjectAsOptions{})...)

	if result.HasError() {
		return result
	}

	out.NSMode = model.Mode.ValueString()

	var options []string
	result.Append(model.Options.ElementsAs(ctx, &options, false)...)

	if result.HasError() {
		return result
	}

	out.Value = strings.Join(options, ",")

	return result
}

func writeNetworks(ctx context.Context, in *types.List, out *map[string]api.ContainerCreateNetworkJson) diag.Diagnostics {
	var result diag.Diagnostics

	models := make([]containerResourceNetworkModel, 0)
	result.Append(in.ElementsAs(ctx, &models, false)...)

	if result.HasError() {
		return result
	}

	for i := range models {
		(*out)[models[i].Id.ValueString()] = api.ContainerCreateNetworkJson{}
	}

	return result
}

func writePortMappings(ctx context.Context, in *types.List, out *[]api.ContainerCreatePortMappingJson) diag.Diagnostics {
	var result diag.Diagnostics

	models := make([]containerResourcePortMappingModel, 0)
	result.Append(in.ElementsAs(ctx, &models, false)...)

	if result.HasError() {
		return result
	}

	for i := range models {
		model := models[i]

		protocolList := make([]string, 0)
		result.Append(model.Protocols.ElementsAs(ctx, &protocolList, false)...)

		if result.HasError() {
			return result
		}

		var hostIpStr string

		if !model.HostIP.IsNull() && !model.HostIP.IsUnknown() {
			hostIpStr = model.HostIP.ValueString()
		}

		*out = append(*out, api.ContainerCreatePortMappingJson{
			ContainerPort: uint16(model.ContainerPort.ValueInt32()),
			HostIP:        hostIpStr,
			HostPort:      uint16(model.HostPort.ValueInt32()),
			Protocol:      strings.Join(protocolList, ","),
		})
	}

	return result
}

func writeSecrets(ctx context.Context, in *types.List, out *[]api.ContainerCreateSecretJson) diag.Diagnostics {
	var result diag.Diagnostics

	models := make([]containerResourceSecretModel, 0)
	result.Append(in.ElementsAs(ctx, &models, false)...)

	if result.HasError() {
		return result
	}

	for i := range models {
		*out = append(*out, api.ContainerCreateSecretJson{
			GID:    uint32(models[i].Gid.ValueInt32()),
			Mode:   uint32(models[i].Mode.ValueInt32()),
			Source: models[i].Secret.ValueString(),
			Target: models[i].Path.ValueString(),
			UID:    uint32(models[i].Uid.ValueInt32()),
		})
	}

	return result
}

func writeUser(ctx context.Context, in *types.Object, out *string) diag.Diagnostics {
	var result diag.Diagnostics

	if in.IsNull() {
		return result
	}

	var model containerResourceUserModel
	result.Append(in.As(ctx, &model, basetypes.ObjectAsOptions{})...)

	if result.HasError() {
		return result
	}

	uid := model.User.ValueString()

	if model.Group.IsNull() {
		*out = uid
	} else {
		gid := model.Group.ValueString()
		*out = fmt.Sprintf("%s:%s", uid, gid)
	}

	return result
}
