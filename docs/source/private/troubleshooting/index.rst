.. _troubleshooting:

Troubleshooting
###############

Intro paragraph. TBD

Attach a debug container to an existing pod
*******************************************

.. code-block:: bash

    kubectl debug -it -n idcs-system grpc-proxy-7d8659575b-fbv4x --target opa-envoy --image=nicolaka/netshoot@sha256:a7c92e1a2fb9287576a16e107166fee7f9925e15d2c1a683dbb1f4370ba9bfe8


Start a new debug pod
***********************

.. code-block:: bash

    kubectl run -it --rm debug-shell --image=nicolaka/netshoot@sha256:a7c92e1a2fb9287576a16e107166fee7f9925e15d2c1a683dbb1f4370ba


