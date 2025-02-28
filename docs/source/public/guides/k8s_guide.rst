.. _k8s_guide:

Intel Kubernetes Service Guide
##############################

The Intel® Kubernetes Service (IKS) gives you the tools to manage Kubernetes clusters
for application development, AI/ML training, and helm chart deployments.

.. tip::
   Currently IKS is only available to premium and enterprise account users.

**Control Plane**

IKS provides managed Kubernetes service in |ITAC|. IKS manages the availability and scalability of the Kubernetes control plane.
For a technical overview, see also Kubernetes `Control Plane Components`_.

Provision Kubernetes Cluster
****************************

Create a Cluster
================

#. Navigate to the |ITAC| console.

#. In the menu at left, click the :guilabel:`Intel Kubernetes Service` menu.

#. Visit the :guilabel:`Overview` tab to view the workflow.

#. Click :guilabel:`Clusters` tab.

#. Click :guilabel:` Launch Cluster`.

#. Complete the required fields under :guilabel:`Cluster details and configuration`.

   a. In :guilabel:`Cluster name`, enter a name.
   #. In :guilabel:`Select cluster K8S version`, select a version.

#. Click :guilabel:`Launch`. After launching, the :guilabel:`State` column shows **Updating**.

#. Under :guilabel:`Cluster Name` column, click your cluster.

   .. note::
      Now your :guilabel:`Cluster name` with :guilabel:`Actions` menu appears below.

Add Node Group to Cluster
=========================

#. From the :guilabel:`Actions` pulldown menu, select :guilabel:`Add node group`.

#. Enter your data in the :guilabel:`Node group configuration` menu.

   a. In :guilabel:`Node type`, choose between Virtual Machine Bare Metal for your node.
      Note the cost per hour. See also `Compare Instance Types`_ below.

   #. In :guilabel:`Node group name`, enter a name.

   #. In :guilabel:`Node quantity`, choose a quantity from 1 to 10.
      Select the number of worker nodes you need in your cluster.

   .. tip::
      You can scale the number of worker nodes up or down.

#. Under :guilabel:`Public Keys`, select :guilabel:`Upload Key` or :guilabel:`Refresh Keys`.

#. Select :guilabel:`Upload Key`, name your key and copy your local SSH public key in the fields shown.

#. Select :guilabel:`Upload Key`.

#. Now, in :guilabel:`Node group configuration`, check the box next to the SSH key you added.

Compare Instance Types
-------------------------

At any time during :guilabel:`Node group configuration`, you may choose :guilabel:`Compare instance types`.
This pop-out screen helps you compare and select your preferred processor.

Launch Kubernetes Cluster
==========================

When you create a cluster, it includes:

* K8S Control-plane
* ETCD Database
* Scheduler
* API Server

.. tip::
   See also :ref:`compute_svc`

#. Select :guilabel:`Launch`.

#. Now that your **Node group** is added, it shows **Updating** in submenu.

#. When adding your **Node Group** is successful, each :guilabel:`Node name` appears and its
   :guilabel:`State` shows :guilabel:`Active`.

Connect to cluster
===================

#. Set the :file:`KUBECONFIG` Environment Variable:

   .. tabs::

      .. tab:: Linux\* OS / macOS\*

         .. code-block:: bash

            export KUBECONFIG=/path/to/your/kubeconfig

      .. tab:: PowerShell

         .. code-block:: bash

            $Env:KUBECONFIG = "C:\path\to\your\kubeconfig"

..
   TODO: When self-service is resolved with firewall-whitelisting,
   add instructions - check with sacashgit.

#. Verify Configuration: Ensure that the current context points to the correct cluster.

   .. code-block:: bash

      kubectl config view

Kubeconfig Admin Access
=======================

Ideally, you export the :file:`KUBECONFIG` to your secret management system and continue.

#. In the :guilabel:`Kubernetes Console`, locate options below :guilabel:`Kube Config`.
#. **Copy** or **Download** the :file:`KUBECONFIG` file and export it to your development environment.
#. For more help on exporting, follow related steps in the next section.

.. caution::

   Exercise caution while downloading, accessing, or sharing this file.


Set Context for Multiple Clusters
=================================

#. Optional: List all available contexts.

   .. code-block:: bash

      kubectl config get-contexts -o=name

#. Change directory, or create one if it doesn't exist.

   .. code-block:: bash

      cd ./kubeconfig

   .. code-block:: bash

      mkdir ./kubeconfig

#. In the :guilabel:`Kubernetes Console`, navigate to :guilabel:`My clusters`, :guilabel:`Kube Config`.

