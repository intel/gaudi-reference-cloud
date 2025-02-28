def dict_union(d1, d2):
    result = {}
    result.update(d1)
    result.update(d2)
    return result

# To combine the KUBE_BUILDER_ASSETS setting with other settings, write:
#     env =  kubebuilder_test_env(path_to_root = "../../../..", env = {"A": "B"}),
#
# This is ugly, there's gotta be a better way
def kubebuilder_test_env(path_to_root, env = {}):
    root = select({
        "//build/kubebuilder:is_arm64_linux": dict_union(env, {
            "KUBEBUILDER_ASSETS": path_to_root + "/external/kubebuilder_tools_arm64/bin"
        }),
        "//build/kubebuilder:is_arm64_darwin": dict_union(env, {
            "KUBEBUILDER_ASSETS": path_to_root + "/external/kubebuilder_tools_arm64_darwin/bin"
        }),
        "//conditions:default": dict_union(env, {
            "KUBEBUILDER_ASSETS": path_to_root + "/external/kubebuilder_tools/bin"
        }),
    })
    return root
