package(default_visibility = ["//visibility:public"])

licenses(["notice"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
    "go_test",
)

go_library(
    name = "go_default_library",
    srcs = [
        "client_linux.go",
        "client_liverestore_linux.go",
        "client_shutdownrestore_linux.go",
        "client.go",
        "container_linux.go",
        "container.go",
        "pausemonitor_linux.go",
        "process_linux.go",
        "process.go",
        "queue_linux.go",
        "remote_linux.go",
        "remote.go",
        "types_linux.go",
        "types.go",
        "utils_linux.go",
    ],
    tags = ["automanaged"],
    deps = [
        "//pkg/api:go_default_library",
        "//pkg/api/v1:go_default_library",
        "//pkg/credentialprovider:go_default_library",
        "//pkg/kubelet/cm:go_default_library",
        "//pkg/kubelet/container:go_default_library",
        "//pkg/kubelet/custommetrics:go_default_library",
        "//pkg/kubelet/dockertools/securitycontext:go_default_library",
        "//pkg/kubelet/events:go_default_library",
        "//pkg/kubelet/images:go_default_library",
        "//pkg/kubelet/leaky:go_default_library",
        "//pkg/kubelet/lifecycle:go_default_library",
        "//pkg/kubelet/metrics:go_default_library",
        "//pkg/kubelet/network:go_default_library",
        "//pkg/kubelet/network/hairpin:go_default_library",
        "//pkg/kubelet/prober/results:go_default_library",
        "//pkg/kubelet/qos:go_default_library",
        "//pkg/kubelet/types:go_default_library",
        "//pkg/kubelet/util/cache:go_default_library",
        "//pkg/kubelet/util/format:go_default_library",
        "//pkg/security/apparmor:go_default_library",
        "//pkg/securitycontext:go_default_library",
        "//pkg/util/exec:go_default_library",
        "//pkg/util/oom:go_default_library",
        "//pkg/util/procfs:go_default_library",
        "//pkg/util/selinux:go_default_library",
        "//pkg/util/strings:go_default_library",
        "//pkg/util/tail:go_default_library",
        "//pkg/util/term:go_default_library",
        "//pkg/util/version:go_default_library",
        "//vendor:github.com/armon/circbuf",
        "//vendor:github.com/docker/distribution/digest",
        "//vendor:github.com/docker/distribution/reference",
        "//vendor:github.com/docker/docker/pkg/jsonmessage",
        "//vendor:github.com/docker/docker/pkg/stdcopy",
        "//vendor:github.com/docker/engine-api/client",
        "//vendor:github.com/docker/engine-api/types",
        "//vendor:github.com/docker/engine-api/types/container",
        "//vendor:github.com/docker/engine-api/types/strslice",
        "//vendor:github.com/docker/engine-api/types/versions",
        "//vendor:github.com/docker/go-connections/nat",
        "//vendor:github.com/golang/glog",
        "//vendor:github.com/google/cadvisor/info/v1",
        "//vendor:golang.org/x/net/context",
        "//vendor:k8s.io/apimachinery/pkg/apis/meta/v1",
        "//vendor:k8s.io/apimachinery/pkg/runtime",
        "//vendor:k8s.io/apimachinery/pkg/runtime/schema",
        "//vendor:k8s.io/apimachinery/pkg/types",
        "//vendor:k8s.io/apimachinery/pkg/util/errors",
        "//vendor:k8s.io/apimachinery/pkg/util/runtime",
        "//vendor:k8s.io/apimachinery/pkg/util/sets",
        "//vendor:k8s.io/client-go/tools/record",
        "//vendor:k8s.io/client-go/util/clock",
        "//vendor:k8s.io/client-go/util/flowcontrol",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
    ],
    data = [
    ],
    library = ":go_default_library",
    tags = [
        "automanaged",
    ],
    deps = [
    ],
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [
        ":package-srcs",
    ],
    tags = ["automanaged"],
)
