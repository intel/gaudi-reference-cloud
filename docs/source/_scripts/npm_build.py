# !/usr/bin/env python3
# Author: cristian.arias.chaves@intel.com
import os, sys
import glob
import shutil

curr_dir = os.getcwd()
build_path = os.path.join(curr_dir, 'source/_build/html')

def copy_node_packages():
    # NPM libraries that must be included
    npm_packages = ['@mermaid-js', 'mermaid', 'd3']

    print(f"Copying npm packages in path: {build_path}")

    for package in npm_packages:
        packageSourcePath = os.path.join('node_modules', package)
        packageCopyPath = os.path.join(build_path, '_static/js', package)
        if os.path.exists(packageCopyPath):
            shutil.rmtree(packageCopyPath)
        shutil.copytree(packageSourcePath, packageCopyPath)

    print(f"Finish copying npm packages to path {build_path}")


def replace_cdns():
    # This list of tuples with the form <original, replacement> that needs to be replaced
    # Use only if path cannot be overwritten in conf.py file
    replacement_array = [
        # This line remove requirejs as is not needed in web browsers
        (b'<script crossorigin="anonymous" integrity="sha256-Ae2Vz/4ePdIu6ZyI/5ZGsYnb+m0JlOmKPjt6XZ9JJkA=" src="https://cdnjs.cloudflare.com/ajax/libs/require.js/2.3.4/require.min.js"></script>',b''),
        (b'navbar-expand-lg', b'navbar-expand-xl')
    ]

    # If NODE_ENV is not present at 'make html', allow None and complete build; this is the case for AWS S3 build
    NODE_ENV = os.environ.get('NODE_ENV', None)

    # **Important** If NODE_ENV is set for dev, build won't work to S3 Buckets, just for local server
    # **Important** The second filepath below must start with /source because it is where parent Makefile resides
    if NODE_ENV == 'dev':
        replacement_array.append((b'/docs/_static/', b'/source/_build/html/_static/'))
    else:
        pass

    print(f"Replacing CDNs in path: {build_path}")

    search_path = os.path.join(build_path, '**/*.html')
    for filepath in glob.iglob(search_path, recursive=True):
            with open(filepath, 'rb') as file:
                s = file.read()
            for (original, replacement) in replacement_array:
                s = s.replace(original, replacement)
                with open(filepath, 'wb') as file:
                    file.write(s)

    print(f"Finish replacement of CDNs in path {build_path}")

def main():
    copy_node_packages()
    replace_cdns()
    return

if __name__ == "__main__":
    main()