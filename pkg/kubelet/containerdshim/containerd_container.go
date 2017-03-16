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
	gocontext "context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/docker/containerd/api/services/execution"
	"github.com/docker/containerd/api/types/mount"
	protobuf "github.com/gogo/protobuf/types"
	"github.com/golang/glog"
	specs "github.com/opencontainers/runtime-spec/specs-go"

	runtimeapi "k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/runtime"
)

// P0
func (cs *containerdService) ListContainers(filter *runtimeapi.ContainerFilter) ([]*runtimeapi.Container, error) {
	return nil, fmt.Errorf("not implemented")
}

// CreateContainer creates a new container in the given PodSandbox
// P0
func (cs *containerdService) CreateContainer(podSandboxID string, containerConfig *runtimeapi.ContainerConfig, sandboxConfig *runtimeapi.PodSandboxConfig) (string, error) {
	glog.V(2).Infof("CreateContainer for pod %s", podSandboxID)
	if podSandboxID == "" {
		return "", fmt.Errorf("PodSandboxId should not be empty")
	}
	if containerConfig == nil {
		return "", fmt.Errorf("container config is nil")
	}
	if sandboxConfig == nil {
		return "", fmt.Errorf("sandbox config is nil for container %q", containerConfig.Metadata.Name)
	}

	// mikebrow todo lookup the pod by id (for poc going without a pod datastructure)

	// mikebrow todo labels and annotations
	// labels := makeLabels(config.GetLabels(), config.GetAnnotations())
	// Apply a the container type label.
	// labels[containerTypeLabelKey] = containerTypeLabelContainer
	// Write the container log path in the labels.
	// labels[containerLogPathLabelKey] = filepath.Join(sandboxConfig.LogDirectory, config.LogPath)
	// Write the sandbox ID in the labels.
	// labels[sandboxIDLabelKey] = podSandboxID

	name := containerConfig.GetMetadata().Name
	if name == "" {
		return "", fmt.Errorf("CreateContainerRequest.ContainerConfig.Name is empty")
	}
	attempt := containerConfig.GetMetadata().Attempt
	var containerName string
	if name == "infra" {
		containerName = fmt.Sprintf("%s-%s", podSandboxID, name)
	} else {
		containerName = fmt.Sprintf("%s-%s-%v", podSandboxID, name, attempt)
	}
	containerID := podSandboxID // mikebrow TODO containerID must be unique crio guys are using stringid.GenerateNonCryptoID() then insuring uniqueness with storage

	tmpDir, err := getTempDir(containerName)
	if err != nil {
		return "", err
	}

	abs, err := filepath.Abs(containerConfig.WorkingDir)
	if err != nil {
		return "", err
	}

	// mikebrow for now configure to bind mount the rootfs
	rootfs := []*mount.Mount{
		{
			Type:   "bind",
			Source: abs,
			Options: []string{
				"rw",
				"rbind",
			},
		},
	}

	processArgs := []string{}
	commands := containerConfig.Command
	args := containerConfig.Args
	if commands == nil && args == nil {
		processArgs = nil
	}
	if commands != nil {
		processArgs = append(processArgs, commands...)
	}
	if args != nil {
		processArgs = append(processArgs, args...)
	}

	s := spec(containerID, processArgs, containerConfig.Tty)

	data, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	create := &execution.CreateRequest{
		ID: containerID,
		Spec: &protobuf.Any{
			TypeUrl: specs.Version,
			Value:   data,
		},
		Rootfs:   rootfs,
		Runtime:  "linux",
		Terminal: containerConfig.Tty,
		Stdin:    filepath.Join(tmpDir, "stdin"),
		Stdout:   filepath.Join(tmpDir, "stdout"),
		Stderr:   filepath.Join(tmpDir, "stderr"),
	}

	// mikebrow todo proper console handling

	response, err := cs.cdClient.Create(gocontext.Background(), create)
	if err != nil {
		return "", err
	}

	return response.ID, nil
}

// StartContainer starts the container.
// P0
func (cs *containerdService) StartContainer(containerID string) error {
	glog.V(2).Infof("StartContainer called with %s", containerID)
	if containerID == "" {
		return fmt.Errorf("containerID should not be empty")
	}
	if _, err := cs.cdClient.Start(gocontext.Background(), &execution.StartRequest{
		ID: containerID,
	}); err != nil {
		glog.V(2).Infof("StartContainer for %s failed with %v", containerID, err)
		return err
	}
	return nil
}

// StopContainer stops a running container with a grace period (i.e., timeout).
// P0
func (cs *containerdService) StopContainer(containerID string, timeout int64) error {
	return fmt.Errorf("not implemented")
}

// RemoveContainer removes the container. If the container is running, the container
// should be force removed.
// P1
func (cs *containerdService) RemoveContainer(containerID string) error {
	return fmt.Errorf("not implemented")
}

// ContainerStatus returns status of the container.
// P0
func (cs *containerdService) ContainerStatus(containerID string) (*runtimeapi.ContainerStatus, error) {
	return nil, fmt.Errorf("not implemented")
}
