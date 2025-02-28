"""IDC Universe Deployer."""

load("@rules_pkg//pkg:tar.bzl", "pkg_tar")
load(":universe_config.bzl",
    "parse_universe_config_file_embed",
    "get_commit_to_components_dict_from_universe_config",
    "trim_universe_config_for_manifests",
    "trim_universe_config_for_push",
    "COMPONENT_ALL",
)

def get_git_archive(ctx, commit, git_archives):
    """Return a File that contains the output of git archive for a commit."""
    found = None
    for git_archive in git_archives:
        if git_archive.basename == "%s.tar" % commit:
            found = git_archive
            break
    if not found:
        fail("Git archive not found for commit %s" % commit)
    return found

def get_commit_metadata(ctx, commit, git_archives_metadata):
    """Return the commit metadata from the file commit_metadata.json."""
    commit_metadata_json = git_archives_metadata.get(commit, "{}")
    commit_metadata = json.decode(commit_metadata_json)
    universe_deployer_commit_metadata_json = commit_metadata.get("universe_deployer_commit_metadata", "{}")
    universe_deployer_commit_metadata = json.decode(universe_deployer_commit_metadata_json)
    return universe_deployer_commit_metadata

def build_and_push_from_commit(ctx, commit, component, universe_config_file, git_archive, legacy_deployment_artifacts):
    """Return an action that runs Builder Pusher for a commit."""
    args = ctx.actions.args()
    deployment_artifacts_tar = ctx.actions.declare_file("%s_deployment_artifacts_%s_%s.tar" % (ctx.label.name, commit, component))
    args.add("--bazel-build-opt=--jobs=" + ctx.attr.jobs_per_commit)
    # When executed with "make universe-deployer", this will result in a nested execution of Bazel (Bazel in Bazel).
    # This runs with an outputRoot from a pool of directories to maximize caching.
    # To prevent re-use and unexpected termination of the Bazel server, use batch mode so that the Bazel server is not used.
    args.add("--bazel-startup-opt=--batch")
    args.add("--commit", commit)
    if component != COMPONENT_ALL:
        args.add("--component", component)
    args.add("--git-archive", git_archive)
    args.add("--legacy-deployment-artifacts=%d" % int(legacy_deployment_artifacts))
    args.add("--max-pool-dirs", ctx.attr.max_pool_dirs)
    args.add("--pool-dir", ctx.attr.pool_dir)
    args.add("--output", deployment_artifacts_tar)

    patch_tar_file = get_patch_tar_for_commit(ctx, commit)
    if patch_tar_file != None:
        args.add("--patch-tar", patch_tar_file.path)

    args.add("--skip-push=%d" % int(ctx.attr.skip_push))
    args.add("--universe-config", universe_config_file)
    ctx.actions.run(
        executable = ctx.executable._builder_pusher,
        outputs = [deployment_artifacts_tar],
        inputs = depset([ctx.executable._builder_pusher, git_archive, universe_config_file, patch_tar_file] + ctx.files.secrets + ctx.files._bazelisk),
        use_default_shell_env = True,
        mnemonic = "BuildAndPushFromCommit",
        arguments = [args],
        progress_message = "Building and pushing {} from commit {}, component {}".format(ctx.label, commit, component),
    )
    return deployment_artifacts_tar

def get_patch_tar_for_commit(ctx, commit):
    """Return the patch tar file for the specified commit."""
    found_patch_tar_file = None
    default_patch_tar_file = None
    for patch_tar_target, commits_str in ctx.attr.patch_tar_for_commits.items():
        patch_tar_file = patch_tar_target.files.to_list()[0]
        if commits_str == "default":
            default_patch_tar_file = patch_tar_file
        else:
            commits = commits_str.split(",")
            if commit in commits:
                found_patch_tar_file = patch_tar_file
    if found_patch_tar_file == None:
        found_patch_tar_file = default_patch_tar_file
    return found_patch_tar_file

