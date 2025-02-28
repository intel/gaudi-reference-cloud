# Based on https://github.com/masmovil/bazel-rules/blob/master/helm/helm-chart-package.bzl
# Modified as follows:
#  - Append a hash of the Helm chart to the Helm chart version
#  - Remove execution requirements that prevents caching.
#  - Remove uneeded functionality to include stamp files.
#  - Fix bug that adds previously deleted files in the chart.

# Load docker image providers
load(
    "@io_bazel_rules_docker//container:providers.bzl",
    "ImageInfo",
    "LayerInfo",
)
load("@com_github_masmovil_bazel_rules//helpers:helpers.bzl", "get_make_value_or_default", "write_sh")

ChartInfo = provider(fields = [
    "chart",
])

def _helm_chart_impl(ctx):
    """Defines a helm chart (directory containing a Chart.yaml).
    Args:
        name: A unique name for this rule.
        srcs: Source files to include as the helm chart. Typically this will just be glob(["**"]).
        update_deps: Whether or not to run a helm dependency update prior to packaging.
    """
    chart_root_path = ""
    tmp_chart_root = ""
    tmp_chart_manifest_path = ""
    tmp_working_dir = "_tmp"
    inputs = [] + ctx.files.srcs

    digest_path = ""
    image_tag = ""
    yq = ctx.toolchains["@com_github_masmovil_bazel_rules//toolchains/yq:toolchain_type"].yqinfo.tool.files.to_list()[0]
    helm_toolchain = ctx.toolchains["@com_github_masmovil_bazel_rules//toolchains/helm-3:toolchain_type"].helminfo
    helm = helm_toolchain.tool.files.to_list()[0]
    helm_cache_path = helm_toolchain.helm_xdg_cache_home
    helm_config_path = helm_toolchain.helm_xdg_config_home
    helm_data_path = helm_toolchain.helm_xdg_data_home

    # declare rule output
    targz = ctx.outputs.targz
    helm_chart_version_file = ctx.outputs.helm_chart_version_file

    inputs += [helm, yq]
    inputs.append(ctx.file.helm_chart_version)
    helm_package_dir_files = []

    # locate chart root path trying to find Chart.yaml file
    for i, srcfile in enumerate(ctx.files.srcs):
        if srcfile.path.endswith("Chart.yaml"):
            chart_root_path = srcfile.dirname
            break

    # move chart files to temporal directory in order to manipulate necessary files
    for i, srcfile in enumerate(ctx.files.srcs):
        if srcfile.path.startswith(chart_root_path):
            out = ctx.actions.declare_file(tmp_working_dir + "/" + srcfile.path)
            inputs.append(out)
            helm_package_dir_files.append(out)

            # extract location of the chart in the new directory
            if srcfile.path.endswith("Chart.yaml"):
                tmp_chart_root = out.dirname
                tmp_chart_manifest_path = out.path

            ctx.actions.run_shell(
                outputs = [out],
                inputs = [srcfile],
                arguments = [srcfile.path, out.path],
                command = "cp $1 $2",
            )

    if tmp_chart_root == "":
        fail("Chart.yaml not found")

    # extract docker image info from dependency rule
    if ctx.attr.image:
        digest_file = ctx.attr.image[ImageInfo].container_parts["digest"]
        digest_path = digest_file.path
        inputs = inputs + [ctx.file.image, digest_file]
    else:
        # extract docker image info from make variable or from rule attribute
        image_tag = get_make_value_or_default(ctx, ctx.attr.image_tag)

    deps = ctx.attr.chart_deps or []

    # copy generated charts by other rules into temporal chart_root/charts directory (treated as a helm dependency)
    for i, dep in enumerate(deps):
        dep_files = dep[DefaultInfo].files.to_list()
        tgz = dep[DefaultInfo].files.to_list()[0]
        out = ctx.actions.declare_file(tmp_working_dir + "/" + chart_root_path + "/charts/" + tgz.basename)
        out_dir = ctx.actions.declare_directory(tmp_working_dir + "/" + chart_root_path + "/charts/" + tgz.basename[:-len(tgz.extension) - 1])
        inputs = inputs + dep_files + [out]
        helm_package_dir_files.append(out)
        ctx.actions.run_shell(
            outputs = [out, out_dir],
            inputs = dep[DefaultInfo].files,
            arguments = [dep[DefaultInfo].files.to_list()[0].path, out.path, out_dir.path],
            command = "rm -rf $3; cp -f $1 $2; tar -C $(dirname $2) -xzf $2",
        )

    additional_templates = ctx.attr.additional_templates or []

    # Copy additional templates to the "templates/" folder of the chart being assembled.
    # This is useful for centralizing common templates and pass them to multiple charts.
    for template in additional_templates:
        for file in template.files.to_list():
            out = ctx.actions.declare_file(tmp_working_dir + "/" + chart_root_path + "/templates/" + file.basename)
            inputs.append(out)
            helm_package_dir_files.append(out)

            ctx.actions.run_shell(
                outputs = [out],
                inputs = [file],
                arguments = [file.path, out.path],
                command = "cp $1 $2",
            )

    exec_file = ctx.actions.declare_file(ctx.label.name + "_helm_bash")

    # Convert helm_package_dir_files to be relative to chart root.
    helm_package_dir_files = [f.path[len(tmp_chart_root)+1:] for f in helm_package_dir_files]

    # Generates the exec bash file with the provided substitutions
    ctx.actions.expand_template(
        template = ctx.file._script_template,
        output = exec_file,
        is_executable = True,
        substitutions = {
            "{CHART_PATH}": tmp_chart_root,
            "{HELM_PACKAGE_DIR_FILES}": " ".join(helm_package_dir_files),
            "{CHART_MANIFEST_PATH}": tmp_chart_manifest_path,
            "{DIGEST_PATH}": digest_path,
            "{IMAGE_TAG}": image_tag,
            "{YQ_PATH}": yq.path,
            "{PACKAGE_OUTPUT_PATH}": targz.path,
            "{IMAGE_REPOSITORY}": ctx.attr.image_repository,
            "{HELM_CHART_VERSION}": ctx.file.helm_chart_version.path,
            "{HELM_CHART_VERSION_FILE}": helm_chart_version_file.path,
            "{APPEND_CHART_HASH_TO_HELM_CHART_VERSION}": str(ctx.attr.append_chart_hash_to_helm_chart_version),
            "{APP_VERSION}": ctx.file.helm_chart_version.path,
            "{HELM_CHART_NAME}": ctx.attr.package_name,
            "{HELM_PATH}": helm.path,
            "{HELM_CACHE_PATH}": helm_cache_path,
            "{HELM_CONFIG_PATH}": helm_config_path,
            "{HELM_DATA_PATH}": helm_data_path,
            "{VALUES_REPO_YAML_PATH}": ctx.attr.values_repo_yaml_path,
            "{VALUES_TAG_YAML_PATH}": ctx.attr.values_tag_yaml_path,
        },
    )

    ctx.actions.run(
        inputs = inputs,
        outputs = [targz, helm_chart_version_file],
        arguments = [],
        executable = exec_file,
    )

    return [
        DefaultInfo(
            files = depset([targz]),
        ),
    ]

