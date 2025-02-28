.. _staas_object:

Object Storage
###############

Model training and inference often require storing large amounts of *unstructured data*.
|ITAC| offers object storage, with a choice of CLI clients, to manage storage buckets.

Prerequisites
*************

* :ref:`get_started`
* :ref:`manage_instance`
* AWS\* S3 CLI client **version 2.13** or higher
* A running instance

.. important::
   This guide assumes you've already created an instance and that you can access it via SSH.
   You can only access private bucket storage from any |ITAC| compute platform, whether virtual or bare metal machines, or |INTC| Kubernetes Services.

Create bucket
*************

#. In the left side menu, select :command:`Storage > Object Storage`.

#. Click :guilabel:`Create bucket`.

#. Enter a bucket name in the  **Name** field. Optional: Enter a **Description**.

#. Click :guilabel:`Enable versioning` if desired.

   .. note::
      Versioning ensures that data recovery is available for a limited time.
      This feature provides data recovery in case of application failure
      or unintentional deletion.

Create bucket user
******************

Follow these steps to enable storage bucket access for a principal.

General rules
=============

* You must create a bucket **principal** to enable access.
* You may create other principals for the same bucket.
* Principals are mapped to buckets through a policy.

.. note::
   "Principals" means users who can consume data in object storage buckets.

#. Click :guilabel:`Manage principals and permissions`.

#. Click :guilabel:`Create principal`.

   Options: You can apply permissions for *all buckets* or *per bucket*.
   To do so, select menu items under :guilabel:`Allowed actions` and :guilabel:`Allow policies`.

#. Select the :guilabel:`Credentials` tab, below the new principal.
   These credentials are used for logging into a bucket.

Later in this workflow, we refer to the :guilabel:`Credentials` tab for login.

Install CLI Client
*******************

For first-time use, install the CLI client software in your instance to access bucket storage.

