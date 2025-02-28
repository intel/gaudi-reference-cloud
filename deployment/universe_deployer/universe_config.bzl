"""IDC Universe Deployer."""

COMPONENT_ALL = "all"

def get_commit_to_components_dict_from_universe_config(universe_config, allow_config_commit, exclude=["HEAD"]):
    """Transforms universe config and returns dict[commit, dict[component, True]].
    When called by the legacy _manifests_tars_impl for prod and staging, allow_config_commit is set False and
    a non-empty configCommit will cause a failure because this is unsupported by _manifests_tars_impl.
    When called by _create_releases_impl, allow_config_commit is set to True.
    Exclude any commits in exclude.
    """
    commits = {}

    def update_from(components):
        for component, universe_component in components.items():
            if "commit" in universe_component:
                commit = universe_component["commit"]
                if commit and not commit in exclude:
                    if not commit in commits:
                        commits[commit] = {}
                    commits[commit][component] = True
            if "configCommit" in universe_component:
                commit = universe_component["configCommit"]
                if commit and not allow_config_commit:
                    fail("configCommit must be empty")

    if "environments" in universe_config:
        for _, deployer_environment in universe_config["environments"].items():
            if "components" in deployer_environment:
                update_from(deployer_environment["components"])
            if "regions" in deployer_environment:
                for _, universe_region in deployer_environment["regions"].items():
                    if "components" in universe_region:
                        update_from(universe_region["components"])
                    if "availabilityZones" in universe_region:
                        for _, universe_availability_zone in universe_region["availabilityZones"].items():
                            if "components" in universe_availability_zone:
                                update_from(universe_availability_zone["components"])
    return commits

def get_commits_from_universe_config(universe_config):
    """Return the distinct set of commits (including configCommits) referenced in a Universe Config."""
    allow_config_commit = True
    return sorted(get_commit_to_components_dict_from_universe_config(universe_config, allow_config_commit).keys())

def get_commits_from_universe_config_json(universe_config_json):
    """Return the distinct set of commits (including configCommits) referenced in a Universe Config JSON file."""
    universe_config = json.decode(universe_config_json)
    return get_commits_from_universe_config(universe_config)

def parse_universe_config_file_embed(universe_config_file_embed):
    """Parse and merge universe config files embedded by bzl_file_embed."""
    environments = {}
    for src_label, file_content in universe_config_file_embed.items():
        parsed = json.decode(file_content)
        if "environments" in parsed:
            for environment, deployer_environment in parsed["environments"].items():
                if environment in environments:
                    fail("Environment %s is defined in multiple files: %s" % (environment, src_label))
                environments[environment] = deployer_environment
    return {"environments": environments}

def _trim_components(components, commit, component):
    """Remove components that don't have the provided commit.
    Additionally, if component is not COMPONENT_ALL, only this component will be included."""
    trimmed = {}

    def add_component(c, universe_component):
        if "commit" in universe_component:
            if universe_component["commit"] == commit:
                trimmed[c] = universe_component

    if components:
        if component == COMPONENT_ALL:
            for c, universe_component in components.items():
                add_component(c, universe_component)
        else:
            if component in components:
                add_component(component, components[component])

    return trimmed

def trim_universe_config_for_manifests(universe_config, commit, component):
    """Trim a Universe Config to the minimum required to build manifests.

    Trimming is performed to maximize Bazel cache hits.
    Irrelevant changes to the Universe Config will be trimmed.
    Only components with the provided commit will be kept.
    Additionally, if component is not COMPONENT_ALL, only this component will be included.
    Containing objects with no commits will be removed from the returned value.
    Based on go/pkg/universe_deployer/universe_config/universe_config.go."""

    trimmed = {"environments": {}}

    if "environments" in universe_config:
        for idc_env, universe_environment in universe_config["environments"].items():
            trimmed_environment = {
                "components": _trim_components(universe_environment.get("components"), commit, component),
                "regions": {},
            }
            if "regions" in universe_environment:
                for region, universe_region in universe_environment["regions"].items():
                    trimmed_region = {
                        "components": _trim_components(universe_region.get("components"), commit, component),
                        "availabilityZones": {},
                    }
                    if "availabilityZones" in universe_region:
                        for availability_zone, universe_availability_zone in universe_region["availabilityZones"].items():
                            trimmed_availability_zone = {
                                "components": _trim_components(universe_availability_zone.get("components"), commit, component),
                            }
                            if len(trimmed_availability_zone["components"]) > 0:
                                trimmed_region["availabilityZones"][availability_zone] = trimmed_availability_zone
                    if len(trimmed_region["components"]) > 0 or len(trimmed_region["availabilityZones"]) > 0:
                        trimmed_environment["regions"][region] = trimmed_region
            if len(trimmed_environment["components"]) > 0 or len(trimmed_environment["regions"]) > 0:
                trimmed["environments"][idc_env] = trimmed_environment

    return trimmed

def trim_universe_config_for_push(universe_config, commit, component):
    """Trim a Universe Config to the minimum required to push containers and charts to the registry.

    Trimming is performed to maximize Bazel cache hits.
    Irrelevant changes to the Universe Config will be trimmed.
    The result will depend only on the set of environments that contain the provided commit, and the component
    Regions will be empty."""

    trimmed_for_commit = trim_universe_config_for_manifests(universe_config, commit, component)
    trimmed = {"environments": {}}

    if "environments" in trimmed_for_commit:
        for idc_env, _ in trimmed_for_commit["environments"].items():
            # If this environment gets pushed to the same Harbor as another environment, then combine to just a single environment
            # to avoid pushing to the same Harbor multiple times.
            # This also improves Bazel caching.
            hint_push_to_environment = universe_config["environments"][idc_env].get("hint_push_to_environment", "")
            if hint_push_to_environment != "":
                idc_env = hint_push_to_environment
            trimmed_environment = {
                "components": {
                    component: {
                        "commit": commit,
                    },
                },
            }
            trimmed["environments"][idc_env] = trimmed_environment

    return trimmed