#. From the :guilabel:`Kubernetes Console`, download (or copy) the KUBECONFIG file to the current directory.

#. Extract the **value** from the KUBECONFIG and paste it into the shell, following the **example below**.

   a. Export KUBECONFIG as an environment variable as shown below.

   .. code-block:: bash

      export KUBECONFIG =/home/sdp/.kube/dev-env

#. Use :command:`kubectl config set context` to modify an existing context or create a new cluster context.

   .. code-block:: bash

      kubectl config set context

#. To view them, enter command.

   .. code-block:: bash

      kubectl get nodes

.. important::

   If you wish to launch another cluster, return to the start of this section and perform all steps again, exporting a different KUBECONFIG file.

Controlling Node Auto-repair Behavior
=====================================

By default, IKS auto-detects the worker node's **unavailability**. If it's unavailable beyond a specific grace period, it will automatically be replaced (`auto-repair`) with a fresh new node of the same type.  If you do not desire this behavior for one or more worker nodes in your cluster, you may turn off the `auto-repair` functionality for any given worker node.

Auto-repair Options
-------------------

If you want to opt out of auto repair mode (where a node will be automatically replaced when it becomes unavailable / unreachable after a grace period elapsed) then you must label the given node with :command:`autorepair=false`.

As long as the node has this label, IKS will not replace the node if it becomes unavailable. The user interface will show the status as :guilabel:`Updating` when unavailable (and not ready in kubernetes), a sign to show it detected an unavailability of a node.  If the node becomes available later, the status will change from :guilabel:`Updating`` to :guilabel:`Active`. During the unavailability of a node, if you remove the `auto-repair` label , then default behavior of auto-replacement of the node resume and IKS will replace the node, as designed.

We do not recommend removing the node from compute console when this label is On (defeats the purpose of a label in the first place). It will result in dangling node in your Kubernetes console.

.. note::
   You can label the node using the out of the box adding and removing label functionality using `kubectl​` commands.

Examples
--------

Add a label to a node to avoid auto replacement:

.. code-block:: bash

   kubectl label node ng-hdmqnphxi-f49b8 iks.cloud.intel.com/autorepair=false

Remove a label from a node to enable auto replacement:

.. code-block:: bash

   kubectl label node ng-hdmqnphxi-f49b8 iks.cloud.intel.com/autorepair-


Manage Kubernetes Cluster
**************************

#. Create a pod.

   .. code-block:: bash

      kubectl apply -f pod-definition.yaml

#. Create a :file:`YAML`` or :file:`JSON` file with your pod specificationsr. See example below.

   .. code-block:: yaml

      apiVersion: v1
      kind: Pod
      metadata:
      name: mypod
      spec:
      containers:
      - name: mycontainer
         image: nginx

#. Replace "mypod" with the name of your pod.

   .. code-block:: bash

      kubectl get pods kubectl describe pod mypod

#. Update a Pod:

   .. code-block:: bash

      kubectl edit pod mypod

   .. note::
      This opens the pod configuration in your default editor. Make changes and save the file.

#. Delete a Pod. Replace mypod with the name of your pod.

   .. code-block:: bash

      kubectl delete pod mypod

Upgrade Kubernetes Cluster
===========================

#. In the :guilabel:`Cluster name`, :guilabel:`Details`, find the :guilabel:`Upgrade` link.

#. Select :guilabel:`Upgrade`.

#. In the :guilabel:`Upgrade K8S Version`, pull-down menu, select your desired version.

#. Click the :guilabel:`Upgrade` button.

#. During the upgrade, the :guilabel:`Details` menu :guilabel:`State` may show **Upgrading controlplane**.

   .. note::
      If the current version is penultimate to the latest version, only the latest version appears.
      When the version upgrade is successful, :guilabel:`Cluster reconciled` appears.

Apply Load Balancer
===================

#. Navigate to the :guilabel:`Cluster name` submenu.

#. In the :guilabel:`Actions` menu, select :guilabel:`Add load balancer`.

#. In the :guilabel:`Add load balancer`, complete these fields.

   a. Select the port number of your service from the dropdown menu.
   #. For **Type**, select *public* or *private*.
   #. Click on :guilabel:`Launch`.

#. In the :guilabel:`Cluster name` submenu, view the :guilabel:`Load Balancer` menu.

#. Your Load Balancer appears with :guilabel:`Name` and :guilabel:`State` shows :guilabel:`Active` .

K8S will automatically perform load balancing for your service.


Add Security Rules
===================

You can create a security rule if you have already created a Load Balancer.

.. note::
   If you haven't created a Load Balancer, return to above section before proceeding.
   After a Cluster is available, you must create a Node Group.

#. Click on your Cluster name.

