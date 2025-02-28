load("//build/multirun:multirun.bzl", "simple_multirun")

def idc_push_group(name, containers=[], charts=[], jobs=0):
    """This defines targets that can be run to push all containers and/or Helm charts concurrently."""
    container_push_srcs = [":%s_container_push" % k for k, v in containers.items()]
    chart_push_srcs = [":%s_chart_push" % k for k, v in charts.items()]
    simple_multirun(
        name = "%s_container_push" % name,
        srcs = container_push_srcs,
        jobs = jobs,
    )
    simple_multirun(
        name = "%s_chart_push" % name,
        srcs = chart_push_srcs,
        jobs = jobs,
    )
    simple_multirun(
        name = "%s_container_and_chart_push" % name,
        srcs = container_push_srcs + chart_push_srcs,
        jobs = jobs,
    )
