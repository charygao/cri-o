package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
    "go_test",
)

go_test(
    name = "go_default_test",
    srcs = ["csr_test.go"],
    data = glob(["testdata/**"]),
    embed = [":go_default_library"],
    deps = ["//staging/src/k8s.io/client-go/util/keyutil:go_default_library"],
)

go_library(
    name = "go_default_library",
    srcs = [
        "cert.go",
        "csr.go",
        "io.go",
        "pem.go",
        "server_inspection.go",
    ],
    importmap = "k8s.io/kubernetes/vendor/k8s.io/client-go/util/cert",
    importpath = "k8s.io/client-go/util/cert",
    deps = ["//staging/src/k8s.io/client-go/util/keyutil:go_default_library"],
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [":package-srcs"],
    tags = ["automanaged"],
)