_helm_chart = rule(
    implementation = _helm_chart_impl,
    attrs = {
        "srcs": attr.label_list(allow_files = True, mandatory = True),
        "image": attr.label(allow_single_file = True, mandatory = False),
        "image_tag": attr.string(mandatory = False),
        "targz": attr.output(
            doc="The packaged Helm chart file"
        ),
        "package_name": attr.string(mandatory = True),
        "helm_chart_version": attr.label(
            allow_single_file = True,
            mandatory = False,
            default = "//build/dynamic:HELM_CHART_VERSION",
            doc="The label of the file with the Helm chart version. This is the prefix if append_chart_hash_to_helm_chart_version is True."),
        "helm_chart_version_file": attr.output(
            doc="This file will be written with the generated Helm chart version, including the hash if append_chart_hash_to_helm_chart_version is True.",
        ),
        "append_chart_hash_to_helm_chart_version": attr.bool(
            default = True,
            doc = "If True, append a hash of the Helm chart contents to helm_chart_version. " +
            "Otherwise, the provided helm_chart_version will be used unchanged.",
        ),
        "image_repository": attr.string(),
        "values_repo_yaml_path": attr.string(default = "image.repository"),
        "values_tag_yaml_path": attr.string(default = "image.tag"),
        "_script_template": attr.label(allow_single_file = True, default = ":helm-chart-package.sh.tpl"),
        "chart_deps": attr.label_list(allow_files = True, mandatory = False),
        "additional_templates": attr.label_list(allow_files = True, mandatory = False),
    },
    toolchains = [
        "@com_github_masmovil_bazel_rules//toolchains/yq:toolchain_type",
        "@com_github_masmovil_bazel_rules//toolchains/helm-3:toolchain_type",
    ],
    doc = "Runs helm package, updating the image tag and version",
)

def helm_chart(**kwargs):
    name = kwargs["name"]
    package_name = kwargs["package_name"]
    _helm_chart(
        targz = "%s.tgz" % package_name,
        helm_chart_version_file = "%s.version" % name,
        **kwargs,
    )
