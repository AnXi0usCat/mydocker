package main

const (
	authUrl     = "https://auth.docker.io/token?service=registry.docker.io&scope=repository:library/%s:pull"
	manifestUrl = "https://registry.hub.docker.com/v2/library/%s/manifests/%s"
	layerUrl    = "https://registry.hub.docker.com/v2/library/%s/blobs/%s"
)

type DockerAuth struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expiresIn"`
	IssuedAt  string `json:"issuedAt"`
}

type DockerManifest struct {
	SchemaVersion int    `json:"schemaVersion"`
	MediaType     string `json:"mediaType"`
	Config        struct {
		MediaType string `json:"mediaType"`
		Digest    string `json:"digest"`
		Size      int    `json:"size"`
	} `json:"config"`
	Layers []struct {
		MediaType string `json:"mediaType"`
		Digest    string `json:"digest"`
		Size      int    `json:"size"`
	} `json:"layers"`
}
