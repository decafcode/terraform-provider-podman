package api

type ImageJson struct {
	Id    string   `json:"id"`
	Names []string `json:"names"`
}

type ImagePullErrorEvent struct {
	Error string `json:"error"`
}

type ImagePullImagesEvent struct {
	Id     string   `json:"id"`
	Images []string `json:"images"`
}

type ImagePullQuery struct {
	Policy    string
	Reference string
}

type ImagePullStreamEvent struct {
	Stream string `json:"stream"`
}

type RegistryAuth struct {
	Email    string
	Password string
	Username string
}