def create_release_for_commit(ctx, commit, component, deployment_artifacts_tar):
    """Return an action that runs Create Release for a commit.
    This runs go/pkg/universe_deployer/cmd/create_release/main.go."""
    args = ctx.actions.args()
    release_info = ctx.actions.declare_file("%s_release_info_%s_%s" % (ctx.label.name, commit, component))
    args.add("--artifact-repository-url-file", ctx.file.artifact_repository_url_file.path)
    args.add("--commit", commit)
    args.add("--component", component)
    args.add("--deployment-artifacts-tar", deployment_artifacts_tar)
    args.add("--output", release_info)
    ctx.actions.run(
        executable = ctx.executable._create_release,
        outputs = [release_info],
        inputs = depset([ctx.executable._create_release, deployment_artifacts_tar] + ctx.files.secrets + ctx.files.artifact_repository_url_file),
        use_default_shell_env = True,
        mnemonic = "CreateReleaseForCommit",
        arguments = [args],
        progress_message = "Creating release {} for commit {}".format(ctx.label, commit),
    )
    return release_info

def build_manifests_from_commit(ctx, commit, component, snapshot, universe_config_file, deployment_artifacts_tar):
    """Return an action that runs Manifests Generator for a commit.
    DEPRECATED: This will be replaced by go/pkg/universe_deployer/cmd/deploy_all_in_k8s/main.go."""
    args = ctx.actions.args()
    manifests_tar = ctx.actions.declare_file("%s_manifests_%s_%s.tar" % (ctx.label.name, commit, component))
    args.add(deployment_artifacts_tar)
    args.add("go/pkg/universe_deployer/cmd/manifests_generator/manifests_generator_/manifests_generator")
    args.add("--commit", commit)
    if component != COMPONENT_ALL:
        args.add("--component", component)
    args.add("--output", manifests_tar)
    args.add("--universe-config", universe_config_file)
    args.add("--snapshot=%d" % int(snapshot))
    ctx.actions.run(
        executable = ctx.executable._commit_runner,
        outputs = [manifests_tar],
        inputs = depset([universe_config_file, deployment_artifacts_tar]),
        use_default_shell_env = True,
        mnemonic = "BuildManifests",
        arguments = [args],
        progress_message = "Building manifests {} from commit {}, component {}".format(ctx.label, commit, component),
    )
    return manifests_tar

def write_json_file(ctx, obj, filename):
    obj_json = json.encode_indent(obj)
    json_file = ctx.actions.declare_file(filename)
    ctx.actions.write(
        output = json_file,
        content = obj_json,
    )
    return json_file

def _create_releases_impl(ctx):
    """Build and push containers, charts, and artifacts for all commits in the universe.    
    This is nearly identical to _manifests_tars_impl through the build_and_push_from_commit step.
    """
    universe_config = parse_universe_config_file_embed(ctx.attr.universe_config_file_embed)
    print("universe_config=%r" % universe_config)
    allow_config_commit = True
    commits = get_commit_to_components_dict_from_universe_config(universe_config, allow_config_commit)
    print("Number of unique commits: %d" % len(commits))

    releases_info = []
    num_builds_of_all_components = 0
    num_builds_of_single_component = 0

    for commit, components in commits.items():
        git_archive = get_git_archive(ctx, commit, ctx.files.git_archives)

        commit_metadata = get_commit_metadata(ctx, commit, ctx.attr.git_archives_metadata)
        print("commit_metadata=%r" % commit_metadata)
        allow_build_of_single_component = commit_metadata.get("universeDeployerAllowBuildOfSingleComponent", False)
        print("allow_build_of_single_component=%s" % allow_build_of_single_component)

        legacy_deployment_artifacts = not allow_build_of_single_component

        # This will build a single component for a new commit.
        # To change to build all components, even for a new commit, set build_single_component=False.
        build_single_component = allow_build_of_single_component

        if build_single_component:
            num_builds_of_single_component += len(components)
        else:
            # Build all components (legacy mode).
            components = {COMPONENT_ALL: True}
            num_builds_of_all_components += 1

        for component, _ in components.items():
            universe_config_for_push = trim_universe_config_for_push(universe_config, commit, component)
            print("universe_config_for_push(commit=%s, component=%s)=%s" % (commit, component, universe_config_for_push))
            universe_config_file_for_push = write_json_file(ctx,
                universe_config_for_push,
                "universe_config_for_push_%s_%s.json" % (commit, component),
            )
            deployment_artifacts_tar = build_and_push_from_commit(ctx, commit, component, universe_config_file_for_push, git_archive, legacy_deployment_artifacts)

            release_info = create_release_for_commit(ctx, commit, component, deployment_artifacts_tar)
            releases_info.append(release_info)

    print("Number of builds of a single component: %d" % num_builds_of_single_component)
    print("Number of builds of all components (legacy): %d" % num_builds_of_all_components)

    return [
        DefaultInfo(
            files = depset(releases_info),
            runfiles = ctx.runfiles(releases_info),
        ),
    ]

