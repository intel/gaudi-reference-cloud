load("@aspect_bazel_lib//lib:copy_file.bzl", "copy_file")
load(":artifacts.bzl", "ARTIFACTS")

def normalize_artifact(k, v):
    """Transform records in ARTIFACTS to a normalized structure."""
    new_value = {}

    charts = {}
    if "chart" in v:
        charts[k] = v["chart"]
    new_value["charts"] = charts

    new_value["components"] = v.get("components", [])

    images = v.get("images", {})
    if "image" in v:
        images[k] = v["image"]
    new_value["images"] = images

    return new_value

NORMALIZED_ARTIFACTS = {k: normalize_artifact(k, v) for k, v in ARTIFACTS.items()}

# dict[image name, image target])
CONTAINERS = {
    image_name: image_target
        for k, v in NORMALIZED_ARTIFACTS.items()
        for image_name, image_target in v["images"].items()
}

# dict[chart name, chart target]
CHARTS = {
    chart_name: chart_target
        for k, v in NORMALIZED_ARTIFACTS.items()
        for chart_name, chart_target in v["charts"].items()
}

# dict[chart name, chart target]
CHARTS_EXCEPT_IDC_VERSIONS = {k: v for k, v in CHARTS.items() if k != "idc-versions"}

# The distinct list of components in NORMALIZED_ARTIFACTS.
# list[component name]
COMPONENTS = {c: True for k, v in NORMALIZED_ARTIFACTS.items() if "components" in v for c in v["components"]}.keys()

# dict[component name, dict(charts=dict[chart name, chart target], images=dict[image name, image target])]
COMPONENT_TO_CHARTS_AND_CONTAINERS_DICT = {c: {
    "charts": {
        chart_name: chart_target
            for k, v in NORMALIZED_ARTIFACTS.items() if c in v["components"]
            for chart_name, chart_target in v["charts"].items()
    },
    "images": {
        image_name: image_target
            for k, v in NORMALIZED_ARTIFACTS.items() if c in v["components"]
            for image_name, image_target in v["images"].items()
    },
} for c in COMPONENTS}

# dict[component name, list[chart name]]
COMPONENT_TO_CHARTS_DICT = {k: [
    chart_name for chart_name, chart_target in v["charts"].items()
] for k, v in COMPONENT_TO_CHARTS_AND_CONTAINERS_DICT.items()}

def idc_container(name, target):
    copy_file(
        name = "%s_container" % name,
        src = "%s.tar" % target,
        out = "containers/%s.tar" % name,
        allow_symlink = True,
    )

def idc_chart(name, target):
    copy_file(
        name = "%s_chart" % name,
        src = target,
        out = "charts/%s.tgz" % name,
        allow_symlink = True,
    )
    copy_file(
        name = "%s_chart_version" % name,
        src = "%s.version" % target,
        out = "chart_versions/%s.version" % name,
        allow_symlink = True,
        visibility = ["//visibility:public"],
    )