#. Select the tab :guilabel:`Worker Node Group`.

#. Select :guilabel:`Add Node Group`.

#. Complete all required fields as shown in `Add Node Group to Cluster`_.
   Then return to this workflow.

#. Wait until the :guilabel:`State` shows "Active" before proceeding.

#. Complete all steps in `Apply Load Balancer`_. Then return here.


Add security rule to your own Load Balancer
-------------------------------------------

#. For your own Load Balancer, click :guilabel:`Edit`.

#. Add an Source IP address to create a security rule.

#. Select a protocol.

#. Click :guilabel:`Save`. The rule is created.


Edit or delete security rule
----------------------------

Optional: After the :guilabel:`State` changes to :guilabel:`Active`:

* You may edit the security rules by selecting :guilabel:`Edit`.
* You may delete the security rule by selecting :guilabel:`Delete`.

Add security rule to default Load Balancer
------------------------------------------

#. Navigate to the Security tab. You may see Load Balancers populated in a table.

   .. note::
      The `public-apiserver` is the default Load Balancer.

#. For the `public-apiserver`, click "Edit".

#. Then add an Source IP address to create a security rule.

#. Select a protocol.

#. Click :guilabel:`Save` The rule is created.

..
   TODO: a.	Is there throttling? Does it have a protocol? Indicate if there is a load balancer algorithm (round-robin, etc.)?

Additional resources
====================

* `kubectl quick reference`_
* `kubectl documentation`_


Configure Ingress, Expose Cluster Services
*******************************************

.. note::
   This requires `helm` version 3 or a helm client utility. See also `Helm Docs`_.

#. Create a cluster with at least one worker node. See `Create a Cluster`_.

#. Create a Load balancer (public) using port 80. See `Apply Load Balancer`_.

   .. note::
      This IP is used in the last step in the URL for testing
      Your port number may differ.

#. Install the ingress controller.

   .. code-block:: bash

      helm upgrade --install ingress-nginx ingress-nginx --repo https://kubernetes.github.io/ingress-nginx --namespace ingress-nginx --create-namespace --set controller.hostPort.enabled=true

#. To install test NGINX POD, Service, and Ingress object, download :download:`ingress-test.yml<../../_snippets/ingress-test.yml>`.

   a. Alternatively, copy the contents of file and save it as :file:`ingress-test.yml`

      .. literalinclude:: ../../_snippets/ingress-test.yml
         :language: yaml

#. Run command to apply.

   .. code-block::  yaml

      kubectl apply -f ingress-test.yaml

#. Visit your browser and test, inserting your IP where shown below.

   .. code-block:: bash

      http://<IP>/test

   #. The IP mentioned here is the Public Load balancer IP.

De-Provision Kubernetes Cluster
*******************************

Delete Cluster Group or Node
============================

Delete Node Group
=================

#. In the :guilabel:`Cluster name` submenu select the :guilabel:`Node group` you wish to delete.
#. Click :guilabel:`Delete` button.

Delete Node
===========

#. Below the :guilabel:`Node name` table, note :guilabel:`Add node` and :guilabel:`Delete node`
#. Click :guilabel:`Delete node` button, as desired.
#. Select :guilabel:`Continue`.


Deploy Example AI/ML Workloads
******************************

Add instance of |INTG2| to a cluster to deploy LLM and Stable Diffusion models.

#. Complete the tutorial `Training a PyTorch Model on Intel Gaudi 2`_.

#. Add nodes to the Intel Kubernetes Cluster.

#. Assure you're able to access the KUBECONFIG file and the Kubernetes Cluster.

   .. seealso::
      `Kubeconfig Admin Access`_

Deploy Stable Diffusion
=======================

To deploy with Stable Diffusion, try an example below. Run this on a |INTG2| instance and deploy it on an IKS cluster.

|INTG2| with Stable Diffusion
-----------------------------

To run Stable diffusion in IKS with |INTG2|, apply the following configuration.

#. Apply configuration if huge pages **is not set** in all nodes.
   Otherwise, skip to the next section.

   .. code-block:: bash

      sudo sysctl -w vm.nr hugepages=156300

#. Verify configuration.

   .. code-block:: bash

      grep HugePages Free /proc/meminfo

      grep HugePages Total /proc/meminfo

#. Esnure that your output is similar to this.

   .. code-block:: console

      HugePages_Free:    34142

      HugePages_Total:   35201

#. Use the suggested settings for model inference.

   .. code-block:: Console

      hugepages2Mi: 500Mi
      memory: 60G

#. Revise your YAML file, using this example.

   .. literalinclude:: ../../_snippets/std-gaudi.yml
      :language: yaml