_create_releases = rule(
    implementation = _create_releases_impl,
    attrs = {
        "artifact_repository_url_file": attr.label(
            allow_single_file = True,
            default = Label("//build/dynamic:ARTIFACT_REPOSITORY_URL"),
            doc = "File that contains the artifact repository URL.",
        ),
        "bazel_build_opts": attr.string(),
        "bazel_startup_opts": attr.string(),
        "jobs_per_commit": attr.string(
            default = "8",
            doc = "The value of the --jobs parameter passed to Bazel to build and push each commit.",
        ),
        "git_archives": attr.label(
            allow_files = True,
            mandatory = True,
            doc = "A Bazel repository created by the git_archive Bazel repository rule.",
        ),
        "git_archives_metadata": attr.string_dict(
            doc = "",
        ),
        "max_pool_dirs": attr.int(
            default = 1,
            doc = "The number of directories within pool-dir. Each directory supports a single concurrent Bazel invocation.",
        ),
        "patch_tar_for_commits": attr.label_keyed_string_dict(
            allow_files = True,
            doc = "A mapping from a label for a tar file to a comma-separated list of commits, " +
            "onto which the tar file will be extracted, replacing any files extracted from each git archive. " +
            "If the list of commits is 'default', the patch will apply to any commit which is not otherwise listed. " +
            "Patches can be used to change Bazel caching or the pushing process. " +
            "Patches must not change the container images, Helm charts, or Argo CD manifests.",
        ),
        "pool_dir": attr.string(
            mandatory = True,
            doc = "All intermediate files including Bazel local cache files will be stored within this directory. "+
			"This directory should be persistent to allow the cached data to be reused.",
        ),
        "secrets": attr.label(
            allow_files = True,
            mandatory = True,
            doc = "Secrets for Harbor.",
        ),
        "skip_push": attr.bool(
            default = False,
            doc = "If True, do not push containers and Helm charts.",
        ),
        "universe_config_file_embed": attr.string_dict(
            mandatory = True,
            doc = "A Universe Config File embedded with bzl_file_embed.",
        ),
        "_bazelisk": attr.label(
            allow_single_file = True,
            default = Label("@bazel_binaries_bazelisk//:bazelisk"),
            doc = "Bazelisk binary",
        ),
        "_builder_pusher": attr.label(
            cfg = "exec",
            default = Label("//go/pkg/universe_deployer/cmd/builder_pusher"),
            doc = "Universe Deployer Builder Pusher binary",
            executable = True,
        ),
        "_create_release": attr.label(
            cfg = "exec",
            default = Label("//go/pkg/universe_deployer/cmd/create_release"),
            doc = "Universe Deployer Create Release binary",
            executable = True,
        ),
        "_commit_runner": attr.label(
            cfg = "exec",
            default = Label("//go/pkg/universe_deployer/cmd/commit_runner"),
            doc = "Universe Deployer Commit Runner binary",
            executable = True,
        ),
    },
    doc = "Create releases (build and push containers, charts, and artifacts) for all commits in a Universe Config file.",
)

