.. _vmaas_default_machine_image_update_procedure:

VMaaS Default Machine Image Update Procedure
############################################

This document describes the procedure to create and update the VMaaS default machine image with latest **ubuntu 22.04** release on *dev-jf*, *staging*, and production environment.

   .. note::
      This document may serve as a reference for additional types of images in other environments

Prerequisites
*******************

AGS Roles

* DEVTOOLS - ARTIFACTORY - IDC_INFLOW - AF01P-IGK - PROJECT VIEWER
* IDC - DEVELOPER
* Full Admin Access to AWS Account 171925926474

Creating New Machine Image (Only if it has not been already created by the inflow process)
*********************************************************************************************

#. Create a new branch in the `Software Inflow Repo`_. 
   Edit the url and checksum field in *records/base/ubuntu-22.04-server-cloudimg-amd64-disk-kvm.yaml* file with the latest ubuntu release image available at `Ubuntu Releases`_.
   Edit the ``records/recipe/tenant/vmaas_base.yaml`` with other fields and components if required
   
   .. note::
      Adding new components to the recipe can be done by either creating new components or reusing the exisiting one under *records/component* directory.

#. Raise a Pull Request with the above changes and wait for the *Image Builder Pipeline* Jenkins Job to finish. On successful completion a qcow2 image along with checksum files
   should be generated as shown below. Refer `Machine Image Inflow PR`_.
  
   .. figure:: ../../_figures/guides/vmaas_update_machine_image_01.png
      :alt: Image Builder Pipeline
   
   .. note::
      Sometimes the overall status shows a failure due to a bug in the pipeline execution but as you can see from the above snippet, every individual component is actually passing
      therefore you can safely use the generated artifacts and checksum files in subsequent steps. But in case the pipeline fails to build any component due an error with recipe, base image
      or component, make sure to fix the error so that all the components are generated correctly.
   
