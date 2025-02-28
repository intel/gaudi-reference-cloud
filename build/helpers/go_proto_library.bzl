load("@aspect_bazel_lib//lib:copy_to_directory.bzl", "copy_to_directory")
load("@aspect_bazel_lib//lib:directory_path.bzl", "make_directory_path")
load("@aspect_bazel_lib//lib:write_source_files.bzl", "write_source_files")

# buildifier: disable=function-docstring-args
def write_go_generated_source_files(name, target, output_files, visibility):
    files_target = "_{}.filegroup".format(name)
    dir_target = "_{}.directory".format(name)

    native.filegroup(
        name = files_target,
        srcs = [target],
        output_group = "go_generated_srcs",
    )

    copy_to_directory(
        name = dir_target,
        srcs = [files_target],
        root_paths = ["**"],
    )

    write_source_files(
        name = name,
        visibility = visibility,
        files = {
            output_file: make_directory_path("_{}_dirpath".format(output_file), dir_target, output_file)
            for output_file in output_files
        },
    )