package api

type ContainerInspectConfigJson struct {
	Cmd    []string
	Env    []string
	Labels map[string]string
	User   string
}

type ContainerInspectMountJson struct {
	Destination string
	Type        string
	Source      string
	Options     []string
}

type ContainerInspectNetworkJson struct {
	NetworkID string
}

type ContainerInspectNetworksJson struct {
	Networks map[string]ContainerInspectNetworkJson
}

type ContainerInspectRestartPolicyJson struct {
	Name string
}

type ContainerInspectHostConfigJson struct {
	RestartPolicy ContainerInspectRestartPolicyJson
}

type ContainerInspectJson struct {
	Config          ContainerInspectConfigJson
	HostConfig      ContainerInspectHostConfigJson
	Id              string
	Image           string
	Name            string
	Mounts          []ContainerInspectMountJson
	NetworkSettings ContainerInspectNetworksJson
}
