:alttitle: Use Model as a Service API to build your AI solution
:category: software
:keywords: model as a service, Model API, Model endpoints, inference
:phrase: Use Model as a Service APIs. Build and deploy AI applications, automate workflows, enhance inference, and generate data to train or fine-tune models.
:rating: 0
:show_urls: false

.. _models:

Model as a Service API
######################

This guide demonstrates how to use the Model as a Service (MaaS) APIs on |ITAC|. Using the scripts included, you'll learn how to construct an API query, using one of two methods: Python; or :command:`curl`. First, complete `Prerequisites`_.

Optional: Jump to :ref:`get_request_workflows`.

Prerequisites
*************

* Complete :ref:`credentials` to generate:

  * `client_id`
  * `client_secret`

After completing these steps, continue below.

Supported Models
****************

View the latest models with Model as a Service (MaaS) APIs.

#. To view the latest models, choose a **GET request** method in :ref:`get_request_workflows`.

   .. tip::
      A GET request retrieves the currently available models in the API. For the most up-to-date model list, always **fetch models dynamically** rather than relying on static documentation.

#. Alternatively, refer to available models in the following table.

.. list-table:: Supported Models
   :widths: auto
   :header-rows: 1
   :class: table-tiber-theme

   * - Product Name
     - Product Id
     - Model Name

   * - maas-model-qwen-2.5-coder-32b
     - 97a75a6d-7eb4-40c9-894b-8e9f5924b555
     - "Qwen/Qwen2.5-Coder-32B-Instruct"

   * - maas-model-qwen-2.5-32b
     - 0a0bffe9-2f62-4ebb-8aee-118ede22b816
     - "Qwen/Qwen2.5-32B-Instruct"

   * - maas-model-llama-3.1-8b
     - ba5d2874-dc83-425e-af98-c810f11dad79
     - "meta-llama/Meta-Llama-3.1-8B-Instruct"

   * - maas-model-llama-3.1-70b
     - 8d728109-0fb2-46c7-a406-1113634d72ab
     - "meta-llama/Meta-Llama-3.1-70B-Instruct"

   * - maas-model-mistral-7b-v0.1
     - 269c3034-e6c7-4359-9e77-c3efedfaa778
     - "mistralai/Mistral-7B-Instruct-v0.1"


API Base URL
============

Use the API base :file:`url` for a GET request:

* :file:`https://us-region-2-sdk-api.cloud.intel.com/v1/maas/`

.. _get_request_workflows:

GET Request Workflows
*********************

Choose a GET request workflow.

* `Python Workflow`_
* `Curl Workflow`_

Python Workflow
===============

Follow these step-by-step instructions to manage authentication, make an API query, and make an inference prompt.

#. Create a python file. Paste these import statements and global variables at top.

   .. code-block:: python

      import json
      import time
      import requests
      from typing import Dict, List, Iterator

      # 1. Define global variables
      CLIENT_ID = "my_client_id"
      CLIENT_SECRET = "my_client_secret"
      CLOUD_ACCOUNT = "my_cloud_account"
      API_BASE_URL = "https://us-region-2-sdk-api.cloud.intel.com/v1/maas"
      AUTH_URL = "https://client-token.api.idcservice.net/oauth2/token"

#. Using variables generated from `Prerequisites`_, replace the values in the code for :file:`my_client_id`, :file:`my_client_secret`, and :file:`my_cloud_account`.

   .. tip:: Navigate to your profile icon in the console app to find the number for "my_cloud_account".

#. Add the function, :command:`get_auth_token` for authentication.

   .. code-block:: python

      import json
      import time
      import requests
      from typing import Dict, List, Iterator

      # 1. Define global variables
      CLIENT_ID = "my_client_id"
      CLIENT_SECRET = "my_client_secret"
      CLOUD_ACCOUNT = "my_cloud_account"
      API_BASE_URL = "https://us-region-2-sdk-api.cloud.intel.com/v1/maas"
      AUTH_URL = "https://client-token.api.idcservice.net/oauth2/token"


      # 2. Authenticate
      def get_auth_token(client_id: str = CLIENT_ID, client_secret: str = CLIENT_SECRET) -> str:
         '''Get authentication token for API access.'''
         response = requests.post(
            url=AUTH_URL,
            data='grant_type=client_credentials',
            headers={'Content-Type': 'application/x-www-form-urlencoded'},
            auth=(client_id, client_secret)
            )
         token_data = response.json()
         return f"{token_data['token_type']} {token_data['access_token']}"

