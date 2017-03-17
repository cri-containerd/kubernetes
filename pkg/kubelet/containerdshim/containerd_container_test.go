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
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	runtimeapi "k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/runtime"
	"k8s.io/kubernetes/pkg/kubelet/dockershim"
)

func init() {
	rand.Seed(time.Now().Unix())
}

// NOTE: To run the test, please make sure `containerd` is running and 'dist' is in $PATH.
// And you should run the test as root.
func TestContainerOperationError(t *testing.T) {
	// get the containerd client
	t.Logf("Should be able to connect with containerd")
	cdClient, err := GetContainerdClient()
	require.NoError(t, err)
	cs := NewContainerdService(cdClient)

	t.Logf("Should be able to start cs")
	err = cs.Start()
	require.NoError(t, err)

	// must fail tests TODO add more
	t.Logf("Should fail to CreateContainer with empty params")
	id, err := cs.CreateContainer("", nil, nil)
	assert.Error(t, err, "PodSandboxId should not be empty")

	t.Logf("Should fail to StartContainer with bad id")
	err = cs.StartContainer(id)
	assert.Error(t, err, "containerID should not be empty")

	t.Logf("Should fail to get ContainersStatus with bad id")
	_, err = cs.ContainerStatus(id)
	assert.Error(t, err, "containerID should not be empty")

	t.Logf("Should fail to StopContainer with bad id")
	err = cs.StopContainer(id, 0)
	assert.Error(t, err, "containerID should not be empty")

	t.Logf("Should fail to RemoveContainer with bad id")
	err = cs.RemoveContainer(id)
	assert.Error(t, err, "containerID should not be empty")
}

func TestContainerOperations(t *testing.T) {
	const (
		storePath     = ".content"
		image         = "docker.io/library/redis"
		podSandboxID  = "fake_pod"
		podName       = "name"
		podNamespace  = "namespace"
		podUID        = "uid"
		containerName = "redis"
		cwd           = "/"
		tty           = false
	)

	var (
		podAttempt       = rand.Uint32()
		containerAttempt = rand.Uint32()
		args             = []string{"redis-server", "--bind", "0.0.0.0"}
	)

	defer os.RemoveAll(storePath)
	// defer os.RemoveAll(containerdCRIRoot)

	t.Logf("Should be able to connect with containerd")
	// get the containerd client
	cdClient, err := GetContainerdClient()
	require.NoError(t, err)
	cs := NewContainerdService(cdClient)

	t.Logf("Should be able to start cs")
	require.NoError(t, cs.Start())

	t.Logf("Should be able to pull image")
	_, err = cs.PullImage(&runtimeapi.ImageSpec{Image: image}, nil)
	require.NoError(t, err)

	containerConfig := &runtimeapi.ContainerConfig{
		Metadata: &runtimeapi.ContainerMetadata{
			Name:    containerName,
			Attempt: containerAttempt,
		},
		Image: &runtimeapi.ImageSpec{Image: image},
		Args:  args,
		Tty:   tty,
	}
	sandboxConfig := &runtimeapi.PodSandboxConfig{
		Metadata: &runtimeapi.PodSandboxMetadata{
			Name:      podName,
			Uid:       podUID,
			Namespace: podNamespace,
			Attempt:   podAttempt,
		},
	} // mikebrow TODO log and console stuff
	t.Logf("Should CreateContainer")
	id, err := cs.CreateContainer(podSandboxID, containerConfig, sandboxConfig)
	require.NoError(t, err)
	t.Logf("Container id should as expected")
	containerMeta, err := dockershim.ParseContainerName(id)
	require.NoError(t, err)
	assert.Equal(t, containerConfig.Metadata, containerMeta)
	t.Logf("Container directory should be created")
	containerDir := getContainerDir(id)
	verifyFileExistence(t, true,
		containerDir,
		filepath.Join(containerDir, "rootfs"),
	)
	verifyContainerStatus(t, cs, id, containerMeta, runtimeapi.ContainerState_CONTAINER_CREATED)

	t.Logf("Should StartContainer")
	require.NoError(t, cs.StartContainer(id))
	verifyContainerStatus(t, cs, id, containerMeta, runtimeapi.ContainerState_CONTAINER_RUNNING)
	verifyFileExistence(t, true,
		filepath.Join(containerDir, "stdin"),
		filepath.Join(containerDir, "stdout"),
		filepath.Join(containerDir, "stderr"),
	)

	t.Logf("Should StopContainer")
	require.NoError(t, cs.StopContainer(id, 0))
	verifyContainerStatus(t, cs, id, containerMeta, runtimeapi.ContainerState_CONTAINER_EXITED)

	t.Logf("Should RemoveContainer")
	require.NoError(t, cs.RemoveContainer(id))
	verifyNoContainerStatus(t, cs, id)
	verifyFileExistence(t, false, filepath.Join(containerDir))
}

func verifyFileExistence(t *testing.T, expectExist bool, files ...string) {
	for _, f := range files {
		_, err := os.Stat(f)
		if expectExist {
			assert.NoError(t, err, "file %q should exist", f)
		} else {
			assert.Error(t, err, "file %q should not exist", f)
		}
	}
}

func verifyNoContainerStatus(t *testing.T, cs ContainerdService, id string) {
	t.Logf("Container should not show up in list containers")
	cntrs, err := cs.ListContainers(nil)
	require.NoError(t, err)
	var cntr *runtimeapi.Container
	for _, c := range cntrs {
		if c.Id == id {
			cntr = c
			break
		}
	}
	require.Nil(t, cntr)

	t.Logf("Container status should not show up")
	_, err = cs.ContainerStatus(id)
	require.Error(t, err)
}

func verifyContainerStatus(t *testing.T, cs ContainerdService, id string, metadata *runtimeapi.ContainerMetadata, state runtimeapi.ContainerState) {
	t.Logf("Container should show up in list containers")
	cntrs, err := cs.ListContainers(nil)
	require.NoError(t, err)
	var cntr *runtimeapi.Container
	for _, c := range cntrs {
		if c.Id == id {
			cntr = c
			break
		}
	}
	require.NotNil(t, cntr)
	assert.Equal(t, state, cntr.State)
	assert.Equal(t, metadata, cntr.Metadata)

	t.Logf("Container status should be expected")
	status, err := cs.ContainerStatus(id)
	require.NoError(t, err)
	assert.Equal(t, id, status.Id)
	assert.Equal(t, state, status.State)
	assert.Equal(t, metadata, status.Metadata)
}
