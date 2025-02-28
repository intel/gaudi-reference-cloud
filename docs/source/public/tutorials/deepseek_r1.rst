:alttitle: Run DeepSeek-R1-Distill-Llama-70B with TGI on Intel XPUs
:category: compute
:keywords: DeepSeek, DeepSeek-R1, Llama 3, TGI
:phrase: Run DeepSeek-R1-Distill-Llama-70B using text-generation-inference (TGI) with the Intel® Data Center GPU Max Series.
:rating: 0
:show_urls: true

.. _deepseek_r1:

Run DeepSeek-R1-Distill-Llama-70B
#################################

Learn to run DeepSeek-R1-Distill-Llama-70B using text-generation-inference (TGI) on Intel XPUs. You may also experiment with using smaller `DeepSeek-R1-Distill Models`_.

The DeepSeek-R1-Distill-Llama-70B model is a distilled version of the DeepSeek-R1 model, derived from the Llama-3.3-70B-Instruct architecture. It's designed to emulate the reasoning capabilities of the original 671 billion parameter model while being more efficient. Balancing performance and efficiency, it's a great choice for complex tasks with reduced computational requirements.

Prerequisites
*************

* Complete :ref:`get_started`
* Review and agree to `DeepSeek-R1 License`_

Compute Instance
****************

.. list-table:: Recommended Compute Instances
   :widths: auto
   :header-rows: 1

   * - Instance Type
     - Processor Model
     - Card Quantity
     - Disk
     - Memory
     - Link

   * - Bare Metal
     - |GPUMAX| 1550
     - 8
     - 2TB
     - 2TB
     - `Go to 1550`_

   * - Bare Metal
     - |GPUMAX| 1100
     - 8
     - 960GB
     - 1TB
     - `Go to 1100`_

Launch Instance
***************

#. Visit the |ITAC| `console`_ home page.

#. Log into your account.

#. Click :guilabel:`Catalog -> Hardware` from the menu at left.

#. Click the filter :guilabel:`GPU`.

#. Select the **instance type**: |GPUMAX|

#. Complete :guilabel:`Instance configuration`. Use one example from `Compute Instance`_ details.

   a. Use one of these configurations: |GPUMAX| 1550 BM or |GPUMAX| 1100 BM.

   #. For :guilabel:`Instance type`, choose one prefixed with bare metal, or ``BM``.

   #. For :guilabel:`Machine image`, use default.

#. Add :guilabel:`Instance name`.

#. Choose an option to connect.

   a. :guilabel:`One-Click connection` **Recommended**

   #. :guilabel:`Public Keys`

#. Click :guilabel:`Launch` to launch your instance.

.. tip::
   See also :ref:`manage_instance`.

Launch Container
****************

Next, access your instance and launch the TGI container. We use the |GPUMAX| 1550 as an example.

.. note::
   For details on the |INTC| Max Series product family, see :ref:`gpu_instances`.

.. code-block:: bash

   docker run -it --rm \
   --privileged \
   --device=/dev/dri \
   --ipc=host \
   --ulimit memlock=-1 \
   --shm-size=1g \
   --cap-add=sys_nice \
   --cap-add=IPC_LOCK \
   -v ${HF_CACHE_DIR:-$HOME/.cache/huggingface}:/root/.cache/huggingface:rw \
   -e HF_HOME=/root/.cache/huggingface \
   -p 80:80 \
   --entrypoint /bin/bash \
   ghcr.io/huggingface/text-generation-inference:3.0.2-intel-xpu

Start Model Server
******************

#. To treat slices on |GPUMAX| 1550 as one, set this in your shell:

   .. code-block:: bash

      bashexport ZE_FLAT_DEVICE_HIERARCHY=COMPOSITE

#. In the container terminal, launch the model:

   .. code-block:: bash

      MODEL_ID=deepseek-ai/DeepSeek-R1-Distill-Llama-70Btext-generation-launcher \
      --model-id ${MODEL_ID} \
      --dtype bfloat16 \
      --max-concurrent-requests 128 \
      --max-batch-size 128 \
      --max-total-tokens 4096 \
      --max-input-length 2048 \
      --max-waiting-tokens 10 \
      --cuda-graphs 0 \
      --num-shard=4 \
      --port 80 \
      --json-output

#. Wait for the model to be fully loaded. You should see this message:

   .. code-block:: console

      {"timestamp":"2025-01-30T20:05:22.031688Z","level":"INFO","message":"Connected"}

Benchmarking
************

If you need to Benchmark using the model, open a new terminal and follow these steps.

#. Get the running container ID:

   .. code-block:: bash

      docker ps --filter ancestor=ghcr.io/huggingface/text-generation-inference:3.0.2-intel-xpu --format "{{.ID}}"

#. Connect to the container, using the output from the previous step.

   .. code-block:: bash

      docker exec -it <CONTAINER_ID> bash

#. Run the benchmark:

   .. code-block:: bash

      MODEL_ID=deepseek-ai/DeepSeek-R1-Distill-Llama-70B
      text-generation-benchmark --tokenizer-name $MODEL_ID

The benchmark will run various configurations and and display output performance metrics when complete.

.. meta::
   :description: Run DeepSeek-R1-Distill-Llama-70B using text-generation-inference (TGI) with the Intel® Data Center GPU Max Series.
   :keywords: DeepSeek, DeepSeek-R1, Llama 3, TGI

.. collectfieldnodes::

.. _Go to 1550: https://console.cloud.intel.com/compute/reserve?backTo=catalog&region=us-region-2&instance-type=bm-spr-pvc-1550-8
.. _Go to 1100: https://console.cloud.intel.com/compute/reserve?backTo=catalog&region=us-region-2&instance-type=bm-spr-pvc-1100-8

.. _DeepSeek R1 on Hugging Face: https://huggingface.co/deepseek-ai/DeepSeek-R1#deepseek-r1-distill-models
.. _DeepSeek-R1-Distill Models: https://huggingface.co/deepseek-ai/DeepSeek-R1#deepseek-r1-distill-models
.. _DeepSeek-R1 License: https://huggingface.co/deepseek-ai/DeepSeek-R1#7-license
.. _console: https://console.cloud.intel.com