#. Create the :command:`get_models` function, where you:

   * Invoke the :command:`get_auth_token` function

   * Pass two parameters for function (from `Prerequisites`_).

   .. code-block:: python

      # 3. Model Listing
      def get_models(client_id: str = CLIENT_ID, client_secret: str = CLIENT_SECRET) -> List[Dict]:
         '''Get list of all available models.'''
         headers = {'Authorization': get_auth_token(client_id, client_secret)}
         url = f'{API_BASE_URL}/models'
         response = requests.get(url, headers=headers)
         return response.json()['models']

#. Optional: If you wish to view the python function's response object, add a print statement with the :file:`response` after its line.

   The response should be similar to what follows.

   .. code-block:: console

      {
      "models":
         [
         {
            "model_name": "Qwen/Qwen2.5-Coder-32B-Instruct",
            "product_id": "97a75a6d-7eb4-40c9-894b-8e9f5924b555",
            "product_name": "maas-model-qwen-2.5-coder-32b"
         },
         {
            "model_name": "Qwen/Qwen2.5-32B-Instruct",
            "product_id": "0a0bffe9-2f62-4ebb-8aee-118ede22b816",
            "product_name": "maas-model-qwen-2.5-32b"
         },
         {
            "model_name": "mistralai/Mistral-7B-Instruct-v0.1",
            "product_id": "269c3034-e6c7-4359-9e77-c3efedfaa778",
            "product_name": "maas-model-mistral-7b-v0.1"
         },
         {
            "model_name": "meta-llama/Meta-Llama-3.1-8B-Instruct",
            "product_id": "ba5d2874-dc83-425e-af98-c810f11dad79",
            "product_name": "maas-model-llama-3.1-8b"
         },
         {
            "model_name": "meta-llama/Meta-Llama-3.1-70B-Instruct",
            "product_id": "8d728109-0fb2-46c7-a406-1113634d72ab",
            "product_name": "maas-model-llama-3.1-70b"
         }
         ]
      }

#. Add a function to manage text generation in the streaming response.

   .. code-block:: python

      # 4. Text Generation
      def generate_text_stream(
         prompt: str,
         model_info: Dict,
         client_id: str = CLIENT_ID,
         client_secret: str = CLIENT_SECRET,
         cloud_account_id: str = CLOUD_ACCOUNT,
         max_tokens: int = 250,
         temperature: float = 0.7
         ) -> Iterator[Dict]:

         '''Generate text with streaming response.'''
         payload = {
            "model": model_info['model_name'],
            "request": {
                  "prompt": prompt,
                  "params": {
                     "max_new_tokens": max_tokens,
                     "temperature": temperature
                  }
            },
            "cloudAccountId": cloud_account_id,
            "productName": model_info['product_name'],
            "productId": model_info['product_id']
         }
         headers = {
            'Authorization': get_auth_token(client_id, client_secret),
            'Content-Type': 'application/json'
         }
         response = requests.post(
            f'{API_BASE_URL}/generatestream',
            headers=headers,
            data=json.dumps(payload),
            stream=True
         )
         return (json.loads(line.decode('utf-8'))
                  for line in response.iter_lines() if line)


#. Finally, add the function, :command:`demonstrate_all_apis`.
   This function:

   * Prints the status of each process

   * Includes exception handing

   * Adds a prompt (or list of prompts).

   * Invokes the function :command:`generate_text_stream` with additional configuration