def _manifests_tars_impl(ctx):
    """Generate Argo CD manifests.
    DEPRECATED: This will be replaced by go/pkg/universe_deployer/cmd/deploy_all_in_k8s/main.go."""
    universe_config = parse_universe_config_file_embed(ctx.attr.universe_config_file_embed)
    allow_config_commit = False
    commits = get_commit_to_components_dict_from_universe_config(universe_config, allow_config_commit)
    print("Number of unique commits: %d" % len(commits))

    manifests_tars = []
    num_builds_of_all_components = 0
    num_builds_of_single_component = 0

    for commit, components in commits.items():
        git_archive = get_git_archive(ctx, commit, ctx.files.git_archives)

        commit_metadata = get_commit_metadata(ctx, commit, ctx.attr.git_archives_metadata)
        print("commit_metadata=%r" % commit_metadata)
        allow_build_of_single_component = commit_metadata.get("universeDeployerAllowBuildOfSingleComponent", False)
        print("allow_build_of_single_component=%s" % allow_build_of_single_component)

        legacy_deployment_artifacts = not allow_build_of_single_component

        # This will build a single component for a new commit.
        # To change to build all components, even for a new commit, set build_single_component=False.
        build_single_component = allow_build_of_single_component

        if build_single_component:
            num_builds_of_single_component += len(components)
        else:
            # Build all components (legacy mode).
            components = {COMPONENT_ALL: True}
            num_builds_of_all_components += 1

        for component, _ in components.items():
            universe_config_for_push = trim_universe_config_for_push(universe_config, commit, component)
            print("universe_config_for_push(commit=%s, component=%s)=%s" % (commit, component, universe_config_for_push))
            universe_config_file_for_push = write_json_file(ctx,
                universe_config_for_push,
                "universe_config_for_push_%s_%s.json" % (commit, component),
            )
            deployment_artifacts_tar = build_and_push_from_commit(ctx, commit, component, universe_config_file_for_push, git_archive, legacy_deployment_artifacts)
            universe_config_for_manifests = trim_universe_config_for_manifests(universe_config, commit, component)
            print("universe_config_for_manifests(commit=%s, component=%s)=%s" % (commit, component, universe_config_for_manifests))
            universe_config_file_for_manifests = write_json_file(ctx,
                universe_config_for_manifests,
                "universe_config_for_manifests_%s_%s.json" % (commit, component),
            )
            snapshot = False
            manifests_tar = build_manifests_from_commit(ctx, commit, component, snapshot, universe_config_file_for_manifests, deployment_artifacts_tar)
            manifests_tars.append(manifests_tar)

    print("Number of builds of a single component: %d" % num_builds_of_single_component)
    print("Number of builds of all components (legacy): %d" % num_builds_of_all_components)

    return [
        DefaultInfo(
            files = depset(manifests_tars),
            runfiles = ctx.runfiles(manifests_tars),
        ),
    ]

_manifests_tars = rule(
    implementation = _manifests_tars_impl,
    attrs = {
        "bazel_build_opts": attr.string(),
        "bazel_startup_opts": attr.string(),
        "jobs_per_commit": attr.string(
            default = "8",
            doc = "The value of the --jobs parameter passed to Bazel to build and push each commit.",
        ),
        "git_archives": attr.label(
            allow_files = True,
            mandatory = True,
            doc = "A Bazel repository created by the git_archive Bazel repository rule.",
        ),
        "git_archives_metadata": attr.string_dict(
            doc = "",
        ),
        "max_pool_dirs": attr.int(
            default = 1,
            doc = "The number of directories within pool-dir. Each directory supports a single concurrent Bazel invocation.",
        ),
        "patch_tar_for_commits": attr.label_keyed_string_dict(
            allow_files = True,
            doc = "A mapping from a label for a tar file to a comma-separated list of commits, " +
            "onto which the tar file will be extracted, replacing any files extracted from each git archive. " +
            "If the list of commits is 'default', the patch will apply to any commit which is not otherwise listed. " +
            "Patches can be used to change Bazel caching or the pushing process. " +
            "Patches must not change the container images, Helm charts, or Argo CD manifests.",
        ),
        "pool_dir": attr.string(
            mandatory = True,
            doc = "All intermediate files including Bazel local cache files will be stored within this directory. "+
			"This directory should be persistent to allow the cached data to be reused.",
        ),
        "secrets": attr.label(
            allow_files = True,
            mandatory = True,
            doc = "Secrets for Harbor.",
        ),
        "skip_push": attr.bool(
            default = False,
            doc = "If True, do not push containers and Helm charts.",
        ),
        "universe_config_file_embed": attr.string_dict(
            mandatory = True,
            doc = "A Universe Config File embedded with bzl_file_embed.",
        ),
        "_bazelisk": attr.label(
            allow_single_file = True,
            default = Label("@bazel_binaries_bazelisk//:bazelisk"),
            doc = "Bazelisk binary",
        ),
        "_builder_pusher": attr.label(
            cfg = "exec",
            default = Label("//go/pkg/universe_deployer/cmd/builder_pusher"),
            doc = "Universe Deployer Builder Pusher binary",
            executable = True,
        ),
        "_commit_runner": attr.label(
            cfg = "exec",
            default = Label("//go/pkg/universe_deployer/cmd/commit_runner"),
            doc = "Universe Deployer Commit Runner binary",
            executable = True,
        ),
    },
    doc = "Generate Argo CD manifests from multiple commits in a Universe Config File. " +
    "DEPRECATED: This will be replaced by go/pkg/universe_deployer/cmd/deploy_all_in_k8s/main.go.",
)

