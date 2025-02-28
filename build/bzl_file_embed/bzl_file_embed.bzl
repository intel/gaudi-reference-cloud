def _bzl_file_embed(repository_ctx):
    build_contents = 'package(default_visibility = ["//visibility:public"])\n'
    repository_ctx.file("BUILD.bazel", build_contents)

    content_map = {}

    for src in repository_ctx.attr.srcs:
        key = str(src)
        content = repository_ctx.read(repository_ctx.path(src))
        content_map[key] = content

    defs_bzl_content = "%s = %r\n" % (repository_ctx.name, content_map)
    repository_ctx.file("defs.bzl", defs_bzl_content)

bzl_file_embed = repository_rule(
    implementation = _bzl_file_embed,
    attrs = {
        "srcs": attr.label_list(
            allow_files = True,
            mandatory = True,
        ),
    },
    doc = ("Bzl File Embed is a Bazel repository rule that allows the contents of an arbitrary file"
        + " in the Bazel workspace to be accessible to Bazel rules during the Bazel analysis stage."),
)