.. code-block:: python

   def demonstrate_all_apis():
      '''Demonstrates authentication, model listing, and text generation using Model as a Service APIs.'''
      print("=== Intel MaaS API Complete Demo ===\n")

      # Section 1: Authentication Demo
      print("1. Authentication Test")
      print("-" * 50)
      try:
         token = get_auth_token()
         print("Authentication successful")
         print(f"Token: {token[:50]}...")
      except Exception as e:
         print(f"Authentication failed: {str(e)}")
      print("\n")

      # Section 2: Model Listing Demo
      print("2. Available Models")
      print("-" * 50)
      try:
         models = get_models(CLIENT_ID, CLIENT_SECRET)
         print(f"Found {len(models)} available models:")
         for model in models:
               print(f"\nModel: {model['model_name']}")
               print(f"Product ID: {model['product_id']}")
               print(f"Product Name: {model['product_name']}")
      except Exception as e:
         print(f"Failed to list models: {str(e)}")
      print("\n")

      # Section 3: Text Generation Demo
      print("3. Text Generation Tests")
      print("-" * 50)
      test_prompts = [
         "Write a poem about programming, in four lines and two stanzas, which uses iambic pentameter in rhyming couplets."
         # "What are the key principles of a good AI application?"
      ]
      for model in models:
         print(f"\nTesting {model['model_name']}")
         print("-" * 30)
         prompt = test_prompts[0]
         print(f"Prompt: {prompt}\n")
         try:
               response = generate_text_stream(
                  prompt=prompt,
                  model_info=model,
                  max_tokens=100,
                  temperature=0.7
               )
               print("Response:")
               for chunk in response:
                  if 'result' in chunk:
                     token = chunk['result']['response']['token']['text']
                     print(token, end='', flush=True)
                     if 'details' in chunk['result']['response']:
                           details = chunk['result']['response']['details']
                           print(f"\n\nCompletion Details:")
                           print(f"- Tokens generated: {details['generated_tokens']}")
                           print(f"- Finish reason: {details['finish_reason']}")
                           if 'seed' in details:
                              print(f"- Seed: {details['seed']}")
               print("\n")
               time.sleep(2)
         except Exception as e:
               print(f"✗ Generation failed: {str(e)}")
      # Section 4: Parameter variation Demo
      print("\n4. Parameter Variation Test")
      print("-" * 50)
      model = models[0]
      prompt = "Write a short story about a robot."
      parameter_sets = [
         {"max_tokens": 50, "temperature": 0.2},
         {"max_tokens": 50, "temperature": 0.8},
         {"max_tokens": 200, "temperature": 0.5}
      ]

      for params in parameter_sets:
         print(f"\nTesting with parameters:")
         print(f"- Max tokens: {params['max_tokens']}")
         print(f"- Temperature: {params['temperature']}")
         try:
               response = generate_text_stream(
                  prompt=prompt,
                  model_info=model,
                  max_tokens=params['max_tokens'],
                  temperature=params['temperature']
               )
               print("\nResponse:")
               for chunk in response:
                  if 'result' in chunk:
                     token = chunk['result']['response']['token']['text']
                     print(token, end='', flush=True)
               print("\n")
               time.sleep(2)
         except Exception as e:
               print(f"✗ Generation failed: {str(e)}")

   if __name__ == "__main__":
      demonstrate_all_apis()

You have successfully completed the Python method. See also the `Complete Python Workflow Script`_.

Modify and Monitor Results
---------------------------

If you like, modify any of the following values and re-run the script. Monitor streaming output to understand the impact of your change.

* Change the value of :command:`temperature` from ``0.7`` to ``0.9``. Do you notice a difference in the response?

* Add **your prompts** after the variable, :command:`test_prompts`. Uncomment the second question.

* Try simplifying the syntax of **your prompt**. Use the most efficient syntax and common English words.

* Observe how some models (with different model parameter size) show a better quality of response, or no response.

.. tip::
   The python script in its entirety is also pasted below for your convenience.

Curl Workflow
=============

For the :command:`curl` API query, use the following command. Replace the :file:`token` with your own.

.. code-block:: bash

   curl --location 'https://us-region-2-sdk-api.cloud.intel.com/v1/maas/models' \
   --header 'Content-Type:  application/json' \
   --header 'Authorization: Bearer ${token}'

Curl GetModels Response
-----------------------

Try the :command:`curl` command. It should be similar to the following.

Replace the following values with your own:

* :file:`cloudAccountId`
* :file:`productName`
* :file:`productId`
* :file:`model`

