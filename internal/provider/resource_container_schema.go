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
				ElementType: types.StringType,
				Optional:    true,
				MarkdownDescription: "Override the default command specified by this " +
					"container's image.",
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
			"env": schema.MapAttribute{
				ElementType: types.StringType,
				MarkdownDescription: "Environment variables to set in this container. This is " +
					"in addition to any environment variables specified by the image.",
				Optional: true,
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
				MarkdownDescription: "ID of the image to launch inside this container. Pass " +
					"the `id` attribute of a `podman_image` resource here.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Required: true,
			},
			"labels": schema.MapAttribute{
				ElementType: types.StringType,
				MarkdownDescription: "Labels to attach to this container in the Podman and " +
					"Docker API.",
				Optional: true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"mounts": schema.ListNestedAttribute{
				MarkdownDescription: "A list of host filesystem locations or block devices to " +
					"mount into the container's mount namespace.\n\n" +
					"\tThe default type is a bind mount (i.e. make a host directory appear " +
					"inside the container), but other possibilities also exist depending on " +
					"the value of the `type` attribute.\n\n" +
					"\tSee [Podman docs](https://docs.podman.io/en/latest/markdown/podman-create.1.html#mount-type-type-type-specific-option) " +
					"for more details, but see the note about the `options` attribute below.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"options": schema.ListAttribute{
							Description: "Mount options as described in the link to Podman's " +
								"docs above. Note that the documentation for this specific " +
								"attribute is [here](https://docs.podman.io/en/latest/markdown/podman-create.1.html#volume-v-source-volume-host-dir-container-dir-options) " +
								"and not in the section describing mounts. This seems to be a " +
								"quirk of the Podman API.\n\n" +
								"\tNote: If you are specifying a bind mount (the default mount " +
								"type) and the host machine has SELinux enabled (which is " +
								"usually the case, since Podman is typically used from Red " +
								"Hat based distributions) then you will want to specify " +
								"`[\"Z\"]` here, otherwise the processes running in the " +
								"container will be denied access to this mount.",
							ElementType: types.StringType,
							Optional:    true,
						},
						"source": schema.StringAttribute{
							MarkdownDescription: "Source path on the host filesystem. Must be " +
								"the empty string for `tmpfs` mounts.",
							Required: true,
						},
						"target": schema.StringAttribute{
							MarkdownDescription: "Destination path inside the container.",
							Required:            true,
						},
						"type": schema.StringAttribute{
							Computed: true,
							Default:  stringdefault.StaticString("bind"),
							MarkdownDescription: "What kind of mount to create. This affects " +
								"the meaning of of the other attributes in this object. If this " +
								"is not specified then the default is `bind`, which is probably " +
								"what you want in most cases.",
							Optional: true,
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplaceIfConfigured(),
				},
				Optional: true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name to assign to this container. Other containers on " +
					"the same Podman network as this container will be able to discover this " +
					"container's private IP address by looking up its name using DNS. Do note, " +
					"however, that these DNS lookups do not work on Podman's default " +
					"network (see description of `networks` below).\n\n" +
					"\tIf you do not specify a name here then a random name will be assigned " +
					"by Podman. Assigning an explicit name is recommended.",
				Optional: true,
			},
			"networks": schema.ListNestedAttribute{
				MarkdownDescription: "A list of Podman networks that this container should " +
					"join. If this list is omitted or empty then the container will be added " +
					"to the default Podman network.\n\n" +
					"\tPodman networks allow participating containers to resolve the private IP" +
					"addresses of other containers on the same network by looking up the names " +
					"of those containers using DNS: this lookup is performed on just the bare " +
					"name of the target container without any further qualifying domains.\n\n" +
					"\tFor compatibility with Docker this DNS lookup functionality is _not_ " +
					"provided on Podman's default network.",
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
				MarkdownDescription: "List of ports to expose on the host's external network " +
					"interfaces.",
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
							MarkdownDescription: "Host IPv4 address to bind to. Binds to all " +
								"interfaces if omitted.",
							CustomType: iptypes.IPv4AddressType{},
							Optional:   true,
						},
						"host_port": schema.Int32Attribute{
							MarkdownDescription: "Host port to bind to.",
							Required:            true,
							Validators: []validator.Int32{
								int32validator.Between(0, 65535),
							},
						},
						"protocols": schema.ListAttribute{
							ElementType: types.StringType,
							MarkdownDescription: "IP protocols to forward. Must be some " +
								"combination of `tcp`, `udp`, and `sctp`. Defaults to " +
								"`[\"tcp\"]`.",
							Optional: true,
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
				MarkdownDescription: "The circumstances under which this container should be " +
					"automatically restarted by Podman.\n\n" +
					"\tIf you want your containers to automatically start at boot time then " +
					"enable the `podman-restart.service` systemd unit on the host. The podman " +
					"CLI command that this unit executes by default will automatically start " +
					"all containers that have a `restart_policy` of `always`.\n\n" +
					"\tValid values are:\n\n" +
					"\t- `no` (default)\n" +
					"\t- `always`\n" +
					"\t- `on-failure`\n" +
					"\t- `unless-stopped`\n",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf("no", "always", "on-failure", "unless-stopped"),
				},
			},
			"secrets": schema.ListNestedAttribute{
				MarkdownDescription: "A list of Podman secrets to mount into the container's " +
					"filesystem. See `uploads` below for an alternative mechanism that " +
					"accomplishes a similar goal.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"path": schema.StringAttribute{
							MarkdownDescription: "Path inside the container to mount the " +
								"secret. Docker uses the convention " +
								"`/run/secrets/<secret_name>`, but Podman allows arbitrary " +
								"paths to be specified.",
							Optional: true,
						},
						"secret": schema.StringAttribute{
							MarkdownDescription: "The name or ID of the Podman secret to mount.",
							Required:            true,
						},
					},
				},
				Optional: true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"secret_env": schema.MapAttribute{
				ElementType: types.StringType,
				MarkdownDescription: "A string-to-string map of Podman secrets to supply to " +
					"the container as environment variables. The keys are environment " +
					"variable names, and the values are names or IDs of Podman secrets.",
				Optional: true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"start_immediately": schema.BoolAttribute{
				Computed: true,
				Default:  booldefault.StaticBool(true),
				MarkdownDescription: "Whether to immediately start this container after it has " +
					"been created and the `uploads` attribute has been processed. Default is " +
					"`true`.",
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
				MarkdownDescription: "The security principal that will be used to launch this " +
					"container, consisting of a user ID and an optional group ID. Each ID can " +
					"be specified as either an integer or a name. If a name is used then its " +
					"corresponding numeric ID will be looked up in the `/etc/passwd` or " +
					"`/etc/group` file inside the container image as appropriate.\n\n" +
					"\tIf this attribute is not specified then the default UID and GID from the" +
					"image will be used.",
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