#. Download the qcow2 and checksum files on your local workstation. This can be done using *JFrog CLI* using below steps

   * Intsall the latest version of JFrog CLI from `JFrog Installation Page`_.
   * Configure JFrog
     
     .. code-block:: bash
        
        jf c add
     
     The above command will ask for various details of the Artifactory server. Use the following to set it up
        
     - url: https://internal-placeholder.com
     - user: your_intel_sso_user
     - access token: can be generated by clicking on the *Set me up* button in the top right corner of `IDC Inflow JFrog Repo`_.


   * Download qcow2 image along with checksum and log files 
     
     .. code-block:: bash

        jf rt dl --skip-checksum idc_inflow-igk-local/PR/PR-195/1/builds/recipe.tenant.vmaas_base.yaml/

     .. note::
        
        In the above command the last argument **idc_inflow-igk-local/PR/PR-195/1/builds/recipe.tenant.vmaas_base.yaml/** can be obtained by copying `QCOW2 URL Link`_
        generated for qcow2 file after completion of Image Builder Pipeline and stripping off the initial part of the URL until https://af01p-igk.com/artifactory/ and
        also removing everything after the last forward slash (/) in the URL.

#. Rename the downloaded qcow2 and checksum file. 
      
   .. note::
      Renaming also requires updating the names in the checksum files. This can be done a bit more easily with the script *misc/rename_image.sh* in the `Software Inflow Repo`_. 

   .. code-block:: bash

      mkdir -p new_qcow2_path/ubuntu-2204-jammy-v20240308
      misc/rename_image.sh PR/PR-195/1/builds/recipe.tenant.vmaas_base.yaml/ci_image.qcow2 new_qcow2_path/ubuntu-2204-jammy-v20240308
      mv new_qcow2_path/ubuntu-2204-jammy-v20240308/ci_image.qcow2 new_qcow2_path/ubuntu-2204-jammy-v20240308/ubuntu-2204-jammy-v20240308.qcow2

#. Create a http server on your workstation so that harvester can download the qcow2 file. This will be useful while testing the new image on *dev-jf*.
   
   .. code-block:: bash

      sudo apt update
      sudo apt install apache2

   By default, Apache listens on port 80 for HTTP traffic. If you want to use some other port say 8000, you need to change the Apache configuration
   to listen on that port.
   
   * Edit the Apache ports configuration file ``/etc/apache2/ports.conf`` and add or modify the ``Listen`` directive to listen on port 8000.
   * Edit the default virtual host configuration file ``/etc/apache2/sites-enabled/000-default.conf`` with the newer port ``VirtualHost \*:8000``.
   * Set iptables rules to port forward from port 8000

     .. code-block:: bash

        sudo iptables -I INPUT -p tcp -m tcp --dport 8000 -j ACCEPT

     .. note:: 
        If using VS Code to a remote SSH server, configure a port forward from port 8000 to localhost:8000

#. Copy the renamed qcow2 and checksum files to ``/var/www/html``
   
   .. code-block:: bash
      
      sudo cp new_qcow2_path/ubuntu-2204-jammy-v20240308.* /var/www/html
      sudo cp new_qcow2_path/ubuntu-2204-jammy-v20240308/ubuntu-2204-jammy-v20240308.qcow2 /var/www/html


New Image on Dev
*****************

#. Create a new branch in the `IDC Monorepo`_.

#. Create a VirtualMachineImage file under ``go/pkg/compute_api_server/testdata/VirtualMachineImage`` with same name as that of the new ubuntu release. The VirtualMachineImage name must be fewer than 32 characters.
   refer `VirtualMachineImage YAML Ubuntu 20230122`_.

   .. note::

      * make sure the fields metadata.harvesterhci.io/imageDisplayName , metadata.name and spec.displayName are set correctly in accordance with the new ubuntu release.
      * spec.checksum points to the sha512 checksum for qcow2 image which can be obtained from file ``/var/www/html/ubuntu-2204-jammy-v20240308.sha512sum``
      * spec.url should point to your workstation server created above as ``http://${USER}internal-placeholder.com:8000/ubuntu-2204-jammy-v20240308.qcow2``

#. Create a MachineImage with same name as of VirtualMachineImage under ``go/pkg/compute_api_server/testdata/MachineImage`` 
   refer `MachineImage YAML Ubuntu 20230122`_.

   .. note::
      
      * make sure the fields metadata.name, spec.displayName are set correctly in accordance with the new ubuntu release.
      * The values for spec.md5sum, spec.sha256sum, spec.sha512sum can be obtained from the files */var/www/html/ubuntu-2204-jammy-v20240308.\** .
   
#. Create a tar file containing all the required VirtualMachineImage YAML files.
   
   .. code-block:: bash

      tar -C go/pkg/compute_api_server/testdata/VirtualMachineImage -cz . | base64

#. SSH into a node in the dev-jf harvester cluster

   .. code-block:: bash

      ssh rancher@10.165.57.245
      rancher@iaas-node-003:~> sudo -i

   When prompted for the password use the *ssh* secrets from `Harvester1 SSH Secret`_ .

#. Extract VirtualMachineImage YAML files onto the Harvester node. You should copy/paste the base64 output from the prior step into a file say *virtualMachineImage_base64.txt* 
   on the harvester node

   .. code-block:: bash

      mkdir -p VirtualMachineImage
      base64 -d VirtualMachineImage_base64.txt | tar -C VirtualMachineImage -xvz

#. Apply VirtualMachineImage Kubernetes resource on the harvester.

   .. code-block:: bash

      kubectl apply -f VirtualMachineImage

#. Confirm that the VirtualMachineImage was imported and downloaded completely on harvester.

   .. code-block:: bash

      kubectl get VirtualMachineImage -o yaml
   
   "status.conditions[type=Imported].status" should be True.
    
   .. figure:: ../../_figures/guides/vmaas_update_machine_image_02.png
      :alt: VirtualMachineImage Download Completed

   .. note::

      In case the image downloading gets stuck at 99% as shown below, check the checksum (sha512) in the VirtualMachineImage.yaml. This usually happens because of wrong checksum
      in the VirtualMachineImage file

   .. figure:: ../../_figures/guides/vmaas_update_machine_image_03.png
      :alt: VirtualMachineImage Download Stuck

#. Create a Pull Request and wait for Jenkins pipeline to finish successfully.