.. code-block:: console

   curl --location 'https://us-region-2-sdk-api.cloud.intel.com/v1/maas/generatestream' \
   --header 'Content-Type:  application/json' \
   --header 'Authorization: Bearer ${token}'
   --data '{
      "cloudAccountId": "cloudAccountId",
      "request": {
         "params": {
               "maxNewTokens": 100,
               "temperature": 0.5
         },
         "prompt": "Tell me a joke"
      },
      "productName": "maas-model-mistral-7b-v0.1",
      "productId": "269c3034-e6c7-4359-9e77-c3efedfaa778",
      "model": "mistralai/Mistral-7B-Instruct-v0.1"
   }'

As shown above, be sure to properly include ``productName`` and ``productId``. They are required within your request payload.

.. caution::

   All passable values shown in the request payload must match those shown in `Supported Models`_.

The response is streamed. Each stream provides data of a certain token and appears like so:

.. code-block:: console

   {
      "result": {
         "response": {
               "token": {
                  "id": 1061,
                  "text": "Data",
                  "logprob": -0.11687851,
                  "special": false
               },
               "top_tokens": [],
               "requestID": "8764a8c3-deb2-4f1d-a60a-a47043b0e9f5"
         }
      }
   }

The last chunk has additional data and appears like so:

.. code-block:: console

   {
      "result": {
         "response": {
               "token": {
                  "id": 330,
                  "text": " \"",
                  "logprob": -4.5329647,
                  "special": false
               },
               "top_tokens": [],
               "generated_text": "Data and information are often used interchangeably, but they have distinct meanings.\n\n**Data** refers to a set of facts, numbers, or observations that are collected, recorded, or stored in a way that can be analyzed or used for reference. Data can be numbers, words, images, or any other form of content. It's the raw material that can be used to inform decisions, answer questions, or solve problems.\n\nFor example: A list of customer names, ages, and purchase history is a collection of data.\n\n**Information**, on the other hand, is a description, explanation, or interpretation of data that provides meaning or context. Information is the result of processing, analyzing, or interpreting data. It's the output of data that can be used to inform, educate, or influence decisions.\n\nTo illustrate the difference, consider this example:\n\n* Data: A list of exam scores (e.g., 90, 80, 95, 70)\n* Information: \"The average exam score is 85, indicating that the students are performing above average.\"\n\nIn this example, the list of scores is data, while the interpretation of those scores (average score and its meaning) is information.\n\nTo summarize:\n\n* Data is the \"",
               "details": {
                  "finish_reason": "FINISH_REASON_LENGTH",
                  "generated_tokens": 250,
                  "seed": "13414724236876468656"
               },
               "requestID": "e8fa7572-9ea8-4acc-816e-2c24ec47cd8a"
         }
      }
   }

You have successfully completed the curl method.

Complete Python Workflow Script
===============================

