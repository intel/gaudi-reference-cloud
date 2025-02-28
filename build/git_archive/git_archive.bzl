
load("//deployment/universe_deployer:universe_config.bzl", "get_commits_from_universe_config_json")

# Increment this when changing git_archive.bzl in a way that affects the cached data.
cache_info_version = 3

def _git_archive_impl(repository_ctx):
    commits = get_commits(repository_ctx)
    metadata_map = {}

    if len(commits) > 0:
        temp_dir = "tmp_%s" % uuidgen(repository_ctx)
        cache_dir = init_cache(repository_ctx, temp_dir)
        git_dir = init_git_repo(repository_ctx, temp_dir)

        for commit in commits:
            tar_file_cache, commit_metadata_file_cache = update_cache(repository_ctx, commit, temp_dir, cache_dir, git_dir)
            load_tar_from_cache(repository_ctx, commit, tar_file_cache)
            commit_metadata = load_metadata_from_cache(repository_ctx, commit, commit_metadata_file_cache)
            metadata_map[commit] = commit_metadata

        success = repository_ctx.delete(temp_dir)
        if not success:
            fail("error deleting directory")

    write_defs(repository_ctx, metadata_map)
    write_build(repository_ctx, commits)

def get_commits(repository_ctx):
    """Build commit list. It will include the static commits attribute plus the commits in the srcs files."""
    commits = list(repository_ctx.attr.commits)
    for src in repository_ctx.attr.srcs:
        src_content = repository_ctx.read(repository_ctx.path(src))
        src_commits = get_commits_from_universe_config_json(src_content)
        commits.extend(src_commits)
    commits = depset(commits).to_list()
    return commits

def init_cache(repository_ctx, temp_dir):
    """Create a cache directory to store tar files and metadata. This will be in the home directory."""
    cache_key = get_cache_key(repository_ctx, temp_dir)
    cache_dir = "%s/.cache/idc_universe_deployer_git_archive/%s" % (repository_ctx.os.environ["HOME"], cache_key)
    exec_result = repository_ctx.execute(["mkdir", "-p", cache_dir])
    if exec_result.return_code != 0:
        fail("error creating directory %s" % cache_dir)
    return cache_dir

def get_cache_key(repository_ctx, temp_dir):
    """Return a cache key that should change only when the cached data should be invalidated."""
    # cache_info should contain all attribute that affect the cached data.
    cache_info = [
        cache_info_version,
        repository_ctx.attr.commit_metadata_source_files,
    ]
    cache_info_file = "%s/cache_info_%s.json" % (temp_dir, uuidgen(repository_ctx))
    write_json_file(repository_ctx, cache_info, cache_info_file)
    exec_result = repository_ctx.execute(["sha256sum", cache_info_file])
    if exec_result.return_code != 0:
        fail("error running sha256sum")
    output = exec_result.stdout
    tokens = output.split(" ")
    cache_key = tokens[0]
    return cache_key

def init_git_repo(repository_ctx, temp_dir):
    """Initialize a bare git repo with a single remote. This does not fetch any commits"""
    remote = repository_ctx.attr.remote
    git_dir = "%s/git" % temp_dir

    exec_result = repository_ctx.execute(["mkdir", "-p", git_dir])
    if exec_result.return_code != 0:
        fail("error creating directory %s" % git_dir)

    exec_result = repository_ctx.execute(
        ["git", "init", "--quiet", "--bare"],
        working_directory = git_dir,
        quiet = False,
    )
    if exec_result.return_code != 0:
        fail("error running git init")

    exec_result = repository_ctx.execute(
        ["git", "remote", "add", "origin", remote],
        working_directory = git_dir,
        quiet = False,
    )
    if exec_result.return_code != 0:
        fail("error running git remote add")

    return git_dir

def update_cache(repository_ctx, commit, temp_dir, cache_dir, git_dir):
    """If not cached, create the git archive tar and store commit metadata in the cache."""
    tar_file_cache = "%s/%s.tar" % (cache_dir, commit)
    commit_metadata_file_cache = "%s/%s_metadata.json" % (cache_dir, commit)

    # Check if files already exists in cache.
    cached = repository_ctx.path(tar_file_cache).exists and repository_ctx.path(commit_metadata_file_cache).exists
    if not cached:
        print("commit %s: Not found in cache" % commit)
        git_fetch(repository_ctx, commit, git_dir)
        update_cache_tar(repository_ctx, commit, git_dir, tar_file_cache)
        update_cache_metadata(repository_ctx, commit, temp_dir, git_dir, commit_metadata_file_cache)

    return tar_file_cache, commit_metadata_file_cache

