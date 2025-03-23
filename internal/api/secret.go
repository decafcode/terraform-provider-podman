package api

type SecretCreateJson struct {
	Id string
}

type SecretInspectSpecJson struct {
	Name string
}

type SecretInspectJson struct {
	Id         string
	SecretData string
	Spec       SecretInspectSpecJson
}
