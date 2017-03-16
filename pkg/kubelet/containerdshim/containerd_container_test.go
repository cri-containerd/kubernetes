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
	"net"
	"os"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/docker/containerd/api/services/execution"
	"github.com/stretchr/testify/assert"

	runtimeapi "k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/runtime"
)

// NOTE: To run the test, please make sure `containerd` is running and 'dist' is in $PATH.
// And you should run the test as root.
func TestContainer(t *testing.T) {
	const (
		storePath = ".content"
		image     = "docker.io/library/redis"
	)
	// Make sure there is no exsiting image.
	if _, err := os.Stat(storePath); err == nil {
		os.RemoveAll(storePath)
	}

	const (
		// The unix socket for containerdshhim <-> containerd communication.
		bindSocket = "/run/containerd/containerd.sock" // mikebrow get these from a config
	)

	//	cmd := exec.Command("sh", "-c", "containerd")
	//	cmd.Start()

	t.Logf("Should be able to connect with containerd")
	// get the containerd client
	dialOpts := []grpc.DialOption{grpc.WithInsecure(), grpc.WithTimeout(100 * time.Second)}
	dialOpts = append(dialOpts, grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
		return net.DialTimeout("unix", bindSocket, timeout)
	}))
	conn, err := grpc.Dial(fmt.Sprintf("unix://%s", bindSocket), dialOpts...)
	assert.NoError(t, err)

	cdClient := execution.NewContainerServiceClient(conn)
	cs := NewContainerdService(cdClient)

	t.Logf("Should be able to start cs")
	erra := cs.Start()
	assert.NoError(t, erra)

	t.Logf("Should be able to pull image")
	_, errb := cs.PullImage(&runtimeapi.ImageSpec{Image: image}, nil)
	assert.NoError(t, errb)

	// must fail tests TODO add more
	t.Logf("Should fail to CreateContainer with empty params")
	id, errc := cs.CreateContainer("", nil, nil)
	assert.Error(t, errc, "PodSandboxId should not be empty")

	t.Logf("Should fail to StartContainer with bad id")
	errd := cs.StartContainer(id)
	assert.Error(t, errd, "containerID should not be empty")

	t.Logf("Should fail to get ContainersStatus with bad id")
	_, erre := cs.ContainerStatus(id)
	assert.Error(t, erre, "not implemented") // mikebrow fix when implemented

	t.Logf("Should fail to ListContainers with empty filter")
	_, errf := cs.ListContainers(nil)
	assert.Error(t, errf, "not implemented") // mikebrow fix when implemented

	var timeout int64
	timeout = 0
	t.Logf("Should fail to StopContainer with bad id")
	errg := cs.StopContainer(id, timeout)
	assert.Error(t, errg, "not implemented") // mikebrow fix when implemented

	t.Logf("Should fail to RemoveContainer with bad id")
	errh := cs.RemoveContainer(id)
	assert.Error(t, errh, "not implemented") // mikebrow fix when implemented

	// must pass tests
	// mikebrow TODO

	//	t.Logf("Should be able to kill containerd")
	//	erri := cmd.Process.Kill()
	//	assert.NoError(t, erri)

	t.Logf("Should be able to clean up storage")
	errj := os.RemoveAll(storePath)
	assert.NoError(t, errj)
}
