/*
Copyright 2017 The Kubernetes Authors.

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
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	runtimeapi "k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/runtime"
)

// NOTE: To run the test, please make sure `dist` and `jq` is in $PATH.
func TestImageOperations(t *testing.T) {
	const (
		storePath = ".content"
		image     = "docker.io/library/redis"
	)
	// Make sure there is no exsiting image.
	if _, err := os.Stat(storePath); err == nil {
		os.RemoveAll(storePath)
	}
	cs := NewContainerdService(nil)

	t.Logf("Should be able to pull image")
	digest, err := cs.PullImage(&runtimeapi.ImageSpec{Image: image}, nil)
	assert.NoError(t, err)
	t.Logf("Should be able to list new images")
	imgs, err := cs.ListImages(nil)
	assert.NoError(t, err)
	assert.Len(t, imgs, 1)
	assert.Equal(t, digest, imgs[0].Id)
	t.Logf("Should be able to get new image status with name")
	img, err := cs.ImageStatus(&runtimeapi.ImageSpec{Image: image})
	assert.NoError(t, err)
	assert.Equal(t, imgs[0], img)
	t.Logf("Should be able to get new image status with digest")
	img, err = cs.ImageStatus(&runtimeapi.ImageSpec{Image: digest})
	assert.NoError(t, err)
	assert.Equal(t, imgs[0], img)
	t.Logf("Should be able to see the image in content store")
	_, err = os.Stat(storePath)
	assert.NoError(t, err)
	output, err := exec.Command("sh", "-c", fmt.Sprintf("ls %s/blobs | wc -l", storePath)).Output()
	assert.NoError(t, err)
	layerNum, err := strconv.Atoi(strings.TrimSpace(string(output)))
	assert.NoError(t, err)
	assert.NotZero(t, layerNum)

	t.Logf("Should have the same digest and pull no new layer if we pull the same image")
	newDigest, err := cs.PullImage(&runtimeapi.ImageSpec{Image: image}, nil)
	assert.NoError(t, err)
	assert.Equal(t, digest, newDigest)
	output, err = exec.Command("sh", "-c", fmt.Sprintf("ls %s/blobs | wc -l", storePath)).Output()
	assert.NoError(t, err)
	newLayerNum, err := strconv.Atoi(strings.TrimSpace(string(output)))
	assert.NoError(t, err)
	assert.Equal(t, layerNum, newLayerNum)

	t.Logf("Should be able to remove image")
	err = cs.RemoveImage(&runtimeapi.ImageSpec{Image: digest})
	assert.NoError(t, err)
	imgs, err = cs.ListImages(nil)
	assert.NoError(t, err)
	assert.Empty(t, imgs)
	img, err = cs.ImageStatus(&runtimeapi.ImageSpec{Image: image})
	assert.NoError(t, err)
	assert.Nil(t, img)

	os.RemoveAll(storePath)
}

// The test must be run as root, because apply layer needs the permission
// to change diretory owner.
func TestCreateRootfs(t *testing.T) {
	const (
		storePath = ".content"
		image     = "docker.io/library/redis"
		rootfs    = "rootfs"
	)
	// Make sure there is no exsiting image.
	if _, err := os.Stat(storePath); err == nil {
		os.RemoveAll(storePath)
	}
	cs := NewContainerdService(nil)

	t.Logf("Should be able to pull image")
	_, err := cs.PullImage(&runtimeapi.ImageSpec{Image: image}, nil)
	assert.NoError(t, err)
	t.Logf("Should be able to create rootfs from the image")
	repo, tag := repoAndTag(image)
	assert.NoError(t, createRootfs(repo, tag, rootfs))
	t.Logf("The rootfs should be created")
	_, err = os.Stat(rootfs)
	assert.NoError(t, err)
	output, err := exec.Command("sh", "-c", fmt.Sprintf("ls %s rootfs | wc -l", rootfs)).Output()
	assert.NoError(t, err)
	dirsNum, err := strconv.Atoi(strings.TrimSpace(string(output)))
	assert.NoError(t, err)
	assert.NotZero(t, dirsNum)

	os.RemoveAll(rootfs)
	os.RemoveAll(storePath)
}