This guide explains how to install the AWS/* S3 CLI, which is one of many options.
For detailed commands, view the `AWS CLI Command Reference`_.

.. tip::
   You choose which CLI Client you wish to use.

Optional CLI clients
=====================

You can install boto3 by following `Boto3 Documentation`_.

You can install the MINIO Client\* software to connect to storage buckets.
Visit `MINIO Documentation`_ to learn more.

Install AWS S3 CLI
-------------------

#. From the console, open an instance where you want to access object storage.
   If you have not created an instance, complete steps in :ref:`get_started` and return.

#. Follow the onscreen instructions and SSH into your instance.

#. Install the AWS\* S3 CLI client, **version 2.13 or higher**. Going forward, we call this the CLI client.

   .. code-block:: bash

      sudo apt install unzip
      curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
      unzip awscliv2.zip
      sudo ./aws/install

#. Verify the CLI client was properly installed.

   .. code-block:: bash

      aws --version

#. Confirm your standard output is similar.

   .. code-block:: console

      aws-cli/2.22.18 Python/3.12.6 Linux/5.15.0-1051-kvm exe/x86_64.ubuntu.22

   .. note::
      Your version may differ. These steps confirm proper installation only.

Installation complete.

Access Storage Bucket
**********************

During each login session, you must enter credentials to access a storage bucket.

.. note::

   Unless you log out of the current principal account, session history is preserved.
   If you log in as different principal, you're required to use new credentials.

Credentials Login
=================

Follow this instruction to generate a password.

#. In the left side menu, select :command:`Storage > Object Storage`.

#. In the :guilabel:`Object Storage` tab, click :guilabel:`Manage Principals and Permissions`.

#. View the principals table and click on the principal name.

#. Click in the :guilabel:`Principals` tab.

#. Click :guilabel:`Generate password`.

   a. Optional: You may skip to `Option 2 - Create .env file`_ to use an alternate credential method.
      Otherwise, continue.

   .. note::
      The **AccessKey** and **SecretKey** are similar to a **Username** and **Password**.

#. Enter command to log in. You'll be prompted for credentials.

   .. code-block:: bash

      aws configure

#. Copy and paste the values *from the console to the CLI client* from previous steps:

   *  **AWS Access Key ID [None]**     - **AccessKey**
   *  **AWS Secret Access Key [None]** - **SecretKey**

#. For :guilabel:`Default region name`, press :kbd:`<Enter>` to accept default.

#. For :guilabel:`Default output format`, type "json", and press :kbd:`<Enter>`.

Configure Environment Variables
********************************

Choose **one option** below. Configuring environment variables simplifies storage bucket queries.

By using an option below, you don't need to add the flag :command:`--endpoint-url https://private.endpoint.url`
See also `Environmental variables for AWS CLI`_.

Option 1 - Export the URL Endpoint for this session
===================================================

#. In the console, navigate to :command:`Storage Bucket > Details`.

#. Find the :guilabel:`Private Endpoint URL`.

   a. Click the copy icon after :guilabel:`Private Endpoint URL`.

#. Run the following command, replacing "your_endpoint_url" with endpoint you copied in the previous step.

   .. code-block:: bash

      export AWS_ENDPOINT_URL='your_endpoint_url'

#. Example: List the bucket.

   .. code-block:: bash

      aws s3 ls

#. Skip to `S3 Bucket Commands`_.

Option 2 - Create .env file
===========================

Create an :file:`.env` file, with settings, to simplify queries.

.. caution::
   If you program the :file:`.env` file to persist (e.g., bash script), note -
   with each AWS login, you must update the AWS_SECRET_ACCESS_KEY.

#. Navigate to the root of your instance.

#. Create an :file:`.env` file.

   .. code-block:: bash

      touch .env

#. Add the following lines to your:file:`.env` file.

   .. code-block:: bash

      sudo vi .env

   .. code-block:: bash

      export AWS_ACCESS_KEY_ID='your_access_key_id'
      export AWS_SECRET_ACCESS_KEY='your_secret_access_key'
      export AWS_ENDPOINT_URL='your_endpoint_url'

   .. caution::
      Enclose all credentials within single quotes.

#. Replace `your_access_key_id`, `your_secret_access_key`, and `your_endpoint_url` using the next steps.

#. Find the :guilabel:`Private Endpoint URL`.

   a. Click the copy icon after :guilabel:`Private Endpoint URL`.

   #. Paste the value for :command:`AWS_ENDPOINT_URL` in :file:`.env`.

#. From :guilabel:`Storage Buckets`, click :guilabel:`Manage buckets principals and permissions`.

#. Click :guilabel:`Credentials`.

   a. Click :guilabel:`Generate password`.

   #. Paste values for **AccessKey** and **SecretKey**, :command:`AWS_ACCESS_KEY_ID`, :command:`AWS_SECRET_ACCESS_KEY` respectively in :file:`.env`.

#. Load environment variables from .env file

   .. code-block:: bash

      source .env

#. Query storage bucket data with `S3 Bucket Commands`_.

#. Example: List the bucket.

   .. code-block:: bash

      aws s3 ls

S3 Bucket Commands
******************

Use the examples for `AWS CLI Command Reference`_ to construct a command.
Using these commands assumes you've already configured your CLI Client to use an option above.

Delete Bucket
**************

#. In the console, navigate to :command:`Storage > Object Storage`.

#. In the :guilabel:`Object Storage` table, find :guilabel:`Bucket`.

.. important::
   Recommended: First remove the principals associated with the bucket.

#. In the :guilabel:`Actions` column, select :guilabel:`Delete`.

#. In the dialog, select :guilabel:`Delete` again to confirm your request.

   a. Assure that the bucket is empty before deletion.

Update user policy
******************

Follow these steps to change permmissions for principals.

Edit or Delete Bucket User
===========================

#. Visit :command:`Storage > Object Storage`.

#. Click :guilabel:`Manage principals and permissions`.

#. In the :guilabel:`Principals` table, find the :guilabel:`Actions` column at right.

#. To Edit, continue. Or skip to next step.

   a. Select :guilabel:`Edit` to modify permissions.

   #. Click :guilabel:`Edit`

   #. Select permissions that apply.

   #. Select :guilabel:`Save` to apply changes.

#. To delete, select :guilabel:`Delete`.

#. In the dialog, select :guilabel:`Delete` again to confirm your request.

Apply Lifecycle Rules
**********************

#. Select :command:`Bucket Name > Lifecycle Rules`.

#. Click :guilabel:`Create rule`.

   Next, the :guilabel:`Add Lifecycle Rule` workflow appears.

   .. note::

      Choose only one, :guilabel:`Delete Marker` or :guilabel:`Expiry Days`.
      Selecting the first means :guilabel:`Expiry Day` and :guilabel:`Non current expiry days` are disabled.
      The :guilabel:`Delete Marker` is related to versioning.

#. Enter a name, following the onscreen instructions.

#. Enter a prefix.

   In our example, we use only the :file:`/cache` directory.

   * If you don't enter a prefix, the rule applies to all items in the bucket
   * If you do enter a prefix, the rule applies to a specfic directory

#. For non current expiry days, you may leave blank (or enter "0")
   if you don't use versioning. See also: `NonCurrentVersionExpiration Docs`_.

#. To edit/delete a Lifecycle Rule, return to :command:`Bucket name > Lifecycle Rules`
   Then click :guilabel:`Edit` or :guilabel:`Delete` and follow the onscreen instructions.

Network Security Group
**********************

To view this feature, navigate to the :command:`Bucket Name > Details`. Network security is enforced using Source IP Filtering, which restricts user access to a bucket from a specific IP using a subnet mask.

.. _MINIO Documentation: https://min.io/docs/minio/linux/reference/minio-mc.html
.. _Boto3 Documentation: https://boto3.amazonaws.com/v1/documentation/api/latest/index.html
.. _NonCurrentVersionExpiration Docs: https://docs.aws.amazon.com/AmazonS3/latest/API/API_control_NoncurrentVersionExpiration.html
.. _AWS CLI Command Reference: https://awscli.amazonaws.com/v2/documentation/api/latest/reference/s3/index.html
.. _Environmental variables for AWS CLI: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html
