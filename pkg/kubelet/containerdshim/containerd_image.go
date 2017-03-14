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

// P0
func (cs *containerdService) ListImages(filter *runtimeapi.ImageFilter) ([]*runtimeapi.Image, error) {
	return nil, fmt.Errorf("not implemented")
}

// P0
func (cs *containerdService) ImageStatus(image *runtimeapi.ImageSpec) (*runtimeapi.Image, error) {
	return nil, fmt.Errorf("not implemented")
}

// P0
func (cs *containerdService) PullImage(image *runtimeapi.ImageSpec, auth *runtimeapi.AuthConfig) (string, error) {
	return "", fmt.Errorf("not implemented")
}

// P1
func (cs *containerdService) RemoveImage(image *runtimeapi.ImageSpec) error {
	return fmt.Errorf("not implemented")
}
