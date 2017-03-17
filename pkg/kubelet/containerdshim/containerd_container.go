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
	"os"
	"path/filepath"

	"github.com/docker/containerd/api/services/execution"
	"github.com/docker/containerd/api/services/shim"
	"github.com/docker/containerd/api/types/mount"
	protobuf "github.com/gogo/protobuf/types"
	"github.com/golang/glog"
	specs "github.com/opencontainers/runtime-spec/specs-go"

	runtimeapi "k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/runtime"
	"k8s.io/kubernetes/pkg/kubelet/dockershim"
)

const (
	containerdCRIRoot = "/tmp/containerd-cri"
	shimbindSocket    = "shim.sock"
)

// P0
func (cs *containerdService) ListContainers(filter *runtimeapi.ContainerFilter) ([]*runtimeapi.Container, error) {
	resp, err := cs.cdClient.List(gocontext.Background(), &execution.ListRequest{})
	if err != nil {
		return nil, err
	}
	var criContainers []*runtimeapi.Container
	for _, c := range resp.Containers {
		criC, err := toCRIContainer(c)
		if err != nil {
			glog.Errorf("Failed to parse container %+v: %v", c, err)
			continue
		}
		criContainers = append(criContainers, criC)
	}

	// Only support state filter now.
	// TODO Add other filters, especially label filter.
	if filter != nil {
		var filtered []*runtimeapi.Container
		if filter.State != nil {
			for _, c := range criContainers {
				if c.State == filter.GetState().State {
					filtered = append(filtered, c)
				}
			}
		}
		criContainers = filtered
	}
	return criContainers, nil
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

	// TODO(P0): Current CRI integration highly rely on label filter.
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
	// mikebrow TODO containerID must be unique crio guys are using stringid.GenerateNonCryptoID() then insuring uniqueness with storage
	containerID := dockershim.MakeContainerName(sandboxConfig, containerConfig)

	containerDir, err := ensureContainerDir(containerID)
	if err != nil {
		return "", err
	}

	// Create container rootfs
	rootfsPath := filepath.Join(containerDir, "rootfs")
	if err := createRootfs(containerConfig.GetImage().GetImage(), rootfsPath); err != nil {
		return "", err
	}

	// mikebrow for now configure to bind mount the rootfs
	rootfs := []*mount.Mount{
		{
			Type:   "bind",
			Source: rootfsPath,
			Options: []string{
				"rw",
				"rbind",
			},
		},
	}

	var processArgs []string
	if containerConfig.Command != nil {
		processArgs = append(processArgs, containerConfig.Command...)
	}
	if containerConfig.Args != nil {
		processArgs = append(processArgs, containerConfig.Args...)
	}

	// TODO: Set other configs, such as envs, working directory etc.
	s := defaultOCISpec(containerID, processArgs, rootfsPath, containerConfig.Tty)

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
		Stdin:    filepath.Join(containerDir, "stdin"), // mikebrow TODO needed for console
		Stdout:   filepath.Join(containerDir, "stdout"),
		Stderr:   filepath.Join(containerDir, "stderr"),
	}
	_, err = prepareStdio(create.Stdin, create.Stdout, create.Stderr, create.Terminal)
	if err != nil {
		return "", err
	}

	// mikebrow TODO proper console handling
	glog.V(2).Infof("CreateContainer for container %s container directory %s", containerID, containerDir)
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
	_, err := cs.cdClient.Start(gocontext.Background(), &execution.StartRequest{ID: containerID})
	return err
}

// StopContainer stops a running container with a grace period (i.e., timeout).
// P0
func (cs *containerdService) StopContainer(containerID string, timeout int64) error {
	glog.V(2).Infof("StopContainer called with %s", containerID)
	// TODO Support grace period.
	if containerID == "" {
		return fmt.Errorf("containerID should not be empty")
	}
	c, err := getShimClient(containerID)
	if err != nil {
		return fmt.Errorf("failed to get shim for container %q: %v", containerID, err)
	}
	_, err = c.Exit(gocontext.Background(), &shim.ExitRequest{})
	return err
}

// RemoveContainer removes the container. If the container is running, the container
// should be force removed.
// P1
func (cs *containerdService) RemoveContainer(containerID string) error {
	glog.V(2).Infof("RemoveContainer called with %s", containerID)
	if containerID == "" {
		return fmt.Errorf("containerID should not be empty")
	}
	_, err := cs.cdClient.Delete(gocontext.Background(), &execution.DeleteRequest{ID: containerID})
	if err != nil {
		return err
	}
	// TODO Support log, keep log and remove container
	return os.RemoveAll(getContainerDir(containerID))
}

// ContainerStatus returns status of the container.
// P0
func (cs *containerdService) ContainerStatus(containerID string) (*runtimeapi.ContainerStatus, error) {
	glog.V(4).Infof("ContainerStatus called with %s", containerID)
	if containerID == "" {
		return nil, fmt.Errorf("containerID should not be empty")
	}
	c, err := cs.cdClient.Info(gocontext.Background(), &execution.InfoRequest{ID: containerID})
	if err != nil {
		return nil, fmt.Errorf("failed to get container info %q: %v", containerID, err)
	}
	return toCRIContainerStatus(c)
}
