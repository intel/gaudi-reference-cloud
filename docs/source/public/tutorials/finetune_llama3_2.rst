.. _finetune_llama3_2:

Fine-tune Meta Llama-3.2-3B-Instruct
####################################

Apply fine-tuning to the Meta\* `Llama-3.2-3B Instruct`_ model, using Low-Rank Adaptation (LoRA) with an Optimum Habana container workflow. Using a lightweight model, with gated access, explore the model's capabilities for translation and accuracy.

.. raw:: html

   <div>
      <ul class="arrow-workflow-container">
         <li class="arrow-box hanging-indent">1. Choose model and processor</li>
         <li class="arrow-box hanging-indent">2. Configure environment</li>
         <li class="arrow-box hanging-indent">3. Launch instance</li>
         <li class="arrow-box hanging-indent">4. Prepare and load data</li>
         <li class="arrow-box hanging-indent is-active">5. Train and evaluate model</li>
         <li class="arrow-box hanging-indent">6. Deploy</li>
         <li class="arrow-box hanging-indent">7. Monitor and optimize</li>
      </ul>
   </div>

.. tip::
   A pre-trained model is used for demonstration. The model is fine-tuned using a causal language modeling (CLM) loss.
   See `Casual language modeling from Hugging Face`_.

Prerequisites
*************

* Complete :ref:`get_started`
* HuggingFace account
* Agree to `Terms of Use`_ of the LLama 3.2-3B model
* Create a read token and request access to the Meta Llama 3.2-3B model

Request Access to Llama3.2-3B-Instruct Model
*********************************************

#. Log into your HuggingFace account.

#. Navigate to the `Llama-3.2-3B Instruct`_ model page.

   .. note::
      This tutorial is validated on Llama-3.2-3B Instruct using a token method for a gated model. See also `Hugging Face Serving Private & Gated Models`_.

#. View the :guilabel:`Model card` tab.

#. Complete a request to access to the **Llama3.2-3B-Instruct Model**.

#. In your Hugging Face account, review :guilabel:`Settings` to confirm your request is :guilabel:`Accepted`.

Compute Instance
****************

.. list-table:: Compute Instance Used
   :widths: auto
   :header-rows: 1
   :class: table-tiber-theme

   * - Instance Type
     - Recommended Processor
     - Cards
     - RAM
     - Disk

   * - Bare Metal (BM)
     - |INTG2|
     - 8 Gaudi 2 HL-225H mezzanine cards with 3rd Gen Intel Xeon® processors
     - 1TB
     - 30TB

Log in and Launch Instance
**************************

#. Visit the |ITAC| `console`_ home page.

#. Log into your account.

#. Click the :guilabel:`Hardware Catalog` from the menu at left.

#. Select the **instance type**: Gaudi2 Deep Learning Server.

#. Complete :guilabel:`Instance configuration`. See `Compute Instance`_ details.

   a. :guilabel:`Instance type`. See `Compute Instance`_ above.

#. :guilabel:`Machine image`. See `Compute Instance`_ above.

#. Add :guilabel:`Instance name` and :guilabel:`Public Keys`.

#. Click :guilabel:`Launch` to launch your instance.

.. tip::
   See also :ref:`manage_instance`.

Run Habana container workflow
*****************************

After launching a bare metal instance, run the container workflow.

#. Launch your instance in a command line interface (CLI).

#. Visit `Habana Prebuilt Containers`_, and copy and run the *latest* commands:

   * Pull Docker image
   * Run Docker image

#. Pull the latest :file:`gaudi-docker` image. **Example code** is shown below.

   .. code-block:: bash

      docker pull vault.habana.ai/gaudi-docker/1.18.0/ubuntu22.04/habanalabs/pytorch-installer-2.4.0:latest

#. Run the habana docker image in **interactive mode**. See example below.

   .. code-block:: bash

      docker run -it \
      --runtime=habana \
      -e HABANA_VISIBLE_DEVICES=all \
      -e OMPI_MCA_btl_vader_single_copy_mechanism=none \
      -v /opt/datasets:/datasets \
      --cap-add=sys_nice \
      --net=host \
      --ipc=host vault.habana.ai/gaudi-docker/1.18.0/ubuntu22.04/habanalabs/pytorch-installer-2.4.0:latest

#. Run the following lines, one code block at a time.

   .. code-block:: bash

      git clone https://github.com/huggingface/optimum-habana.git

   .. code-block:: bash

      cd optimum-habana && python3 setup.py install

   .. code-block:: bash

      cd examples/language-modeling && pip install -r requirements.txt

