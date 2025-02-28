# Based on https://github.com/masmovil/bazel-rules/blob/master/helm/helm-push.bzl
# and https://github.com/masmovil/bazel-rules/blob/master/helm/helm-release.bzl

load("//build/helpers:helpers.bzl", "write_sh")

def _helm_push_impl(ctx):
    """Push a helm chart to an OCI helm repository
    """

    helm3_binary = ctx.toolchains["//build/toolchains/helm3:toolchain_type"].helminfo.tool.files.to_list()
    helm3_path = helm3_binary[0].short_path
    chart = ctx.file.chart
    user = ctx.expand_make_variables("repository_username", ctx.attr.repository_username, {})
    user_pass = ctx.expand_make_variables("repository_password", ctx.attr.repository_password, {})
    repo_url = ctx.expand_make_variables("repository_url", ctx.attr.repository_url, {})
    repository_name = ctx.expand_make_variables("repository_name", ctx.attr.repository_name, {})

    # Generates the exec bash file with the provided substitutions
    exec_file = write_sh(
      ctx,
      "helm_bash",
      """
        #!/usr/bin/env bash
        set -e

        run_with_retry() {
          local attempts=$1
          local sleep=$2
          shift 2
          for i in $(seq 1 ${attempts}); do
              [ $i -gt 1 ] && echo Failed. Will retry in ${sleep} seconds... && sleep ${sleep}
              "$@" && s=0 && break || s=$?
          done
          return $s
        }

        if [ "{USERNAME}" != "" ] && [ "{PASSWORD}" != "" ]; then
          echo "{PASSWORD}" | {HELM3_PATH} registry login -u {USERNAME} --password-stdin {REPOSITORY_URL}
        fi

        run_with_retry 6 10 {HELM3_PATH} push {CHART_PATH} oci://{REPOSITORY_URL}/{REPOSITORY_NAME}
      """,
      {
        "{CHART_PATH}": chart.short_path,
        "{USERNAME}": user,
        "{PASSWORD}": user_pass,
        "{REPOSITORY_URL}": repo_url,
        "{REPOSITORY_NAME}": repository_name,
        "{HELM3_PATH}": helm3_path,
      }
    )

    runfiles = ctx.runfiles(
      files = [
        chart,
      ] + helm3_binary
    )

    return [DefaultInfo(
      executable = exec_file,
      runfiles = runfiles,
    )]

helm_push = rule(
    implementation = _helm_push_impl,
    attrs = {
      "chart": attr.label(allow_single_file = True, mandatory = True),
      "repository_name": attr.string(mandatory = True),
      "repository_url": attr.string(mandatory = True),
      "repository_username": attr.string(mandatory = False),
      "repository_password": attr.string(mandatory = False),
    },
    doc = "Push helm chart to an OCI helm repository",
    executable = True,
    toolchains = ["//build/toolchains/helm3:toolchain_type"],
)
