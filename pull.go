package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

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

func authenticate(image string) *DockerAuth {
	url := fmt.Sprintf(authUrl, image)
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Authenticated successfully")
	defer response.Body.Close()

	buf, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Read authentiction response payload")

	var auth DockerAuth
	json.Unmarshal(buf, &auth)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Unamrshalled json payload from the auth endpoint")

	return &auth
}

func getManifest(auth *DockerAuth, image, version string) *DockerManifest {
	url := fmt.Sprintf(manifestUrl, image, version)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Set("Authorization", "Bearer "+auth.Token)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Successfully returned response from manifest endpoint")
	defer resp.Body.Close()

	// error could be nil but the response still might have a non 200 status
	if resp.StatusCode != http.StatusOK {
		log.Fatal(fmt.Sprintf("Response from the manifest has a non 200 status code %d", resp.StatusCode))
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("successfully read the manifest payload")

	var manifest DockerManifest
	err = json.Unmarshal(buf, &manifest)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Unmarshalled manifest payload to a go type")
	log.Println(manifest)

	return &manifest
}

func downloadLayer(auth *DockerAuth, url, outfile string) {
	output, err := os.Create(outfile)
	if err != nil {
		log.Fatal(err)
	}
	defer output.Close()

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Set("Authorization", "Bearer "+auth.Token)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Successful response from the image endpoint")
	defer resp.Body.Close()

	// error could be nil but the response still might have a non 200 status
	if resp.StatusCode != http.StatusOK {
		log.Fatal(fmt.Sprintf("Response from dockerhub has a non 200 status code %d", resp.StatusCode))
	}

	_, err = io.Copy(output, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Downlaoded layer")
}

func extract(filename, dest string) {
	cmd := exec.Command("tar", "-xzf", filename, "-C", dest)
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	// delete archive
	err = os.Remove(filename)
	if err != nil {
		log.Fatal(err)
	}
}

func download(image, dest string) {
	name, version, ok := strings.Cut(image, ":")
	if !ok {
		name = image
		version = "latest"
	}
	log.Println(fmt.Sprintf("Resolving: image %s, version %s", name, version))

	auth := authenticate(name)
	manifest := getManifest(auth, image, version)
	log.Println("manifest download complete")

	for i, layer := range manifest.Layers {
		log.Println(layer.Digest)
		url := fmt.Sprintf(layerUrl, name, layer.Digest)
		outfile := filepath.Join(dest, fmt.Sprintf("layer-%d.tar", i))
		downloadLayer(auth, url, outfile)
		extract(outfile, dest)
	}
}

func delete(dest string) {
	err := os.RemoveAll(dest + "/")
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to remove working directory %s", err))
	}
}
