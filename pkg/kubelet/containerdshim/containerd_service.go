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
	"net"
	"net/http"
	"time"

	"github.com/golang/glog"
	"google.golang.org/grpc"

	internalapi "k8s.io/kubernetes/pkg/kubelet/api"
	runtimeapi "k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/runtime"

	execution "github.com/docker/containerd/api/services/execution"
	_ "github.com/docker/containerd/api/services/shim"
	_ "github.com/docker/containerd/api/types/container"
	_ "github.com/docker/containerd/api/types/mount"
	_ "github.com/opencontainers/image-spec/specs-go"
	_ "github.com/opencontainers/runtime-spec/specs-go"
)

type ContainerdService interface {
	internalapi.RuntimeService
	internalapi.ImageManagerService
	Start() error
	// For serving streaming calls.
	http.Handler
}

type containerdService struct {
	// containerd client
	cdClient execution.ContainerServiceClient
}

func NewContainerdService(cdClient execution.ContainerServiceClient) ContainerdService {
	return &containerdService{cdClient: cdClient}
}

// The unix socket for containerdshhim <-> containerd communication.
const containerdBindSocket = "/run/containerd/containerd.sock" // mikebrow TODO get these from a config

// GetContainerdClient returns a grpc client for containerd exection service.
func GetContainerdClient() (execution.ContainerServiceClient, error) {
	// get the containerd client
	dialOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(100 * time.Second),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", containerdBindSocket, timeout)
		}),
	}
	conn, err := grpc.Dial(fmt.Sprintf("unix://%s", containerdBindSocket), dialOpts...)
	if err != nil {
		return nil, err
	}
	return execution.NewContainerServiceClient(conn), nil
}

// P4
func (cs *containerdService) Version(_ string) (*runtimeapi.VersionResponse, error) {
	return &runtimeapi.VersionResponse{
		Version:           "0.1.0",
		RuntimeName:       "containerd-poc",
		RuntimeVersion:    "1.0.0",
		RuntimeApiVersion: "1.0.0",
	}, nil
}

func (cs *containerdService) Start() error {
	glog.V(2).Infof("Start containerd service")
	return nil
}

// P4
func (cs *containerdService) UpdateRuntimeConfig(runtimeConfig *runtimeapi.RuntimeConfig) error {
	return nil
}

// P4
func (cs *containerdService) Status() (*runtimeapi.RuntimeStatus, error) {
	runtimeReady := &runtimeapi.RuntimeCondition{
		Type:   runtimeapi.RuntimeReady,
		Status: true,
	}
	networkReady := &runtimeapi.RuntimeCondition{
		Type:   runtimeapi.NetworkReady,
		Status: true,
	}
	return &runtimeapi.RuntimeStatus{Conditions: []*runtimeapi.RuntimeCondition{runtimeReady, networkReady}}, nil
}

// P3
func (cs *containerdService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}
