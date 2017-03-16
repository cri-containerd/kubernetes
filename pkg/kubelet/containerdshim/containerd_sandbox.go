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

	runtimeapi "k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/runtime"
)

// RunPodSandbox creates and runs a pod-level sandbox.
// P0
func (cs *containerdService) RunPodSandbox(config *runtimeapi.PodSandboxConfig) (string, error) {
	return "", fmt.Errorf("not implemented")
}

// StopPodSandbox stops the sandbox. If there are any running containers in the
// sandbox, they should be force terminated.
// P0
func (cs *containerdService) StopPodSandbox(podSandboxID string) error {
	return fmt.Errorf("not implemented")
}

// RemovePodSandbox deletes the sandbox. If there are any running containers in the
// sandbox, they should be force deleted.
// P1
func (cs *containerdService) RemovePodSandbox(podSandboxID string) error {
	return fmt.Errorf("not implemented")
}

// PodSandboxStatus returns the Status of the PodSandbox.
// P0
func (cs *containerdService) PodSandboxStatus(podSandboxID string) (*runtimeapi.PodSandboxStatus, error) {
	return nil, fmt.Errorf("not implemented")
}

// ListPodSandbox returns a list of SandBoxes.
// P0
func (cs *containerdService) ListPodSandbox(filter *runtimeapi.PodSandboxFilter) ([]*runtimeapi.PodSandbox, error) {
	return nil, fmt.Errorf("not implemented")
}