#. Populate the new machine image on the dev-jf

   .. code-block:: bash

      export IDC_ENV=dev-jf
      make secrets
      HELMFILE_OPTS="apply --selector name=us-dev-1-populate-machine-image" make run-helmfile-only
    
    
   If the machine image is populated correctly , the ``compute-db`` pod will have the new image. You can verfiy the same by exec into the ``compute-db container``

   .. figure:: ../../_figures/guides/vmaas_update_machine_image_04.png
      :alt: New Machine Image In compute-db 

#. Create an instance using new image via `Dev Jf Console`_.

   Confirm that IDC console defaults to the new MachineImage but allows the old MachineImage to be selected as shown below

   .. figure:: ../../_figures/guides/vmaas_update_machine_image_05.png
      :alt: Launch Instance with New image

   .. note::
      In case you do not have sufficient credits to reserve  an instance , you can create a $100 coupon for testing and redeem it with the following steps

   .. tip::
      Use Postman to execute the APIs

   * Generate an admin bearer token for dev-jf
      
     API: ``https://internal-placeholder.com/token?email=fintest3@intel.com&groups=IDC.Admin``

     .. figure:: ../../_figures/guides/vmaas_update_machine_image_06.png
        :alt: Dev jf admin bearer token
      
   * Create a $100 coupon
     
     - Copy the genereated token in the prior step into ``Authorization --> Bearer Token``
     - API: ``https://internal-placeholder.com/v1/billing/coupons``
     - Body:
       
       .. code-block:: bash

          {
            "amount":100,
            "creator":"Test User",
            "numUses":6,
            "isStandard": false
         }
     
     .. figure:: ../../_figures/guides/vmaas_update_machine_image_07.png
        :alt: Create coupon
       
   * Redeem the coupon on `Dev Jf Console`_.

     .. figure:: ../../_figures/guides/vmaas_update_machine_image_08.png
        :alt: Redeem coupon

   Confirm that the newly created instance is in *Ready* State

   .. figure:: ../../_figures/guides/vmaas_update_machine_image_09.png
      :alt: Instance Ready

#. SSH into the instance and make sure the ``sudo apt update`` works correctly

   .. code-block:: bash

      ssh -J guest-dev-jf@10.165.62.252 ubuntu@172.16.3.37
      
      ubuntu@my-instance-1:~$ sudo apt update


New Image on Staging/Prod
***************************

Harvester uses VirtualMachineImage custom resources to copy machine image qcow files from an S3 bucket to the Harvester cluster

#. Ensure that the machine image qcow2 file and checksum files has been uploaded to the `Machine Image S3 Bucket`_ *vmaas/images* folder.
   This can be done using aws cli from you workstation using below steps. For detailed info on using AWS refer `AWS IDC Usage`_.

   * Install AWS CLI
     
     .. code-block:: bash

        ARCH=amd64
        PLATFORM=$(uname -s)_$ARCH

        cd ~/Downloads
        curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
        unzip awscliv2.zip
        sudo ./aws/install

   * Use AWS CLI
     
     .. code-block:: bash

        aws configure sso

        SSO session name (Recommended): intel-sso
        Attempting to automatically open the SSO authorization page in your default browser.
        If the browser does not open or you wish to use a different device to authorize this request, open the following URL:

        https://device.sso.us-west-2.amazonaws.com/

        Then enter the code:

        ....

        There are 3 AWS accounts available to you.
        Using the account ID 171925926474
        The only role available to you is: AWSAdministratorAccess
        Using the role name "AWSAdministratorAccess"
        CLI default client Region [us-west-2]:
        CLI default output format [None]:
        CLI profile name [AWSAdministratorAccess-171925926474]:

        To use this profile, specify the profile name using --profile, as shown:

        aws s3 ls --profile AWSAdministratorAccess-171925926474

   * Push the qcow2 and checksum files to the S3 bucket

     .. tip:: 
        Before uploading the qcow2 and checksum files to S3, make sure you have executed step 6 of ``Creating a New Machine Image`` section. 
      
     .. code-block:: bash

        aws s3 cp /var/www/html/ubuntu-2204-jammy-v20240308.sha512sum s3://catalog-fs-dev/vmaas/images/ubuntu-2204-jammy-v20240308.sha512sum --profile AWSAdministratorAccess-171925926474
        aws s3 cp /var/www/html/ubuntu-2204-jammy-v20240308.sha256sum s3://catalog-fs-dev/vmaas/images/ubuntu-2204-jammy-v20240308.sha256sum --profile AWSAdministratorAccess-171925926474
        aws s3 cp /var/www/html/ubuntu-2204-jammy-v20240308.md5sum s3://catalog-fs-dev/vmaas/images/ubuntu-2204-jammy-v20240308.md5sum --profile AWSAdministratorAccess-171925926474
        aws s3 cp /var/www/html/ubuntu-2204-jammy-v20240308.qcow2 s3://catalog-fs-dev/vmaas/images/ubuntu-2204-jammy-v20240308.qcow2 --profile AWSAdministratorAccess-171925926474

      
     .. note::
        here *ubuntu-2204-jammy-v20240308* refers to the ubuntu release to which machine image is being updated to
       

   * Confirm that the qcow2 image and checksum files have been uploaded to *vmaas/images* folder of the `Machine Image S3 Bucket`_.

     .. figure:: ../../_figures/guides/vmaas_update_machine_image_10.png
        :alt: S3 Bucket Upload