def _git_pusher_impl(ctx):
    runfiles = ctx.runfiles(
        files = [
            ctx.executable._git_pusher,
            ctx.file._yq,
        ] + ctx.files.manifests_tar,
    )
    executable = ctx.actions.declare_file("%s_git_pusher" % ctx.label.name)
    script = (
        '%s' % ctx.executable._git_pusher.short_path +
        ' --authoritative-git-branch "%s"' % ctx.attr.authoritative_git_branch +
        ' --authoritative-git-remote "%s"' % ctx.attr.authoritative_git_remote +
        ' --manifests-git-branch "%s"' % ctx.attr.manifests_git_branch +
        ' --manifests-git-remote "%s"' % ctx.attr.manifests_git_remote +
        ' --manifests-tar "%s"' % ctx.file.manifests_tar.short_path +
        ' --push-to-new-branch=%d' % int(ctx.attr.push_to_new_branch) +
        ' --yq-binary "%s"' % ctx.file._yq.short_path +
        ' $@'
    )
    ctx.actions.write(executable, script, is_executable = True)
    return [
        DefaultInfo(
            executable = executable,
            runfiles = runfiles,
            ),
        ]

_git_pusher = rule(
    implementation = _git_pusher_impl,
    attrs = {
        "authoritative_git_branch": attr.string(
            default = "main",
            doc = "The monorepo Git branch that is authoritative manifests_tar.",
        ),
        "authoritative_git_remote": attr.string(
            mandatory = True,
            doc = "The monorepo Git remote that is authoritative manifests_tar.",
        ),
        "manifests_git_branch": attr.string(
            default = "main",
            doc = "The Argo CD manifests Git branch that this will clone.",
        ),
        "manifests_git_remote": attr.string(
            mandatory = True,
            doc = "The Argo CD manifests Git remote that this will clone and create a new branch in.",
        ),
        "manifests_tar": attr.label(
            allow_single_file = True,
            mandatory = True,
            doc = "The output of the _manifests_tars rule.",
        ),
        "push_to_new_branch": attr.bool(
            default = True,
            doc = "If True, a new branch will be created in the remote. If False, manifests_git_branch will be updated.",
        ),
        "_git_pusher": attr.label(
            cfg = "exec",
            default = Label("//go/pkg/universe_deployer/cmd/git_pusher"),
            doc = "Universe Deployer Git Pusher binary",
            executable = True,
        ),
        "_yq": attr.label(
            allow_single_file = True,
            default = Label("@yq_linux_amd64_file//file"),
            doc = "yq binary",
        ),
    },
    doc = "Push Argo CD manifests to a Git repository. " + 
    "DEPRECATED: This will be replaced by go/pkg/universe_deployer/cmd/deploy_all_in_k8s/main.go.",
    executable = True,
)

def universe_deployer(name, authoritative_git_branch, authoritative_git_remote, manifests_git_branch, manifests_git_remote, push_to_new_branch, secrets, **kwargs):
    """DEPRECATED: This will be replaced by go/pkg/universe_deployer/cmd/deploy_all_in_k8s/main.go."""
    _manifests_tars(
        name = "%s_tars" % name,
        secrets = secrets,
        **kwargs,
    )
    pkg_tar(
        name = name,
        deps = ["%s_tars" % name],
    )
    _git_pusher(
        name = "%s_git_pusher" % name,
        authoritative_git_branch = authoritative_git_branch,
        authoritative_git_remote = authoritative_git_remote,
        manifests_git_branch = manifests_git_branch,
        manifests_git_remote = manifests_git_remote,
        manifests_tar = name,
        push_to_new_branch = push_to_new_branch,
    )

def create_releases(name, **kwargs):
    _create_releases(
        name = name,
        **kwargs,
    )
