def register_idc_toolchains():
    native.register_toolchains(
        # A dummy toolchain that enables compiling Go on MacOS for Linux without
        # setting up a real C++ toolchain. This should no longer be required once
        # rules_go supports optional toolchains, which are to be added in Bazel v6.
        "//build/toolchains:macos_dummy_cpp_toolchain_linux_x86",
        "//build/toolchains:macos_dummy_cpp_toolchain_linux_arm64",

        "//build/toolchains/helm3:helm3_linux_toolchain_amd64",
    )
