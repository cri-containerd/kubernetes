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
        "containerd_container.go",
        "containerd_image.go",
        "containerd_sandbox.go",
        "containerd_service.go",
        "containerd_streaming.go",
        "doc.go",
    ],
    tags = ["automanaged"],
    deps = [
        "//pkg/kubelet/api:go_default_library",
        "//pkg/kubelet/api/v1alpha1/runtime:go_default_library",
        "//vendor:github.com/docker/containerd/api/services/content",
        "//vendor:github.com/docker/containerd/api/services/execution",
        "//vendor:github.com/docker/containerd/api/services/shim",
        "//vendor:github.com/docker/containerd/api/types/container",
        "//vendor:github.com/docker/containerd/api/types/mount",
        "//vendor:github.com/golang/glog",
        "//vendor:github.com/opencontainers/image-spec/specs-go",
        "//vendor:github.com/opencontainers/runtime-spec/specs-go",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["containerd_image_test.go"],
    library = ":go_default_library",
    tags = ["automanaged"],
    deps = [
        "//pkg/kubelet/api/v1alpha1/runtime:go_default_library",
        "//vendor:github.com/stretchr/testify/assert",
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
        "//pkg/kubelet/containerdshim/remote:all-srcs",
    ],
    tags = ["automanaged"],
)