HugePages Settings by Model
----------------------------

.. list-table:: HugePages Settings
   :header-rows: 1
   :class: table-tiber-theme

   * - Model Name
     - hugepages-2Mi
     - Memory
     - Number of Cards

   * - *runwayml/stable-diffusion-v1-5*
     - 500Mi
     - 6OG
     - 1

   * - *meta-llama/Meta-Llama-3-70B-Instruct*
     - 9800Mi
     - 250G
     - >= 2

   * - *mistralai/Mixtral-8x7B-Instruct-v0.1*
     - 9800Mi
     - 250G
     - >= 2

   * - *mistralai/Mistral-7B-v0.1*
     - 600Mi
     - 5OG
     - 1

Generate Image with Stable Diffusion
-------------------------------------

Consider using this YAML deployment for Helm Chart resources.

#. Download the Helm Charts from the `STD Helm Charts`_.

#. Configuration for hugepages, as noted above, is already applied.

   .. note::
      This YAML file overrides default configuration. Apply your custom configuration
      to this file to ensure your settings are applied.

   .. literalinclude:: ../../_snippets/std-values.yml
      :language: yaml

#. Next, run the install command.

   .. code-block:: bash

      helm install std std-chart -f ./std-values.yaml

#. Access the result using the load balancer IP.

   .. note::
      Ensure you followed the section `Apply Load Balancer`_.

#. Construct a full URL for the Load Balancer by following this two-step process.

   a. Replace the value of :file:`<Load Balancer IP>` with your own, as shown below.

      .. code-block:: console

         http://<Load Balancer IP>/std/generate_image

   #. Add the prompt, including parameters, as the second part of the URL.

      Example: The second part starts with "prompts="

      .. code-block:: console

         http://<Load Balancer IP>/std/generate_image/prompts=dark sci-fi , A huge radar on mountain ,sunset, concept art&height=512&width=512&num_inference_steps=50&guidance_scale=7.5&batch_size=1&negative_prompts=''&seed=100&num_images_per_prompt=1

   #. Paste the full URL in a browser and press :kbd:`<Enter>`.

   #. Change the value of "prompts=", as desired.

      Example 2: Change the second part of the URL. Replace the text, starting with "prompts=", as shown below.

      .. code-block:: console

         http://<Load Balancer IP>/std/generate_image/prompts=Flying Cars&height=512&width=512&num_inference_steps=50&guidance_scale=7.5&batch_size=1&negative_prompts=''&seed=100&num_images_per_prompt=1

   #. Paste the full URL in a browser and press :kbd:`<Enter>`.

      .. tip::
         Your image will differ. Any image that you generate may require managing copyright permissions.

See `Helm Docs`_ for more details.

Generate Text with Stable Diffusion
-------------------------------------

Consider using this sample YAML deployment for Text Generation Interface (TGI).
Refer to `HugePages Settings by Model`_.

.. note::
   To use this sample template, you must provide your own ``HUGGING_FACE_HUB_TOKEN`` value.

.. literalinclude:: ../../_snippets/tgi-lama3.yml
   :language: yaml

#. Download the `TGI Helm Charts`_.

#. To deploy TGI with Mistral with Helm:

   .. code-block:: bash

      helm install mistral tgi-chart -f ./mistral-values.yaml

   .. note::
      See also `Huggingface Text Generation Inference`_ and `text-generation-launcher arguments`_ .

#. Access the result with the load balancer IP.

   a. Follow the section `Apply Load Balancer`_.

#. Replace the value of :file:`<Load Balancer IP>`, shown below, with your own.

   `http://<Load Balancer IP>/mistral/generate`

.. _text-generation-launcher arguments: https://huggingface.co/docs/text-generation-inference/en/basic_tutorials/launcher
.. _Huggingface Text Generation Inference: https://github.com/huggingface/text-generation-inference
.. _Helm Docs: https://v3.helm.sh/
.. _kubectl quick reference: https://kubernetes.io/docs/reference/kubectl/quick-reference/
.. _kubectl documentation: https://kubernetes.io/docs/home/
.. _STD Helm Charts: https://github.com/rajeshkumarramamurthy/genailab/tree/main/gaudi/std-chart
.. _TGI Helm Charts: https://github.com/rajeshkumarramamurthy/genailab/tree/main/gaudi/tgi-chart
.. _Control Plane Components: https://kubernetes.io/docs/concepts/overview/components/#control-plane-components
.. _Training a PyTorch Model on Intel Gaudi 2: https://docs.habana.ai/en/latest/Intel_DevCloud_Quick_Start/Intel_DevCloud_Quick_Start.html
