load("@io_bazel_rules_docker//go:image.bzl", "go_image")
load("@io_bazel_rules_docker//container:container.bzl", "container_image")

BASE_IMAGE = select({
    "//build/docker:fastbuild_arm64": "@image_static_arm64//image",
    "//build/docker:opt_arm64": "@image_static_arm64//image",
    "//build/docker:debug_arm64": "@image_static_arm64_debug//image",
    "//build/docker:fastbuild_x86_64": "@image_static_x86_64//image",
    "//build/docker:opt_x86_64": "@image_static_x86_64//image",
    "//build/docker:debug_x86_64": "@image_static_x86_64_debug//image",
})

# idc_go_image uses an appropriate base image based on bazel
# compilation_mode and x86_64 v. arm64
def idc_go_image(name, base = BASE_IMAGE, **kwargs):
    go_image(
        name = name,
        base = base,
        **kwargs
    )

# idc_container_image uses an appropriate base image based on bazel
# compilation_mode and x86_64 v. arm64
def idc_container_image(name, base = BASE_IMAGE, **kwargs):
    container_image(
        name = name,
        base = base,
        **kwargs
    )
