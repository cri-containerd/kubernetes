package containerdshim

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	specs "github.com/opencontainers/runtime-spec/specs-go"

	"k8s.io/kubernetes/pkg/kubelet/leaky"
)

const (
	// kubePrefix is used to identify the containers/sandboxes on the node managed by kubelet
	kubePrefix = "k8s"
	// sandboxContainerName is a string to include in the docker container so
	// that users can easily identify the sandboxes.
	sandboxContainerName = leaky.PodInfraContainerName
	// Delimiter used to construct docker container names.
	nameDelimiter = "_"
)

var rwm = "rwm"

const rootfsPath = "rootfs"

func spec(id string, args []string, tty bool) *specs.Spec {
	return &specs.Spec{
		Version: specs.Version,
		Platform: specs.Platform{
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
		},
		Root: specs.Root{
			Path:     rootfsPath,
			Readonly: true,
		},
		Process: specs.Process{
			Args: args,
			Env: []string{
				"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
			},
			Terminal:        tty,
			Cwd:             "/",
			NoNewPrivileges: true,
		},
		Mounts: []specs.Mount{
			{
				Destination: "/proc",
				Type:        "proc",
				Source:      "proc",
			},
			{
				Destination: "/dev",
				Type:        "tmpfs",
				Source:      "tmpfs",
				Options:     []string{"nosuid", "strictatime", "mode=755", "size=65536k"},
			},
			{
				Destination: "/dev/pts",
				Type:        "devpts",
				Source:      "devpts",
				Options:     []string{"nosuid", "noexec", "newinstance", "ptmxmode=0666", "mode=0620", "gid=5"},
			},
			{
				Destination: "/dev/shm",
				Type:        "tmpfs",
				Source:      "shm",
				Options:     []string{"nosuid", "noexec", "nodev", "mode=1777", "size=65536k"},
			},
			{
				Destination: "/dev/mqueue",
				Type:        "mqueue",
				Source:      "mqueue",
				Options:     []string{"nosuid", "noexec", "nodev"},
			},
			{
				Destination: "/sys",
				Type:        "sysfs",
				Source:      "sysfs",
				Options:     []string{"nosuid", "noexec", "nodev"},
			},
			{
				Destination: "/run",
				Type:        "tmpfs",
				Source:      "tmpfs",
				Options:     []string{"nosuid", "strictatime", "mode=755", "size=65536k"},
			},
			{
				Destination: "/etc/resolv.conf",
				Type:        "bind",
				Source:      "/etc/resolv.conf",
				Options:     []string{"rbind", "ro"},
			},
			{
				Destination: "/etc/hosts",
				Type:        "bind",
				Source:      "/etc/hosts",
				Options:     []string{"rbind", "ro"},
			},
			{
				Destination: "/etc/localtime",
				Type:        "bind",
				Source:      "/etc/localtime",
				Options:     []string{"rbind", "ro"},
			},
		},
		Hostname: id,
		Linux: &specs.Linux{
			Resources: &specs.LinuxResources{
				Devices: []specs.LinuxDeviceCgroup{
					{
						Allow:  false,
						Access: &rwm,
					},
				},
			},
			Namespaces: []specs.LinuxNamespace{
				{
					Type: "pid",
				},
				{
					Type: "ipc",
				},
				{
					Type: "uts",
				},
				{
					Type: "mount",
				},
				{
					Type: "network",
				},
			},
		},
	}
}

func getTempDir(id string) (string, error) {
	err := os.MkdirAll(filepath.Join(os.TempDir(), "ctr"), 0700)
	if err != nil {
		return "", err
	}
	tmpDir, err := ioutil.TempDir(filepath.Join(os.TempDir(), "ctr"), fmt.Sprintf("%s-", id))
	if err != nil {
		return "", err
	}
	return tmpDir, nil
}
