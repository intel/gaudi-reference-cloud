.. _expose_app_tunnels:

Expose Local Apps with Tunnels
###############################

Quickly make available applications or tests running on |ITAC| via a public URL for demos, hackathons, or collaboration. Try the Cloudflare\* tool or the Gradio\* tool to expose local services to the public.

This tutorial shows how to set up tunnels for various workloads (e.g., web services, web apps, TGI servers, vLLM services). These examples assume your service is running locally on a specific port, e.g., ``http://localhost:8080``.

Prerequisites
*************

* Linux operating system (OS)
* Create an instance with :ref:`get_started`
* Use **One-Click connection** for your instance

**Optional**

* You may have an image generation service running locally
* You may need to quickly create a public URL to demo it or test its functionality

Cloudflare TryCloudflare
**************************

**Cloudflare TryCloudflare** lets you quickly expose local services to the Internet with no Cloudflare account required. While intended for testing or temporary use, Cloudflare also offers premium tunnels for production-grade needs.

Benefits of Cloudflare Tool
***************************

* **Quick Setup:** A single command is enough to expose your service
* **Secure:** Uses Cloudflare’s global network, protecting your IP address
* **Temporary Usage:** Ideal for demos or quick sharing during development


Set Up TryCloudflare
********************

#. Access your instance. See :ref:`connect_to_instance`.

#. Install `cloudflared`.

   .. code-block:: bash

      sudo apt update
      sudo apt install -y wget
      wget https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
      sudo dpkg -i cloudflared-linux-amd64.deb

#. Expose a Local Service. Replace port `8080` with your service port number.

   .. code-block:: bash

      cloudflared tunnel --url http://localhost:8080

#. Share the Public URL.

   After running the command, you’ll receive a public URL like:

   .. code-block:: console

      https://<random-subdomain>.trycloudflare.com

   .. note::

      This link is temporary and only works while the `cloudflared` process is running.

Example Workloads
*****************

* **Web Service**: Start your web service (e.g., Flask, FastAPI, Node.js) locally on port `8080` and use `cloudflared` to expose it.
* **TGI or VLLM Server**: Run your inference server locally, bound to a port, and use `cloudflared` to expose it for collaborators or clients to test.

For Production Use
==================

For permanent, production-grade tunnels with custom domains, better reliability, and SLAs, consider upgrading to `Cloudflare premium tunnel services`_.

Gradio Share Feature
********************

**Gradio** is widely used for creating interfaces for machine learning models, but its `share=True` option can also be repurposed to expose general-purpose applications.

Benefits of Gradio Tool
***********************

- **Ease of Use:** Designed for developers, with minimal configuration required
- **Temporary Links:** Ideal for demos with links valid for up to 72 hours
- **Integrated with Python:** Perfect for exposing Python-based services

Prerequisites
==============

* `pip`
* `python3`

Set Up Gradio Tunnels
*********************

#. Access your instance. See :ref:`connect_to_instance`.

#. Install Gradio.

   .. code-block:: bash

      pip install gradio

#. Expose a Local Service.

   Create a minimal Gradio app to expose your service (replace `8080` with your service’s port):

   .. code-block:: python

      import gradio as gr

      def expose_app():
         return "Service running on http://localhost:8080"

      demo = gr.Interface(fn=expose_app, inputs=[], outputs="text")
      demo.launch(share=True)

#. Share the Public URL.

   Running the script provides a public URL like:

   .. code-block:: bash

      https://<random-subdomain>.gradio.live

   .. note::
      This link is active for 72 hours.

Example Workloads
******************

- **Web Service**: Use Gradio to describe and share your web service running on `localhost:8080`.

- **Web App**: Expose your local React, Vue, or other frontend applications by sharing the port they’re hosted on.

- **Inference Servers**: Provide lightweight interaction interfaces for machine learning models.

For Production Use
==================

Gradio share links are temporary and not intended for production. For permanent hosting, consider deploying on platforms like `Hugging Face Spaces`_.

.. _Cloudflare premium tunnel services: https://www.cloudflare.com/products/tunnels/
.. _Hugging Face Spaces: https://huggingface.co/spaces
