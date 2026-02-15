package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (*containerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Podman Container resource",
		Attributes: map[string]schema.Attribute{
			"command": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Override the default command specified by this container's image.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"container_host": schema.StringAttribute{
				MarkdownDescription: "URL of the container host where this resource resides",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"entrypoint": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Override the container entry point supplied by the image.",
				Optional:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"env": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Environment variables to set in this container. This is in addition to any environment variables specified by the image.",
				Optional:            true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Container ID assigned by Podman",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"image": schema.StringAttribute{
				MarkdownDescription: "ID of the image to launch inside this container. Pass the `id` attribute of a `podman_image` resource here.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Required: true,
			},
			"labels": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Labels to attach to this container in the Podman and Docker API.",
				Optional:            true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"mounts": schema.ListNestedAttribute{
				MarkdownDescription: "A list of host filesystem locations or block devices to mount into the container's mount namespace.\n\n" +
					"  The default type is a bind mount (i.e. make a host directory appear inside the container), but other possibilities also exist depending on the value of the `type` attribute.\n\n" +
					"  See [Podman docs](https://docs.podman.io/en/v5.5.2/markdown/podman-create.1.html#mount-type-type-type-specific-option) for more details, but be sure to also consult the note about the `options` attribute below.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"options": schema.ListAttribute{
							ElementType: types.StringType,
							MarkdownDescription: "Mount options as described in the link to Podman's docs above. Note that the documentation for this specific attribute is [here](https://docs.podman.io/en/v5.5.2/markdown/podman-create.1.html#volume-v-source-volume-host-dir-container-dir-options) and not in the section describing mounts. This seems to be a quirk of the Podman API.\n\n" +
								"  Note: If you are specifying a bind mount (the default mount type) and the host machine has SELinux enabled (which is usually the case, since Podman is typically used from Red Hat based distributions) then you will want to specify `[\"Z\"]` here, otherwise the processes running in the container will be denied access to this mount.",
							Optional: true,
						},
						"source": schema.StringAttribute{
							MarkdownDescription: "Source path on the host filesystem. Must be the empty string for `tmpfs` mounts.",
							Required:            true,
						},
						"target": schema.StringAttribute{
							MarkdownDescription: "Destination path inside the container.",
							Required:            true,
						},
						"type": schema.StringAttribute{
							Computed:            true,
							Default:             stringdefault.StaticString("bind"),
							MarkdownDescription: "What kind of mount to create. This affects the meaning of of the other attributes in this object. If this is not specified then the default is `bind`, which is probably what you want in most cases.",
							Optional:            true,
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplaceIfConfigured(),
				},
				Optional: true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name to assign to this container. Other containers on the same Podman network as this container will be able to discover this container's private IP address by looking up its name using DNS. Do note, however, that these DNS lookups do not work on Podman's default network (see description of `networks` below).\n\n" +
					"  If you do not specify a name here then a random name will be assigned by Podman. Assigning an explicit name is strongly recommended.",
				Optional: true,
			},
			"networks": schema.ListNestedAttribute{
				MarkdownDescription: "A list of Podman networks that this container should join. If this list is omitted or empty then the container will be added to the default Podman network.\n\n" +
					"  Podman networks allow participating containers to resolve the private IP addresses of other containers on the same network by looking up the names of those containers using DNS: this lookup is performed on just the bare name of the target container without any further qualifying domains.\n\n" +
					"  For compatibility with Docker this DNS lookup functionality is _not_ provided on Podman's default network.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "ID of the Podman network to join.",
							Required:            true,
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplaceIfConfigured(),
				},
				Optional: true,
			},
			"port_mappings": schema.ListNestedAttribute{
				MarkdownDescription: "List of ports to expose on the host's external network interfaces.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"container_port": schema.Int32Attribute{
							MarkdownDescription: "Container port to expose.",
							Required:            true,
							Validators: []validator.Int32{
								int32validator.Between(0, 65535),
							},
						},
						"host_ip": schema.StringAttribute{
							MarkdownDescription: "Host IPv4 or IPv6 address to bind to. Binds to all IPv4 addresses by default.",
							CustomType:          iptypes.IPAddressType{},
							Optional:            true,
						},
						"host_port": schema.Int32Attribute{
							MarkdownDescription: "Host port to bind to.",
							Required:            true,
							Validators: []validator.Int32{
								int32validator.Between(0, 65535),
							},
						},
						"protocols": schema.ListAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "IP protocols to forward. Must be some combination of `tcp`, `udp`, and `sctp`. Defaults to `[\"tcp\"]`.",
							Optional:            true,
							Validators: []validator.List{
								listvalidator.ValueStringsAre(
									stringvalidator.OneOf("tcp", "udp", "sctp"),
								),
							},
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplaceIfConfigured(),
				},
				Optional: true,
			},
			"restart_policy": schema.StringAttribute{
				MarkdownDescription: "The circumstances under which this container should be automatically restarted by Podman.\n\n" +
					"  If you want your containers to automatically start at boot time then enable the `podman-restart.service` systemd unit on the host. The podman CLI command executed by this unit will automatically start all containers that have a `restart_policy` of `always`.\n\n" +
					"  Valid values are:\n\n" +
					"  - `no` (default)\n" +
					"  - `always`\n" +
					"  - `on-failure`\n" +
					"  - `unless-stopped`\n",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("no", "always", "on-failure", "unless-stopped"),
				},
			},
			"secrets": schema.ListNestedAttribute{
				MarkdownDescription: "A list of Podman secrets to mount into the container's filesystem. See `uploads` below for an alternative mechanism that accomplishes a similar goal.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"gid": schema.Int32Attribute{
							Computed:            true,
							Default:             int32default.StaticInt32(0),
							MarkdownDescription: "Numerical group ID that owns the secret file. Defaults to 0 (root). Group names can not be specified here due to Podman API limitations.",
							Optional:            true,
							Validators: []validator.Int32{
								int32validator.Between(0, 65535),
							},
						},
						"mode": schema.Int32Attribute{
							Computed:            true,
							Default:             int32default.StaticInt32(0400),
							MarkdownDescription: "The numerical file mode to set on the file that holds the secret. HCL does not have any syntax for octal literals, so you may want to use Terraform's `parseint` function here (e.g. `parseint(\"400\", 8)`). The default value is octal `0400` (i.e. 256 in decimal).",
							Optional:            true,
						},
						"path": schema.StringAttribute{
							MarkdownDescription: "Path inside the container to mount the secret. Docker uses the convention `/run/secrets/<secret_name>`, but Podman allows arbitrary paths to be specified.",
							Required:            true,
						},
						"secret": schema.StringAttribute{
							MarkdownDescription: "The name or ID of the Podman secret to mount.",
							Required:            true,
						},
						"uid": schema.Int32Attribute{
							Computed:            true,
							Default:             int32default.StaticInt32(0),
							MarkdownDescription: "Numerical user ID that owns the secret file. Defaults to 0 (root). User names can not be specified here due to Podman API limitations.",
							Optional:            true,
							Validators: []validator.Int32{
								int32validator.Between(0, 65535),
							},
						},
					},
				},
				Optional: true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"secret_env": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "A string-to-string map of Podman secrets to supply to the container as environment variables. The keys are environment variable names, and the values are names or IDs of Podman secrets.",
				Optional:            true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"selinux_options": schema.ListAttribute{
				ElementType: types.StringType,
				MarkdownDescription: "Specify SELinux labelling options for this container.\n\n" +
					"  Mostly intended for advanced use cases. Semi-documented in an old version of Podman's documentation under the [--security-opt](https://docs.podman.io/en/v4.6.1/markdown/options/security-opt.html) command line argument. The options starting with `label=` are valid values for this array, but note that you will need to specify these options without the `label=` prefixes.\n\n" +
					"  The most commonly used value for this option is `[\"disable\"]`, which disables SELinux labelling. This lowers the security of the container, but it can be useful if you need to give the container access to Podman's API socket, since the standard SELinux policy will not let you do this by default even if you use the `Z` mount option when mounting the socket.",
				Optional: true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"start_immediately": schema.BoolAttribute{
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether to immediately start this container after it has been created and the `uploads` attribute has been processed. Default is `true`.",
				Optional:            true,
			},
			"uploads": schema.ListNestedAttribute{
				MarkdownDescription: "A list of files to upload to this container. Files uploaded during container creation will be uploaded before the container is started, if applicable. Changes to this attribute will result in the changed files being re-uploaded to the existing container.\n\n" +
					"  File content is not stored as part of Terraform state, so this mechanism can be used to supply secret data to the container such as private keys. However, it should only be used to upload small files, like secrets or configuration.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"base64": schema.BoolAttribute{
							MarkdownDescription: "Set this flag to true to interpret the `contents` attribute as base64-encoded binary data. Otherwise the `contents` attribute will be stored into the target file as UTF-8 text.",
							Optional:            true,
						},
						"content": schema.StringAttribute{
							MarkdownDescription: "The text or binary content to upload to the destination path (see `base64` attribute). This is a write-only attribute, so if you change this attribute then you will need to change the `contents_hash` attribute as well in order to trigger a re-upload.\n\n" +
								"  Use Terraform's `file(...)` function to set this attribute from a local file. If you have set the `base64` attribute to `true` then you will need to use Terraform's `filebase64(...)` function instead.",
							Required:  true,
							WriteOnly: true,
						},
						"content_hash": schema.StringAttribute{
							MarkdownDescription: "A hash of the value of `contents`. The actual hash algorithm does not matter as long as the hash changes every time the contents change. Changing the value of this attribute will trigger a re-upload of this file.",
							Required:            true,
						},
						"gid": schema.Int32Attribute{
							Computed:            true,
							Default:             int32default.StaticInt32(0),
							MarkdownDescription: "The numerical group ID that will own the file that will be created. Defaults to 0 (root). Group names can not be specified here due to Podman API limitations.",
							Optional:            true,
						},
						"mode": schema.Int32Attribute{
							Computed:            true,
							Default:             int32default.StaticInt32(0644),
							MarkdownDescription: "The numerical file mode to set on the file that will be created. HCL does not have any syntax for octal literals, so you may want to use Terraform's `parseint` function here (e.g. `parseint(\"644\", 8)`). The default value is octal `0644` (i.e. 420 in decimal).",
							Optional:            true,
						},
						"path": schema.StringAttribute{
							MarkdownDescription: "The absolute path of the file to create inside the container.",
							Required:            true,
						},
						"uid": schema.Int32Attribute{
							Computed:            true,
							Default:             int32default.StaticInt32(0),
							MarkdownDescription: "The numerical user ID that will own the file that will be created. Defaults to 0 (root). User names can not be specified here due to Podman API limitations.",
							Optional:            true,
						},
					},
				},
				Optional: true,
			},
			"user": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"group": schema.StringAttribute{
						MarkdownDescription: "Group ID",
						Optional:            true,
					},
					"user": schema.StringAttribute{
						MarkdownDescription: "User ID",
						Required:            true,
					},
				},
				MarkdownDescription: "The security principal that will be used to launch this container, consisting of a user ID and an optional group ID. Each ID can be specified as either an integer or a name. If a name is used then its corresponding numeric ID will be looked up in the `/etc/passwd` or `/etc/group` file inside the container image as appropriate.\n\n" +
					"  If this attribute is not specified then the default UID and GID from the image will be used.",
				Optional: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
			},
			"user_namespace": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"mode": schema.StringAttribute{
						Required: true,
					},
					"options": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
				},
				Optional: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}
