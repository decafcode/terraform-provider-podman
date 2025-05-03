package provider

import (
	"archive/tar"
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

	// See docs for diffUpload(), this is a bit fiddly.

	var uploads []*containerResourceUploadModel
	resp.Diagnostics.Append(diffUploads(ctx, oldData.Uploads, newData.Uploads, &uploads)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if len(uploads) > 0 {
		err := c.ContainerArchive(ctx, id, func(arc *tar.Writer) error {
			for i := range uploads {
				if uploads[i] != nil {
					var content types.String
					d := req.Config.GetAttribute(
						ctx,
						path.Root("uploads").AtListIndex(i).AtName("content"),
						&content)

					if d.HasError() {
						resp.Diagnostics.Append(d...)

						return fmt.Errorf("error extracting upload content from Terraform config")
					}

					err := writeUpload(ctx, arc, uploads[i], content.ValueString())

					if err != nil {
						return err
					}
				}
			}

			return nil
		})

		if err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("uploads"), "File upload failed", err.Error())

			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newData)...)
}

// A slightly fiddly diffing algorithm for the contents of the `uploads` list attribute. The `out`
// parameter returns a slice of pointers to model objects, and this array has the same size as
// the `after` List. Elements in this array are non-nil if the non-write-only attributes of the
// element have changed, otherwise the entry is nil.
//
// In the case where there are no changes whatsoever an empty slice is returned instead.
//
// The calling function will then check to see if the resulting diff is nonempty, and if it is
// nonempty then the calling function will begin streaming a tar to the Podman host. Each non-nil
// entry in the array will be uploaded to the Podman host. The array is returned in this sparse
// format because the indexes of the non-nil entries in this array need to correspond to the
// indexes of the original List in the attribute, and this in turn is important because we need to
// extract the write-only `content` field from the Config and not the Plan:
//
//   - `State` contains the previous non-write-only attributes.
//   - `Plan` contains the new values and automatic defaults, but no write-only attributes.
//   - `Config` contains write-only attributes, but no defaulted attributes.
//
// All three Lists need to be consulted for an update to the `uploads` attribute to be applied
// correctly.
func diffUploads(ctx context.Context, before types.List, after types.List, out *[]*containerResourceUploadModel) diag.Diagnostics {
	var result diag.Diagnostics

	beforeAry := make([]containerResourceUploadModel, 0)
	result.Append(before.ElementsAs(ctx, &beforeAry, false)...)

	if result.HasError() {
		return result
	}

	afterAry := make([]containerResourceUploadModel, 0)
	result.Append(after.ElementsAs(ctx, &afterAry, false)...)

	if result.HasError() {
		return result
	}

	idx := make(map[containerResourceUploadKey]bool, 0)

	for i := range beforeAry {
		idx[beforeAry[i].key()] = true
	}

	haveChanges := false
	delta := make([]*containerResourceUploadModel, 0)

	for i := range afterAry {
		item := afterAry[i]

		if !idx[item.key()] {
			haveChanges = true
			delta = append(delta, &item)
		} else {
			delta = append(delta, nil)
		}
	}

	if haveChanges {
		*out = delta
	} else {
		*out = nil
	}

	return result
}
