.. _testing_guide:

Testing Guide
#############

How to run E2E tests
--------------------

#.  In your pull request (PR), click the Labels icon and add the labels corresponding to the tests you want to run.

    #.  **test-bazel-large**: VMaaS E2E test. This takes about 15 minutes.
        All PRs that might impact VMaaS or its dependencies (including most global services) should run this test at least once.

    #.  **bazel-bm**: BMaaS E2E test. This takes about 65 minutes.

#.  Rerun the Jenkins pipeline. You can do this by pushing an empty commit (``git commit --allow-empty -m empty && git push``).

Troubleshooting failed E2E tests
--------------------------------

If any of E2E tests fail, use the following steps to troubleshoot it.

#.  In your PR, click the Details link next to the failed test
    (*ci/cloudbees/stage/Build/Bazel Test Large (VM & BM)/Bazel Test Large BM* or 
    *ci/cloudbees/stage/Build/Bazel Test Large (VM & BM)/Bazel Test Large VM*).

#.  Click on the *Go to classic* icon at the top-right corner of the page.

#.  Click on *Build Artifacts*.

#.  Navigate to the folder corresponding to the failed test such as ``bazel-testlogs/go/test/compute/e2e/bm/bm_test``.

#.  Right click on the file ``test.log`` and choose "Save link as..." to download the file locally.
    You can also view the file in your browser but because this file can be very large, your experience may be poor.

#.  Open the saved file in your favorite file viewer.
    ``less`` is recommended for viewing large log files.

#.  The end of the log file usually shows the test cleanup process which is not relevant.
    Instead, search for ``~~~ IF ANY TESTS FAILED, DETAILS WILL BE LOGGED ABOVE THIS LINE. ~~~``,
    then scroll *up* to view the reason for the test failure.

#.  To see pod logs:

    #.  Search through ``test.log`` for the text ``testEnvironmentId``.
        The value is a random string of 8 characters such as "90a0bd1a".

    #.  Login to `Kibana <https://internal-placeholder.com/s/idc/app/r/s/FeOaT>`__.
        Credentials are in `Vault <https://internal-placeholder.com/ui/vault/secrets/o11y/kv/projects%2Fidc%2Fadmin/details?version=1>`__.
        See the `IDC O11y Wiki <https://internal-placeholder.com/x/WA-Msw>`__ for more information.

    #.  In the search box, type ``Resource.deployment.id:90a0bd1a``.
        You may also find it useful to filter for ``Resource.deployment.environment:test-e2e-compute-bm``.

    #.  During E2E tests, the CronJob debug-tools-collector runs every 3 minutes and runs commands such as 
        ``kubectl get pods`` and ``kubectl get baremetalhosts``.
        The output is logged and can be viewed by filtering field ``cronjob`` for the value ``debug-tools-collector``.
        Each command runs in a different container, which you can filter for.
        Refer to the 
        `debug-tools chart values <https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/main/deployment/charts/debug-tools/values.yaml>`__
        for the list of commands.

#.  You can temporarily disable all sections of the Jenkins pipeline by editing the file ``/Jenkinsfile``
    and changing ``IS_PROD = true`` to ``IS_PROD = false``.
    Then enable only the "Check Pull Request", "Bazel Test Large VM", and "Bazel Test Large BM" stages,
    by changing the associated ``expression { IS_PROD }`` to ``expression { true }``.

    You may commit these changes into your branch but
    **be sure to revert these changes before marking your PR as Ready for review.**

#.  You can force the test to run on a particular Jenkins agent by changing
    ``agent { label 'BazelBM' }`` to ``agent { label 'TaaSBM01' }``.

#.  To get kubectl access to the cluster during the test execution, you can SSH into the Jenkins agent
    and run ``kind export kubeconfig --name idc-global``.
    You can then run ``kubectl`` or ``k9s`` to interact with the Kubernetes cluster.

#.  You may also want to view the real-time Bazel test logs by viewing the log file in the path such as
    ``/home/sdp/workspace/BMAAS-Orchestrator_PR-9348_2/bazel-testlogs/go/test/compute/e2e/bm/bm_test/test.log``.
