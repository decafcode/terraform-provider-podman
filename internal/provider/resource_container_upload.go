package provider

import (
	"archive/tar"
	"context"
	"encoding/base64"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func writeUpload(ctx context.Context, arc *tar.Writer, model *containerResourceUploadModel, content string) error {
	var bytes []byte

	if model.Base64.ValueBool() {
		b, err := base64.StdEncoding.DecodeString(content)

		if err != nil {
			return err
		}

		bytes = b
	} else {
		bytes = []byte(content)
	}

	err := arc.WriteHeader(&tar.Header{
		Typeflag: tar.TypeReg,
		Gid:      int(model.Gid.ValueInt32()),
		Mode:     int64(model.Mode.ValueInt32()),
		Name:     model.Path.ValueString(),
		Size:     int64(len(bytes)),
		Uid:      int(model.Uid.ValueInt32()),
	})

	if err != nil {
		return err
	}

	written, err := arc.Write(bytes)

	if err != nil {
		return err
	}

	if written != len(bytes) {
		return fmt.Errorf("short write %d/%d", written, len(bytes))
	}

	tflog.Trace(ctx, "Uploading file", map[string]any{
		"content_hash": model.ContentHash.ValueString(),
		"size":         len(bytes),
	})

	return nil
}
