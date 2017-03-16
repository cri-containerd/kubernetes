/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package containerdshim

import (
	"fmt"
	"os/exec"
	"strings"

	runtimeapi "k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/runtime"
)

// containerd doesn't have metadata store now, so save the metadata ourselves.
// imageStore is a map from image digest to image metadata.
// No need to store detailed image information layers now, because for the POC
// we'll re-fetch it when creating the rootfs.
var imageStore map[string]*runtimeapi.Image = map[string]*runtimeapi.Image{}

// P0
// Ignore the filter for now.
func (cs *containerdService) ListImages(filter *runtimeapi.ImageFilter) ([]*runtimeapi.Image, error) {
	var images []*runtimeapi.Image
	for _, image := range imageStore {
		images = append(images, image)
	}
	return images, nil
}

// P0
// The image here could be either digest or reference in current implementation.
func (cs *containerdService) ImageStatus(image *runtimeapi.ImageSpec) (*runtimeapi.Image, error) {
	// Try digest first.
	if img, ok := imageStore[image.Image]; ok {
		return img, nil
	}
	// Try image reference.
	for _, img := range imageStore {
		for _, t := range img.RepoTags {
			if image.Image == t {
				return img, nil
			}
		}
	}
	return nil, nil
}

// P0
// For the POC code, docker image must be `docker.io/library/image:tag` or `docker.io/library/image`.
func (cs *containerdService) PullImage(image *runtimeapi.ImageSpec, auth *runtimeapi.AuthConfig) (string, error) {
	repo, tag := repoAndTag(image.Image)
	digest, err := imageDigest(repo, tag)
	if err != nil {
		return "", fmt.Errorf("failed to get image digest %q: %v", image.Image, err)
	}
	if err := pullImage(repo, tag); err != nil {
		return "", fmt.Errorf("failed to pull image %q: %v", image.Image, err)
	}
	if _, ok := imageStore[digest]; !ok {
		imageStore[digest] = &runtimeapi.Image{
			Id:          digest,
			RepoDigests: []string{digest},
			// Use fake image size, because we don't care about it in the POC.
			Size_: 1024,
		}
	}
	img := imageStore[digest]
	// Add new image tag
	for _, t := range img.RepoTags {
		if image.Image == t {
			return digest, nil
		}
	}
	img.RepoTags = append(img.RepoTags, image.Image)
	// Return the image digest
	return digest, nil
}

// P1
func (cs *containerdService) RemoveImage(image *runtimeapi.ImageSpec) error {
	// Only remove image from the internal metadata for now. Note that the image
	// must be digest here in current implementation.
	delete(imageStore, image.Image)
	return nil
}

const mediaType = "mediatype:application/vnd.docker.distribution.manifest.v2+json"

func repoAndTag(image string) (string, string) {
	repoAndTag := strings.Split(image, ":")
	if len(repoAndTag) < 2 {
		return image, "latest"
	}
	return repoAndTag[0], repoAndTag[1]
}

func imageDigest(repo, tag string) (string, error) {
	output, err := exec.Command("sh", "-c", fmt.Sprintf("dist fetch %s %s %s | shasum -a256", repo, tag, mediaType)).Output()
	if err != nil {
		return "", fmt.Errorf("failed to get image digest %s:%s, output: %s, err: %v", repo, tag, output, err)
	}
	return "sha256:" + string(output), nil
}

func pullImage(repo, tag string) error {
	output, err := exec.Command("sh", "-c", fmt.Sprintf("dist fetch %s %s %s | jq -r '.layers[] | \"dist fetch %s \"+.digest + \"| dist ingest --expected-digest \"+.digest+\" --expected-size \"+(.size | tostring) +\" %s@\"+.digest' | xargs -I{} -P10 -n1 sh -c \"{}\"", repo, tag, mediaType, repo, repo)).Output()
	if err != nil {
		return fmt.Errorf("failed to pull image %s:%s, output: %s, err: %v", repo, tag, output, err)
	}
	return nil
}
