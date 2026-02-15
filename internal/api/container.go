package api

import "time"

type ContainerCreateDeviceJson struct {
	Path string `json:"path"`
}

type ContainerCreateHealthConfigJson struct {
	Interval      time.Duration `json:",omitempty"`
	Retries       int32         `json:",omitempty"`
	StartInterval time.Duration `json:",omitempty"`
	StartPeriod   time.Duration `json:",omitempty"`
	Test          []string      `json:",omitempty"`
	Timeout       time.Duration `json:",omitempty"`
}

type ContainerCreateMountJson struct {
	Destination string
	Options     []string
	Source      string
	Type        string
}

type ContainerCreateNamespaceJson struct {
	NSMode string `json:"nsmode"`
	Value  string `json:"value"`
}

type ContainerCreateNetworkJson struct {
	// Nothing yet
}

type ContainerCreatePortMappingJson struct {
	ContainerPort uint16 `json:"container_port"`
	HostIP        string `json:"host_ip"`
	HostPort      uint16 `json:"host_port"`
	Protocol      string `json:"protocol,omitempty"`
}

type ContainerCreateSecretJson struct {
	Source string
	Target string
	UID    uint32
	GID    uint32
	Mode   uint32
}

type ContainerCreateJson struct {
	Name          string                                `json:"name,omitempty"`
	Image         string                                `json:"image,omitempty"`
	Command       []string                              `json:"command,omitempty"`
	Devices       []ContainerCreateDeviceJson           `json:"devices,omitempty"`
	Env           map[string]string                     `json:"env,omitempty"`
	Entrypoint    []string                              `json:"entrypoint,omitempty"`
	HealthConfig  *ContainerCreateHealthConfigJson      `json:"healthconfig,omitempty"`
	Labels        map[string]string                     `json:"labels"`
	Mounts        []ContainerCreateMountJson            `json:"mounts,omitempty"`
	Netns         ContainerCreateNamespaceJson          `json:"netns"`
	Networks      map[string]ContainerCreateNetworkJson `json:"networks,omitempty"`
	PortMappings  []ContainerCreatePortMappingJson      `json:"portmappings,omitempty"`
	RestartPolicy string                                `json:"restart_policy"`
	SecretEnv     map[string]string                     `json:"secret_env,omitempty"`
	Secrets       []ContainerCreateSecretJson           `json:"secrets,omitempty"`
	SelinuxOpts   []string                              `json:"selinux_opts,omitempty"`
	User          string                                `json:"user"`
	Userns        ContainerCreateNamespaceJson          `json:"userns"`
}

type ContainerCreatedJson struct {
	Id       string   `json:"id"`
	Warnings []string `json:"warnings"`
}

type ContainerInspectJson struct {
	Name string
}