#. Continue below.

Export Hugging Face Token
**************************

You must get a Hugging Face token for gated model access. This tutorial uses a command line interface (CLI) method.
You are still working in the container, continued from above.

#. Follow steps in `Hugging Face Serving Private & Gated Models`_.

#. In your CLI, export token as an environment variable like shown below.

   a. Paste in your own token, replacing "Your_Huggingface_Token".

   .. code-block:: bash

      export HF_TOKEN="Your_Huggingface_Token"

#. Optional: Show token to confirm that the variable is set.

   .. code-block:: bash

      echo $HF_TOKEN

Apply Single-card Finetuning
****************************

The dataset is taken from `Stanford Alpaca repo`_. Note the flagged commands below.

#. Run single-card fine-tuning of **Meta-Llama-3.2-3B-Instruct**.

   .. code-block:: bash

      python3 run_lora_clm.py \
      --model_name_or_path meta-llama/Llama-3.2-3B-Instruct \
      --dataset_name tatsu-lab/alpaca \
      --bf16 True \
      --output_dir ./model_lora_llama \
      --num_train_epochs 3 \
      --per_device_train_batch_size 16 \
      --evaluation_strategy "no" \
      --save_strategy "no" \
      --learning_rate 1e-4 \
      --warmup_ratio  0.03 \
      --lr_scheduler_type "constant" \
      --max_grad_norm  0.3 \
      --logging_steps 1 \
      --do_train \
      --do_eval \
      --use_habana \
      --use_lazy_mode \
      --throughput_warmup_steps 3 \
      --lora_rank=8 \
      --lora_alpha=16 \
      --lora_dropout=0.05 \
      --lora_target_modules "q_proj" "v_proj" \
      --dataset_concatenation \
      --max_seq_length 512 \
      --low_cpu_mem_usage True \
      --validation_split_percentage 4 \
      --adam_epsilon 1e-08

   .. tip::
      This may take some time. Your results may differ.

#. Observe how the **rate of loss** decreases over time during model fine-tuning.

#. At roughly 20% completion, **loss shows: 1.0791**.

   .. figure:: ../../_figures/tutorials/finetune_llama3.2-3B_20_02.png
      :alt: Rate of loss at 20% completion
      :align: center

      Rate of loss at 20% completion

#. While at roughly 80% completion, **loss shows: 0.9827**.

   .. figure:: ../../_figures/tutorials/finetune_llama3.2-3B_80_03.png
      :alt: Rate of loss at 80% completion
      :align: center

      Rate of loss at 80% completion

#. View the full results of fine-tuning below.

   .. figure:: ../../_figures/tutorials/finetune_llama3.2-3B_100_04.png
      :alt: Example - Finetuning final results
      :align: center

      Example - Finetuning final results

View Results in JSON
********************

While optional, this section shows how to view or save the output for training and evaluation.

#. Change directory to the  :command:`output_dir`, as defined previously in :command:`python3 run_lora_clm.py`

   .. code-block:: bash

      cd model_lora_llama

#. Look for the updated weights in the output directory, :file:`model_lora_llama`.

#. View **All** results.

   .. code-block:: bash

      cat all_results.json

#. Run command to view only **evaluation** results.

   .. code-block:: bash

      cat eval_results.json

#. Your evaluation output may differ.

   .. code-block:: console

      "epoch": 3.0,
      "eval_accuracy": 0.7230934426629573,
      "eval_loss": 1.0917302370071411,
      "eval_runtime": 7.0072,
      "eval_samples": 405,
      "eval_samples_per_second": 65.31,
      "eval_steps_per_second": 8.228,
      "max_memory_allocated (GB)": 91.15,
      "memory_allocated (GB)": 42.34,
      "perplexity": 2.979424726273657,
      "total_memory_available (GB)": 94.62

#. Run command to view only **training** results.

   .. code-block:: bash

      cat eval_results.json

#. Your training output may differ.

   .. code-block:: console

      "epoch": 3.0,
      "max_memory_allocated (GB)": 58.44,
      "memory_allocated (GB)": 42.35,
      "total_flos": 2.6248302923297587e+17,
      "total_memory_available (GB)": 94.62,
      "train_loss": 1.1588270471830313,
      "train_runtime": 992.6524,
      "train_samples_per_second": 31.11,
      "train_steps_per_second": 1.945

#. Optional: You may use the :file:`README.md` in the directory :file:`model_lora_llama` as a template to record test results.

ChatBot Prompt Exercise
***********************