.. code-block:: python

   import json
   import time
   import requests
   from typing import Dict, List, Iterator

   # 1. Define global variables
   CLIENT_ID = "my_client_id"
   CLIENT_SECRET = "my_client_secret"
   CLOUD_ACCOUNT = "my_cloud_account"
   API_BASE_URL = "https://us-region-2-sdk-api.cloud.intel.com/v1/maas"
   AUTH_URL = "https://client-token.api.idcservice.net/oauth2/token"

   # 2. Authentication
   def get_auth_token(client_id: str = CLIENT_ID, client_secret: str = CLIENT_SECRET) -> str:
      '''Get authentication token for API access.'''
      response = requests.post(
         url=AUTH_URL,
         data='grant_type=client_credentials',
         headers={'Content-Type': 'application/x-www-form-urlencoded'},
         auth=(client_id, client_secret)
      )
      token_data = response.json()
      return f"{token_data['token_type']} {token_data['access_token']}"

   # 3. Model Listing
   def get_models(client_id: str = CLIENT_ID, client_secret: str = CLIENT_SECRET) -> List[Dict]:
      '''Get list of all available models.'''
      headers = {'Authorization': get_auth_token(client_id, client_secret)}
      url = f'{API_BASE_URL}/models'
      response = requests.get(url, headers=headers)
      return response.json()['models']

   # 4. Text Generation
   def generate_text_stream(
      prompt: str,
      model_info: Dict,
      client_id: str = CLIENT_ID,
      client_secret: str = CLIENT_SECRET,
      cloud_account_id: str = CLOUD_ACCOUNT,
      max_tokens: int = 250,
      temperature: float = 0.7
   ) -> Iterator[Dict]:
      '''Generate text with streaming response.'''
      payload = {
         "model": model_info['model_name'],
         "request": {
               "prompt": prompt,
               "params": {
                  "max_new_tokens": max_tokens,
                  "temperature": temperature
               }
         },
         "cloudAccountId": cloud_account_id,
         "productName": model_info['product_name'],
         "productId": model_info['product_id']
      }
      headers = {
         'Authorization': get_auth_token(client_id, client_secret),
         'Content-Type': 'application/json'
      }
      response = requests.post(
         f'{API_BASE_URL}/generatestream',
         headers=headers,
         data=json.dumps(payload),
         stream=True
      )
      return (json.loads(line.decode('utf-8'))
               for line in response.iter_lines() if line)

   def demonstrate_all_apis():
      '''Run comprehensive demonstration of all API capabilities.'''
      print("=== Intel MaaS API Complete Demo ===\n")

      # Section 1: Authentication Demo
      print("1. Authentication Test")
      print("-" * 50)
      try:
         token = get_auth_token()
         print("Authentication successful")
         print(f"Token: {token[:50]}...")
      except Exception as e:
         print(f"Authentication failed: {str(e)}")
      print("\n")

      # Section 2: Model Listing Demo
      print("2. Available Models")
      print("-" * 50)
      try:
         models = get_models(CLIENT_ID, CLIENT_SECRET)
         print(f"Found {len(models)} available models:")
         for model in models:
               print(f"\nModel: {model['model_name']}")
               print(f"Product ID: {model['product_id']}")
               print(f"Product Name: {model['product_name']}")
      except Exception as e:
         print(f"Failed to list models: {str(e)}")
      print("\n")

      # Section 3: Text Generation Demo
      print("3. Text Generation Tests")
      print("-" * 50)
      test_prompts = [
         "Write a poem about programming, in four lines and two stanzas, which uses iambic pentameter in rhyming couplets."
         # "What are the key principles of a good AI application?"
      ]
      for model in models:
         print(f"\nTesting {model['model_name']}")
         print("-" * 30)
         prompt = test_prompts[0]
         print(f"Prompt: {prompt}\n")
         try:
               response = generate_text_stream(
                  prompt=prompt,
                  model_info=model,
                  max_tokens=100,
                  temperature=0.9
               )
               print("Response:")
               for chunk in response:
                  if 'result' in chunk:
                     token = chunk['result']['response']['token']['text']
                     print(token, end='', flush=True)
                     if 'details' in chunk['result']['response']:
                           details = chunk['result']['response']['details']
                           print(f"\n\nCompletion Details:")
                           print(f"- Tokens generated: {details['generated_tokens']}")
                           print(f"- Finish reason: {details['finish_reason']}")
                           if 'seed' in details:
                              print(f"- Seed: {details['seed']}")
               print("\n")
               time.sleep(2)
         except Exception as e:
               print(f"✗ Generation failed: {str(e)}")
      # Section 4: Parameter variation Demo
      print("\n4. Parameter Variation Test")
      print("-" * 50)
      model = models[0]
      prompt = "Write a short story about a robot."
      parameter_sets = [
         {"max_tokens": 50, "temperature": 0.2},
         {"max_tokens": 50, "temperature": 0.8},
         {"max_tokens": 200, "temperature": 0.5}
      ]

      for params in parameter_sets:
         print(f"\nTesting with parameters:")
         print(f"- Max tokens: {params['max_tokens']}")
         print(f"- Temperature: {params['temperature']}")
         try:
               response = generate_text_stream(
                  prompt=prompt,
                  model_info=model,
                  max_tokens=params['max_tokens'],
                  temperature=params['temperature']
               )
               print("\nResponse:")
               for chunk in response:
                  if 'result' in chunk:
                     token = chunk['result']['response']['token']['text']
                     print(token, end='', flush=True)
               print("\n")
               time.sleep(2)
         except Exception as e:
               print(f"✗ Generation failed: {str(e)}")

   if __name__ == "__main__":
      demonstrate_all_apis()

.. meta::
   :description: Use Model as a Service APIs on |ITAC|. Build and deploy AI applications, automate workflows, and train or fine-tune models.
   :keywords: model as a service, Model API, Model endpoints, inference

.. collectfieldnodes::