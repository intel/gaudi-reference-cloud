# https://github.com/ash2k/bazel-tools/tree/master/multirun
load("@com_github_ash2k_bazel_tools//multirun:def.bzl", "multirun", "command")

def simple_multirun(name, srcs, jobs=1):
    """Run multiple targets in a single bazel run command."""
    [command(
        name = "%s_command_%d" % (name, i),
        command = srcs[i],
    ) for i in range(len(srcs))]
    multirun(
        name = name,
        commands = ["%s_command_%d" % (name, i) for i in range(len(srcs))],
        jobs = jobs,
    )
