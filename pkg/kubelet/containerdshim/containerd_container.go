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
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/docker/containerd/api/services/execution"
	"github.com/docker/containerd/api/types/mount"
	protobuf "github.com/gogo/protobuf/types"
	"github.com/golang/glog"
	specs "github.com/opencontainers/runtime-spec/specs-go"

	runtimeapi "k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/runtime"
	"k8s.io/kubernetes/pkg/kubelet/dockershim"
)

// containerStore is used to store container metadata.
// TODO: Consider to checkpoint ourselves or use containerd metadata store.
var containerStore map[string]*runtimeapi.ContainerStatus = map[string]*runtimeapi.ContainerStatus{}
var containerStoreLock sync.RWMutex

// P0
func (cs *containerdService) ListContainers(filter *runtimeapi.ContainerFilter) ([]*runtimeapi.Container, error) {
	containerStoreLock.RLock()
	defer containerStoreLock.RUnlock()
	resp, err := cs.containerService.List(gocontext.Background(), &execution.ListRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers from containerd: %v", err)
	}

	var containers []*runtimeapi.Container
	for _, status := range containerStore {
		container := statusToContainer(status)
		// Set default state as exited, because any container without running
		// containerd container is in exited state.
		container.State = runtimeapi.ContainerState_CONTAINER_EXITED
		for _, c := range resp.Containers {
			if c.ID == container.Id {
				container.State = toCRIContainerState(c.Status)
				break
			}
		}
		containers = append(containers, container)
	}

	// Only support state filter now.
	// TODO Add other filters, especially label filter.
	if filter != nil {
		var filtered []*runtimeapi.Container
		if filter.State != nil {
			for _, c := range containers {
				if c.State == filter.GetState().State {
					filtered = append(filtered, c)
				}
			}
		}
		containers = filtered
	}
	return containers, nil
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
		return "", fmt.Errorf("sandbox config is nil for container %q", containerConfig.GetMetadata().GetName())
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

	name := containerConfig.GetMetadata().GetName()
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
	if err := cs.createRootfs(containerConfig.GetImage().GetImage(), rootfsPath); err != nil {
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
	if containerConfig.GetCommand() != nil {
		processArgs = append(processArgs, containerConfig.GetCommand()...)
	}
	if containerConfig.GetArgs() != nil {
		processArgs = append(processArgs, containerConfig.GetArgs()...)
	}

	// TODO: Set other configs, such as envs, working directory etc.
	s := defaultOCISpec(containerID, processArgs, rootfsPath, containerConfig.GetTty())

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
		Terminal: containerConfig.GetTty(),
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
	response, err := cs.containerService.Create(gocontext.Background(), create)
	if err != nil {
		return "", err
	}

	containerStoreLock.Lock()
	defer containerStoreLock.Unlock()
	containerStore[containerID] = &runtimeapi.ContainerStatus{
		Id:          containerID,
		Metadata:    containerConfig.GetMetadata(),
		CreatedAt:   time.Now().Unix(),
		Image:       containerConfig.GetImage(),
		ImageRef:    isImagePulled(containerConfig.GetImage().GetImage()),
		Labels:      containerConfig.GetLabels(),
		Annotations: containerConfig.GetAnnotations(),
		Mounts:      containerConfig.GetMounts(),
	}
	return response.ID, nil
}

// StartContainer starts the container.
// P0
func (cs *containerdService) StartContainer(containerID string) error {
	glog.V(2).Infof("StartContainer called with %s", containerID)
	if _, ok := containerStore[containerID]; !ok {
		return fmt.Errorf("container not found %s", containerID)
	}
	_, err := cs.containerService.Start(gocontext.Background(), &execution.StartRequest{ID: containerID})
	if err != nil {
		return err
	}
	containerStore[containerID].StartedAt = time.Now().Unix()
	return nil
}

// StopContainer stops a running container with a grace period (i.e., timeout).
// P0
func (cs *containerdService) StopContainer(containerID string, timeout int64) error {
	glog.V(2).Infof("StopContainer called with %s", containerID)
	// TODO Support grace period.
	if _, ok := containerStore[containerID]; !ok {
		return fmt.Errorf("container not found %s", containerID)
	}
	_, err := cs.containerService.Delete(gocontext.Background(), &execution.DeleteRequest{ID: containerID})
	if err != nil {
		return err
	}
	containerStore[containerID].FinishedAt = time.Now().Unix()
	return err
}

// RemoveContainer removes the container. If the container is running, the container
// should be force removed.
// P1
func (cs *containerdService) RemoveContainer(containerID string) error {
	glog.V(2).Infof("RemoveContainer called with %s", containerID)
	if _, ok := containerStore[containerID]; !ok {
		return fmt.Errorf("container not found %s", containerID)
	}
	// TODO Support log, keep log and remove container
	containerDir := getContainerDir(containerID)
	rootfsPath := filepath.Join(containerDir, "rootfs")
	if err := exec.Command("umount", rootfsPath).Run(); err != nil {
		return fmt.Errorf("failed to umount rootfs %s: %v", rootfsPath, err)
	}
	if err := os.RemoveAll(containerDir); err != nil {
		return err
	}
	delete(containerStore, containerID)
	return nil
}

// ContainerStatus returns status of the container.
// P0
func (cs *containerdService) ContainerStatus(containerID string) (*runtimeapi.ContainerStatus, error) {
	glog.V(4).Infof("ContainerStatus called with %s", containerID)
	status, ok := containerStore[containerID]
	if !ok {
		return nil, fmt.Errorf("container not found %v", containerID)
	}
	status.State = runtimeapi.ContainerState_CONTAINER_EXITED
	c, err := cs.containerService.Info(gocontext.Background(), &execution.InfoRequest{ID: containerID})
	if err != nil {
		if !strings.Contains(err.Error(), "container does not exist") {
			return nil, fmt.Errorf("failed to get container info %q: %v", containerID, err)
		}
		return status, nil
	}
	status.State = toCRIContainerState(c.Status)
	return status, nil
}