def git_fetch(repository_ctx, commit, git_dir):
    exec_result = repository_ctx.execute(
        ["git", "fetch", "origin", commit],
        working_directory = git_dir,
        quiet = False,
    )
    if exec_result.return_code != 0:
        fail("error running git fetch")

def update_cache_tar(repository_ctx, commit, git_dir, tar_file_cache):
    tar_file_temp = "%s.tmp.%s" % (tar_file_cache, uuidgen(repository_ctx))

    exec_result = repository_ctx.execute(
        ["git", "archive", "--format=tar", "--output", tar_file_temp, commit],
        working_directory = git_dir,
        quiet = False,
    )
    if exec_result.return_code != 0:
        fail("error running git archive")

    exec_result = repository_ctx.execute(
        ["mv", tar_file_temp, tar_file_cache],
    )
    if exec_result.return_code != 0:
        fail("Unable to move %s to %s" % (tar_file_temp, tar_file_cache))

def update_cache_metadata(repository_ctx, commit, temp_dir, git_dir, commit_metadata_file_cache):
    # Read commit metadata source files from git repo.
    commit_metadata = {}
    for key, filename in repository_ctx.attr.commit_metadata_source_files.items():
        exec_result = repository_ctx.execute(
            ["git", "show", "%s:%s" % (commit, filename)],
            working_directory = git_dir,
            quiet = True,
        )
        if exec_result.return_code == 0:
            commit_metadata[key] = exec_result.stdout
        else:
            print("commit %s: Unable to read metadata file %s" % (commit, filename))
            # Ignore the error and continue.

    # Write commit metadata as a single json file to the cache.
    commit_metadata_file_temp = "%s/%s_metadata.json" % (temp_dir, commit)
    write_json_file(repository_ctx, commit_metadata, commit_metadata_file_temp)
    exec_result = repository_ctx.execute(
        ["mv", commit_metadata_file_temp, commit_metadata_file_cache],
    )
    if exec_result.return_code != 0:
        fail("Unable to move %s to %s" % (commit_metadata_file_temp, commit_metadata_file_cache))

def load_tar_from_cache(repository_ctx, commit, tar_file_cache):
    """Symlink tar file from cache to repository."""
    tar_file = "%s.tar" % commit
    repository_ctx.symlink(tar_file_cache, tar_file)

def load_metadata_from_cache(repository_ctx, commit, commit_metadata_file_cache):
    """Read metadata from cache."""
    commit_metadata = repository_ctx.read(commit_metadata_file_cache)
    return commit_metadata

def write_defs(repository_ctx, metadata_map):
    defs_bzl_content = "%s_metadata = %r\n" % (repository_ctx.name, metadata_map)
    repository_ctx.file("defs.bzl", defs_bzl_content)

def write_build(repository_ctx, commits):
    srcs = ",\n".join(['"%s.tar"' % commit for commit in commits])
    build_contents = (
        'package(default_visibility = ["//visibility:public"])\n' +
        'filegroup(\n' +
        'name="%s",\n' % repository_ctx.name +
        'srcs=[\n%s\n],\n' % srcs +
        ')\n'
    )
    repository_ctx.file("BUILD.bazel", build_contents)

def uuidgen(repository_ctx):
    """Return a uuid"""
    exec_result = repository_ctx.execute(["uuidgen"])
    if exec_result.return_code != 0:
        fail("error running uuidgen")
    return exec_result.stdout.strip()

def write_json_file(repository_ctx, obj, path):
    obj_json = json.encode_indent(obj)
    repository_ctx.file(path, obj_json)

git_archive = repository_rule(
    implementation = _git_archive_impl,
    attrs = {
        "commits": attr.string_list(
            mandatory = False,
            doc = "An optional list of full Git commit hashes whose archive should be created. Combined with the srcs attribute.",
        ),
        "commit_metadata_source_files": attr.string_dict(
            mandatory = False,
            doc = "A dict whose values are names of files. For each commit and file, the file will be read from the git repository, making the content available in defs.bzl.",
        ),
        "remote": attr.string(
            mandatory = True,
            doc = "The Git remote from which to pull the Git commits. Git must be able to pull from the remote without asking for a password.",
        ),
        "srcs": attr.label_list(
            allow_files = True,
            mandatory = False,
            doc = "A list of Universe Config files. All referenced Git commit hashes will be archived. Combined with the commits attribute.",
        ),
    },
    doc = "A repository rule that runs 'git archive' to generate a tar file from each of multiple commits.",
)