In this exercise, you launch a bare Jupyter Notebook and use the Python 3 kernel to prototype a Chatbot. Given that Llama-3.2-3B-Instruct is optimized for multilingual dialogue use cases, let's test its ability to translate.

Your objective is to test translation accuracy on `Llama-3.2-3B Instruct`_ , based on a simple prompt.

.. tip::
   Feel free to change the prompt to make it your own. To do so, modify the text that appears in the *array of dictionaries* entitled  "messages". Specifically, modify the string **value** for the key **content**.

#. Click `Learning`_ in the console.

#. Select the button :guilabel:`Connect now`.

#. Click `AI Accelerator`.

#. In the JupyterLab interface, select :guilabel:`File -> New -> Notebook`

#. At the dialog :guilabel:`Select Kernel`, choose **Python 3**.

#. Next, copy and paste the code from each section below in your Jupyter Notebook.

Chat Prompt Notebook
====================

Follow along to test a simple chat response using `Llama-3.2-3B Instruct`_ model.

.. tip::
   Be patient. Allow code execution to complete in each cell before proceeding.

.. note::
   If you encounter an error with dependencies, select :guilabel:`Kernel -> Restart Kernel and Clear Outputs of All Cells`.
   Then restart from the top down in your Notebook and execute cells consecutively.

What you need
-------------

* `Learning`_ panel open
* This tutorial open

.. code-block:: ipython

   pip install --upgrade transformers

.. code-block:: ipython

   pip install jinja2==3.1.4 && pip show jinja2

.. note::
   Assure `Jinja2` version shows: Version: 3.1.0 or higher.

.. code-block:: ipython

   pip install accelerate==0.26.0

.. code-block:: ipython

   pip install sentencepiece

.. code-block:: ipython

   import huggingface_hub

.. code-block:: ipython

   my_token = 'EnterYourTokenHere'
   from huggingface_hub import login
   login(token = my_token)

.. code-block:: ipython

   #PROMPT 1 - Test knowledge of literature and translation

   import torch
   from transformers import pipeline

   model_id = "meta-llama/Llama-3.2-3B-Instruct"
   pipe = pipeline(
      "text-generation",
      model=model_id,
      torch_dtype=torch.bfloat16,
      device_map="auto",
      pad_token_id=50256,
   )

   messages = [
      {"role": "system", "content": "You possess a doctorate in European literature, and you specialize in Italian renaissance poetry"},
      {"role": "user", "content": "Show the first two lines of Dante's Inferno in the original language in which they were written. Then on a new line, translate them into English, and explain their significance. Format the response as markdown"},
      ]
   outputs = pipe(
      messages,
      max_new_tokens=512, # change to '256' if desired
   )

   response = str(outputs[0]["generated_text"][-1]["content"])
   print(response)

.. note::
   If you receive a `UserWarning: torch.utils...` during execution, ignore it and allow code to complete.

Next steps
**********

* Copy the `messages` array into `gradio-powered space with Llama-3.2-3B`_ on Hugging Face and compare the result.
* Learn more about `Llama 3.2 models from Hugging Face`_.
* Explore more free resources in `Learning`_.
* Visit `Optimum Habana Language Model Training`_ to test more models.

.. _Llama-3.2-3B Instruct: https://huggingface.co/meta-llama/Llama-3.2-3B-Instruct
.. _Llama 3.2 models from Hugging Face: https://huggingface.co/blog/llama32
.. _gradio-powered space with Llama-3.2-3B: https://huggingface.co/spaces/huggingface-projects/llama-3.2-3B-Instruct
.. _Learning: https://console.cloud.intel.com/learning
.. _Casual language modeling from Hugging Face: https://huggingface.co/docs/transformers/en/tasks/language_modeling
.. _Hugging Face Serving Private & Gated Models: https://huggingface.co/docs/text-generation-inference/en/basic_tutorials/gated_model_access
.. _Habana Prebuilt Containers: https://docs.habana.ai/en/latest/AWS_User_Guides/Habana_Deep_Learning_AMI.html#run-using-containers-on-habana-base-ami
.. _Terms of Use: https://llama.meta.com/llama3/license/
.. _console: https://console.cloud.intel.com
.. _Optimum Habana Language Model Training: https://github.com/huggingface/optimum-habana/tree/main/examples/language-modeling
.. _Optimum Habana Language Model Training Custom Files: https://github.com/huggingface/optimum-habana/tree/main/examples/language-modeling#custom-files
.. _Stanford Alpaca repo: https://github.com/tatsu-lab/stanford_alpaca
