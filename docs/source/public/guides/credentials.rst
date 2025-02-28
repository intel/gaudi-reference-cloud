:alttitle: Use credentials to access services
:category: credentials
:keywords: credentials, credentials as a service, authorization, token, API
:phrase: Generate a client secret and client ID to directly interact with a variety of services on the |ITAC|.
:rating: 0
:show_urls: false

.. _credentials:

Credentials as a Service
########################

Generate credentials on |ITAC| to directly interact with regional services. Follow this guide to generate a client secret and client ID.

.. note::
   You can create a maximum of two credentials.

Access Credentials
******************

#. Select the :guilabel:`User Profile` pull-down menu.

#. Click :guilabel:`Account Settings`.

#. View the :guilabel:`Credentials` tab.

Generate Secret
***************

#. In the :guilabel:`Credentials` tab, click :guilabel:`Generate Client Secret`.

#. Following onscreen instructions under :guilabel:`Secret name`, enter a name.

#. Click the button to :guilabel:`Generate Secret`.

#. Copy the client secret. Store the secret securely.

   .. warning::
      Be sure to copy the client secret in a secure location. This secret is only shown once.
      You will only have one chance to copy it; otherwise, you must regenerate it.


Delete Secret
*************

#. In the :guilabel:`Credentials` tab, under :guilabel:`Actions`, select :guilabel:`Delete`.

#. In the dialog, under :guilabel:`Name`, enter the client secret name.

   .. note::
      A check mark appears when your typed entry matches the name.

#. Click :guilabel:`Delete`.

Tips
****

* Credentials are mainly used as a token in API requests.
* To use user-credentials with the API, follow "generate secret" process (above) and include it in the Authorization Header as a bearer token.
* For multi-user accounts, the owner and member must both create their own unique credentials.

.. meta::
   :description: Generate credentials on Intel® Tiber™ AI Cloud to interact with regional services via API requests.
   :keywords: client secret, generate token, API request, API call

.. collectfieldnodes::