#. Create a VirtualMachineImage file under ``build/environments/<env>/<region>/VirtualMachineImage``. The host IP address in the url field should point to the regional NGINX caching proxy server .
   This server proxies requests to the above S3 bucket.
   
   .. note::
      use env=prod for production and env=staging for staging

   .. tip:: 
      In order to create VirtualMachineImage file , you can simply copy the VirtualMachineImage file created at ``go/pkg/compute_api_server/testdata/VirtualMachineImage`` and just add
      *spec.retry: 3* and update *spec.url* with NGINX caching server


#. Create a MachineImage file with same name as that of VirtualMachineImage file under ``build/environments/<env>/MachineImage``. 
   
   .. tip::
      In order to create MachineImage file, you can simply copy the MachineImage file under ``go/pkg/compute_api_server/testdata/MachineImage``
   
   .. code-block:: bash

      cp go/pkg/compute_api_server/testdata/MachineImage/ubuntu-2204-jammy-v20240308.yaml build/environments/<env>/MachineImage

#. Create a tar file containing all required VirtualMachineImage YAML files.

   .. code-block:: bash

      tar -C build/environments/<env>/<region>/VirtualMachineImage -cz . | base64

#. SSH into a node in the harvester1 cluster refer `Connect Staging Harvester1 Cluster`_.

   * Sudo as root and test kubectl

     .. code-block:: bash

        rancher@pdx05-c01-azvm003:~>
        sudo -i

        pdx05-c01-azvm003:~ #
        kubectl get nodes

#. Extract VirtualMachineImage YAML files onto the Harvester node. You should copy/paste the base64 output from the prior step into a file say *virtualMachineImage_base64.txt* 
   on the harvester node

   .. code-block:: bash

      mkdir -p VirtualMachineImage
      base64 -d VirtualMachineImage_base64.txt | tar -C VirtualMachineImage -xvz

#. Apply VirtualMachineImage Kubernetes resource.

   .. code-block:: bash

      kubectl apply -f VirtualMachineImage/ubuntu-2204-jammy-v20240308.yaml
   
   .. note::
      If the following error encountered while applying the resource ``error validating "VirtualMachineImage/ubuntu-2204-jammy-v20240308.yaml": error validating data: ValidationError(VirtualMachineImage.spec): unknown field "retry" in
      io.harvesterhci.v1beta1.VirtualMachineImage.spec; if you choose to ignore these errors, turn validation off with --validate=false``, run the above command with *\-\-validate=false*

#. Confirm that the VirtualMachineImage is downloaded on harvester

   .. code-block:: bash

      kubectl get VirtualMachineImage ubuntu-2204-jammy-v20240308 -o yaml

   "status.conditions[type=Imported].status" should be True.

   .. figure:: ../../_figures/guides/vmaas_update_machine_image_11.png

