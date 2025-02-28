.. _bazel:

Bazel
#####

This document provides tips for using Bazel in IDC.

Debugging Remote Cache Hits
***************************

If Bazel performance is poor, Bazel caching should be investigated.

Refer to `Bazel remote cache`_ for the general technique.
Below are the specific steps that can be used with IDC.

Create Bazel Execution Logs
===========================

#.  Create a new branch in the monorepo and check it out.

#.  Edit Makefile target test-universe-deployer.
    Add the lines with "clean" and "execution_log_binary_file":

    .. code:: console

      test-universe-deployer: bazel ## Run Universe Deployer unit tests.
         $(BAZEL) clean
         $(BAZEL) test $(BAZEL_OPTS) \
            --execution_log_binary_file=local/exec.log \
            //deployment/universe_deployer:universe_config_tests \

#.  Edit Jenkinsfile to specify a single agent.

    .. code:: console

      agent {label 'pdx75-c01-baci001'}

#.  Touch files to force Jenkins to run relevant stages.

    .. code:: bash

      touch go/touch
      touch universe_deployer/touch

#.  Push to Github and create a PR.

#.  Use Jenkins Workspaces to download the file local/exec.log.
    Rename this file to exec1.log.

#.  Repeat all of the above steps with a different branch, different PR, and ensure you have different commit hashes.
    Rename the final file to exec2.log.

Compare Execution Logs
======================

#.  Upload execution logs to ${HOME}/idc/local.

#.  Build Bazel execlog parser.

    .. code:: bash

      cd ${HOME}
      git clone https://github.com/bazelbuild/bazel
      cd bazel
      git checkout 5.4.0
      echo USE_BAZEL_VERSION=5.4.0 > .bazeliskrc
      bazel build --java_runtime_version=remotejdk_11 src/tools/execlog:parser

#.  Run Bazel execlog parser.

    .. code:: bash

      bazel-bin/src/tools/execlog/parser \
      --log_path=${HOME}/idc/local/exec1.log \
      --log_path=${HOME}/idc/local/exec2.log \
      --output_path=${HOME}/idc/local/exec1.txt \
      --output_path=${HOME}/idc/local/exec2.txt

#.  Compare execlog output.

    .. code:: bash

      diff \
      ${HOME}/idc/local/exec1.txt \
      ${HOME}/idc/local/exec2.txt \
      > ${HOME}/idc/local/exec12.txt

#.  Investigate any differences shown in exec12.txt.
    In particular, differences in environment variables will result in different cache keys.

See Also
********

-  `PR #7769 Improve Bazel cache hit rate with incompatible_strict_action_env <https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/pull/7769>`__



.. _Bazel remote cache: https://bazel.build/remote/cache-remote