#. Repeat steps 5 to 8 for each SPR harvester cluster in all the regions

#. Create a Pull Request or push changes to the same branch the one created in the previous section. Wait for Jenkins pipeline to finish successfully and upon receiving approval, merge the PR
   refer `Machine Image Update PR`_.

   .. figure:: ../../_figures/guides/vmaas_update_machine_image_12.png
      :alt: PR for Updating Default VMaaS Machine Image  

#. Create another PR so as to update environment with latest image. This can be done by updating ``computePopulateMachineImage`` for all the regions in file
   ``deployment/universe_deployer/environments/<env>.json`` with the commit id of the PR merged in the previous step. Wait for Jenkins pipeline to finish
   successfully and upon receiving approval, merge the PR refer `Machine Image Update Staging PR`_.

#. Once the above PR is merged, Jenkins pipeline will run on the commit in the main branch. Wait for the pipeline to finish stage ``Bazel Universe Deployer``.
   Once it is finished click on the **Details** link and the bottom of the logs you will find a link to create a PR on argocd repo with a generated branch name

   .. figure:: ../../_figures/guides/vmaas_update_machine_image_13.png
      :alt: Jenkins Pipeline

   .. figure:: ../../_figures/guides/vmaas_update_machine_image_14.png
      :alt: Link to create branch on argocd

#. Create another PR by clicking on the agrocd PR link refer `Machine Image Update Argocd PR`_.
   
   .. note::
      Make sure the base branch points to main and the branch in comparison is the one generated by the universe deployer. 
      In Addtion just make sure the updated values.yaml are pointing to the the correct commit id from main

#. Upon receiving approval merge the PR and wait for sometime for the job **populate-machine-image-git-to-grpc-synchr** to finish in all the regions.
   Once the job is complete , the new machine image should be available in all the regions.

#. Verify that new image is available via UI and make sure the console defaults to the new MachineImage but allows the old MachineImage to be selected

   .. figure:: ../../_figures/guides/vmaas_update_machine_image_15.png
      :alt: UI with New Machine Image

.. _Software Inflow Repo: https://github.com/intel-innersource/frameworks.cloud.devcloud.services.inflow
.. _IDC Monorepo: https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc
.. _Ubuntu Releases: https://cloud-images.ubuntu.com/releases/jammy
.. _JFrog Installation Page: https://jfrog.com/getcli
.. _IDC Inflow JFrog Repo: https://internal-placeholder.com/ui/repos/tree/General/idc_inflow-igk-local
.. _Dev Jf Console: https://dev-jf-internal-placeholder.com
.. _Machine Image S3 Bucket: https://s3.console.aws.amazon.com/s3/buckets/catalog-fs-dev
.. _AWS IDC Usage: https://internal-placeholder.com/x/LSl2sQ
.. _Connect Staging Harvester1 Cluster: https://internal-placeholder.com/pages/viewpage.action?pageId=3118345036#IDCPDX05usstaging1StagingRegion-ForHarvesterclusterharvester1
.. _Machine Image Update PR: https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/pull/4808
.. _Machine Image Update Staging PR: https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/pull/4990
.. _Machine Image Update Argocd PR: https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc-argocd/pull/2018
.. _Machine Image Inflow PR: https://github.com/intel-innersource/frameworks.cloud.devcloud.services.inflow/pull/195
.. _QCOW2 URL Link: https://internal-placeholder.com/artifactory/idc_inflow-igk-local/PR/PR-195/3/builds/recipe.tenant.vmaas_base.yaml/ci_image.qcow2
.. _VirtualMachineImage YAML Ubuntu 20230122: https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/main/build/environments/prod/us-region-1/VirtualMachineImage/ubuntu-2204-jammy-v20230122.yaml
.. _MachineImage YAML Ubuntu 20230122: https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/main/build/environments/prod/MachineImage/ubuntu-2204-jammy-v20230122.yaml
.. _Harvester1 SSH Secret: https://internal-placeholder.com/ui/vault/secrets/dev-idc-env/kv/list/shared/harvester1