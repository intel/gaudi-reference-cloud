.. _release_notes:

Release Notes
#############

19 February 2025
*****************

- Version information

  - Intel® Tiber™ AI Cloud 2.2.0

- Release information:

  - General Availability (GA)

  - Release Date: 02/19/2025

  - Publish date of release notes: 02/19/2025

- Summary:

  - We released Cloud Monitor for bare metal and virtual machine (VM) compute instances. We also released Cloud monitor for |IKS|, where both
    **Kubernetes API Server** and **Kubernetes etcd** metrics are available. We hope users leverage these capabilities to improve management of instances, instance groups, and Kubernetes environments.

  - We released **Credentials as a Service**, which enables users to interact with regional services by generating a `client_secret` and
    `client_id`.

  - **Documentation**

    - First, we developed and released the :ref:`credentials` guide to accompany this new feature in the console UI.

    - Second, we released a new subsection, "Cloud Monitor", with guides explaining how to use logging and metrics for instances or Kubernetes environments.

    - Third, we developed a new Sphinx extension to support automated creation of the "Software" table for JupyterLab dependencies.

06 February 2025
*****************

- Version information

  - Intel® Tiber™ AI Cloud 2.1.3

- Release information:

  - General Availability (GA)

  - Release Date: 02/06/2025

  - Publish date of release notes: 02/06/2025

- Summary:

  - We implemented a security enhancement: Cloud Account members will be logged out automatically if an administrator revokes their access.

  - We released two new tutorials and optimized our instances pages for small screens.

  - **Documentation**

    - We released two new tutorials. The first, :ref:`deepseek_r1`, shows how to run on |ITAC| DeepSeek-R1-Distill-Llama-70B model with the |GPUMAX| 1550 or 1100 series. The second, published on Hugging Face, shows how to fine-tune `Meta Llama 3.2-Vision-Instruct Multimodal LLM on Intel Accelerators`_.

    - We completed tooling and design functionality to support users viewing instance specifications on multiple device types. For example, users can now view :ref:`ai_instances`, including pricing, on tablet or smartphones without missing any details.

19 December 2024
*****************

- Version information

  - Intel® Tiber™ AI Cloud 2.1.2

- Release information:

  - General Availability (GA)

  - Release Date: 12/19/2024

  - Publish date of release notes: 12/19/2024

- Summary:

  - We're thrilled to announce a significant price reduction on our :ref:`Intel® Data Center GPU Max Series <gpu_instances>` instances. Power up your AI projects with our more affordable GPU instances without breaking the bank.

  - We just launched our new public site at `cloud.intel.com`_. Explore our expanded product offerings, learn about our latest features, and discover upcoming events and collaborations.

  - **Documentation**

    - We fixed :ref:`staas_object`, adding the **Prerequisite** of using AWS\* CLI client version 2.13 or higher. This version is required in order to access the ``AWS_ENDPOINT_URL``.

    - We refactored python tooling, for ``sphinx-build`` automation, to incorporate a default United States dollar (USD) denomination for ``Price \hr``, and to use two (trailing) decimal points for all compute instances specifications.


16 December 2024
*****************

- Version information

  - Intel® Tiber™ AI Cloud 2.1.1

- Release information:

  - General Availability (GA)

  - Release Date: 12/16/2024

  - Publish date of release notes: 12/16/2024

- Summary:

  - Intel® Geti™ is now part of our software catalog. Intel Geti eases laborious data labeling, model training and optimization tasks across the AI model development process, empowering everyone to build OpenVINO™ optimized computer vision models, suitable for deployment at scale.

- Improvements

  - **Documentation**

    - Users may now may view Instance specifications, subdivided by recommended use case, in the :ref:`reference` section. Tables show instance names, price /hr, memory, disk, and more.
      This reference is intended to answer frequently asked questions and promote transparency to our customers.

  - Premium users now have access to Chat support via the support menu.

  - Miscellaneous improvements were made to region, learning catalog and account preferences.

05 December 2024
*****************

- Version information

  - Intel® Tiber™ AI Cloud 2.1.0

- Release information:

  - General Availability (GA)

  - Release Date: 12/05/2024

  - Publish date of release notes: 12/05/2024

- Summary:

  - Authorization as a Service (AaaS) is now available for Object Storage and File Storage. The default administrative user may assign roles and permissions to designate which users have access to storage.

  - JupyterLab environment was updated to include OpenVINO™ toolkit and the Intel® XPU Backend for Triton\* software.

- Improvements

  - **Documentation**

    - We made fixes to document layout design and improved hyperlink behavior.

21 November 2024
*****************

- Version information

  - Intel® Tiber™ AI Cloud 2.0.10

- Release information:

  - General Availability (GA)

  - Release Date: 11/21/2024

  - Publish date of release notes: 11/21/2024

- Summary:

  - This documentation only release includes a complete redesign of the documentation home page, five new service descriptions, and two new documents: "Start for free" and "Quick Start".

- Improvements

  - **Documentation**

    - We updated the |ITAC| documentation home page to highlight curated workflows, under the headers **Discover** and **Get Started**. Users may now launch a learning node by following "Quick Start" document. Or, users may enter "Start for free" and then click buttons to: access the Learning page (to launch Jupyter Notebooks); or request pre-release/early-release |INTC| hardware in Preview. All documentation has been updated to reflect that users in all regions may choose to connect to an instance using **One-Click** connection. The option remains to connect via a local Terminal with public SSH keys.

    - Five new service descriptions were added, providing definitions and descriptions of services such as Compute, |INTC| Kubernetes Service, Learning, Preview, and Storage.  With these new descriptions, we intend to ease onboarding for Technology Managers and Partners who seek a comprehensive understanding of our services, including their features, scope, and examples.

    - We also added an |ITAC| Overview, under **About** menu, to help users understand the key benefits of using |ITAC|.

19 November 2024
*****************

- Version information

  - Intel® Tiber™ AI Cloud 2.0.10

- Release information:

  - General Availability (GA)

  - Release Date: 11/19/2024

  - Publish date of release notes: 11/19/2024

- Summary:

  - Users can now access compute instances using a "One-Click" method. When used, it launches JupyterLab for all bare metal and virtual machine (VM) instances. This feature streamlines accessibility for users, including the ability to transfer files (upload/download) from an HTTPS browser.  This feature is now **available in all regions**.

14 November 2024
*****************

- Version information

  - Intel® Tiber™ AI Cloud 2.0.9

- Release information:

  - General Availability (GA)

  - Release Date: 11/14/2024

  - Publish date of release notes: 11/14/2024

- Summary:

  - Enable "One-Click" method for JupyterLab on all bare metal and virtual machine (VM) instances to provide streamlined accessibility for users, including the ability for file transfer (upload/download) from HTTPS browser. This feature is currently **only available in Region 3**.

  - Region 2 is now the default.

  - Users can now scale-up storage in |INTC| Kubernetes Service (IKS). To view, select |INTC| Kubernetes Service > Storage > Clusters tab.

- Improvements:

  - **Usability changes**:

    - The Console home page now includes a "Recently visited Links" widget.

    - Learning catalog items are now available in all regions.

    - Improved speed and performance in compute instances and Kubernetes\* clusters pages.

    - Accessibility: Users can now interact in dialog or pop-up menus via keyboard by pressing a key.

    - The Console buttons and UI elements are now squared.

  - **Documentation**

    - In :ref:`preview_cat`, we added instructions on how to connect to an instance, using remote desktop protocol, via :guilabel:`One-Click` connection.

    - In :ref:`manage_instance`, we added a new section on how to "Stop an Instance" and "Restart an instance"/

    - Tutorials

      - We added a tutorial on "Fine-tune Meta Llama-3.2-3B-Instruct", showing how to fine-tune Llama-3.2-3B on an |INTC| Gaudi 2 instance.

07 November 2024
*****************

- Version information

  - Intel® Tiber™ AI Cloud 2.0.8

- Release information:

  - General Availability (GA)

  - Release Date: 11/07/2024

  - Publish date of release notes: 11/07/2024

- Summary:

  - **One-Click connection**: Users can now conveniently access a new compute instances via One-Click connection, where instance launches in JupyterLab environment.  Available in us-region-3.

30 October 2024
*****************

- Version information

  - Intel® Tiber™ AI Cloud 2.0.7

- Release information:

  - General Availability (GA)

  - Release Date: 10/30/2024

  - Publish date of release notes: 10/31/2024

- Summary:

  - **Load Balancers**: Users can distribute traffic creating a Load Balancer for their instances with this new feature.

  - **Granite Rapids - General Availability**: |ITAC| now offers 6th generation |INTC| Xeon® Scalable processors, boasting up to 128 performance cores. These processors are specifically designed for high-performance computing applications.

  - **Object Storage - General Availability**: Object Storage capabilities are generally available to premium and enterprise users across all regions. Model training and inference often require storing large amounts of unstructured data. |ITAC| offers object storage, with a choice of CLI clients, to manage storage buckets.

28 October 2024
*****************

- Version information

  - Intel® Tiber™ AI Cloud 2.0.6

- Release information:

  - General Availability (GA)

  - Release Date: 10/28/2024

  - Publish date of release notes: 10/28/2024

- Summary:

  - In Intel Kubernetes Service (IKS), added ability for user to apply Security Rules, also known as firewall settings.

24 October 2024
*****************

- Version information

  - Intel® Tiber™ AI Cloud 2.0.5

- Release information:

  - General Availability (GA)

  - Release Date: 10/24/2024

  - Publish date of release notes: 10/24/2024

- Summary:

  - Updated from PyTorch 2.5rc to PyTorch 2.5, as applied to Learning Catalog Jupyter Notebooks.

- Improvements:

  - **Learning**

    - Changed name from Training Catalog to Learning Catalog in console UI.

  - **Usability changes**:

    - Redesigned the navigation sidebar menu; this menu is now open by default.

    - Modified the account summary widget so it adapts to display in smaller screens.
      Account summary widget now includes new items, and allows users to launch items with one click.

    - When users view the details of an item, the console UI automatically remembers the last selected tab.

    - When users stop a bare metal instance, user must confirm via a pop-up dialog.

    - In "How to connect to an instance", IP addresses adapt to the user-selected region.

    - Before launching an instance in "Instance Types", drop-down display is fixed.

    - In Intel Kubernetes Service, users can access the :file:`kubeconfig` file from the Details page.

03 October 2024
*****************

- Version information

  - Intel® Tiber™ AI Cloud 2.0.4

- Release information:

  - General Availability (GA)

  - Release Date: 10/03/2024

  - Publish date of release notes:  10/03/2024

- Summary:

  - **Brand**: Intel® Tiber™ AI Cloud - new product name for the AI production & deployment environment (for enterprise and large AI startups).

18 September 2024
*****************

- Version information

  - Intel® Tiber™ AI Cloud 2.0.3

- Release information:

  - General Availability (GA)

  - Release Date: 09/18/2024

  - Publish date of release notes:  09/18/2024

- Summary:

  - **Compute**: A new virtual machine (VM) instance type is available, Intel® Data Center GPU Max Series (1100). This new instance type supports a single GPU card per VM. The Intel® Data Center GPU Max Series is designed to manage AI and HPC data center workloads.

  - **Training**: In Training, when clicking a card, a panel shows Considerations and Environment (if available), including software dependencies for each Jupyter Notebook. This feature improves transparency for developers.

  - **PyTorch 2.5**: Upgraded AI framework, available in the JupyterLab environment, from PyTorch 2.4 to PyTorch 2.5.

  - **Regions**: A new US region, "us-region-3", was enabled.

  - **Supercomputing**: Now available in us-region-3, supercomputing enables users to train and deploy AI workloads at scale. This service allows customers to spin up clusters with a large quantity of bare metal CPU, GPU or AI processor nodes optimized for intensive workloads. Note: Currently, this service is only available only for select customers.

  - **Preview**: Users can now grant access to their Intel partners for co-development efforts. This enables close collaboration between Intel staff and ecosystem partners when evaluating pre-release hardware.

    - Users can now sign up to be a wait-listed for upcoming instance types. The first instance type to be added to the wait list is the Intel® Gaudi® AI Accelerator on bare metal.

- Improvements:

  - **Usability changes**:

    - In Training, the card behavior was changed so that when users click a card, a details page is shown for that item.

    - Users must enter a cloud coupon and reauthenticate to access Jupyter Notebooks inside Training.

    - Grids in the UI now support bigger page sizes by default. Also, the first column is now frozen in small screens.

    - The footer in console home page is fixed.

  - **Accessibility changes**: Added an ARIA label to Modals for improved screen reader behavior. Documentation footer link color is readable now.

-  **Documentation**

   -  Preview: Added three more documents separate from Preview Catalog: "Preview Storage", "Preview Keys", and "Preview". This change simplifies workflows into shorter, more succinct documents for developers.

- **Miscellaneous**:

  - Release notes for 22 July were modified to show when support was added for `PyTorch 2.4`_ in the JupyterLab environment.

04 September 2024
*****************

- Version information

  - Intel® Tiber™ AI Cloud 2.0.2

- Release information:

  - General Availability (GA)

  - Release Date: 09/04/2024

  - Publish date of release notes:  09/04/2024

- Summary:

  - **Preview Catalog**: Users can request to use Intel® Core™ Ultra processor family inside our preview catalog.
  - **Software catalog**: We added SeekrFlow\* to our software catalog. SeekrFlow product page provides information on its usage. 
  - **Landing page changes (cloud.intel.com)**: Added reference to Intel Core Ultra processor family.

- Improvements:

  - **Usability changes**: Clarifies descriptive UI components to improve user discoverability and functionality.

27 August 2024
**************

- Version information

  - Intel® Tiber™ AI Cloud 2.0.1

- Release information:

  - General Availability (GA)

  - Release Date: 08/27/2024

  - Publish date of release notes:  08/27/2024

- Summary

  - **Training - JupyterLab environment**: Users can now explore JupyterLab Notebook training that runs on Intel® Gaudi® 2 AI accelerator and Intel® Data Center Max Series GPU. Several installed oneAPI components, including the |INTC| oneAPI Toolkit, were updated to 2024.2. Several enhancements were made to increase the capacity, stability, and performance of the environment.

- Improvements:

  - **Documentation site**: The home page and some *Guides* and *Tutorials* landing pages now include badges: "AI-Dev",
    "Video", and "Tools". Using a consistent visual language, badges improve user navigation by providing users with visual clues, helping them quickly navigate from parent to child pages and find the resources they need.

  - **Documentation - JupyterLab Tutorial** - We added several new sections to the :ref:`jupyter_learning` tutorial to
    indicate the hardware and software on which Jupyter  Notebooks depend. We added a table showing the latest versions of
    |INTC| software, including the versions for several oneAPI components like the Intel® oneAPI Toolkit, Intel® Distribution
    for Python, and many more. Another table was added to identify which kernels are supported on Intel® Gaudi® 2 AI
    Accelerator or the Intel® Data Center GPU Max Series.

12 August 2024
**************

- Version information

  - Intel® Tiber™ AI Cloud 2.0.0

- Release information:

  - General Availability (GA)

  - Release Date: 08/12/2024

  - Publish date of release notes:  08/12/2024

- Summary

  - We're thrilled to roll out Intel® Tiber™ AI Cloud UI 2.0, a significant leap forward in our user interface, tailored to our growing suite
    of services. With this update, users will experience an improved navigation menu, enhanced aesthetics, and improved accessibility across various devices and screen sizes.

- New features:

  - **Contextual Documentation**: We've integrated our documentation guides and tutorials directly into the console. Now, by clicking the
    'Documentation' button on every page, users can browse documentation that's relevant to the current dashboard content. Guidance and resources are at your fingertips, exactly when and where you need them.

    **Get started sections**: When users first arrive at the console, they're offered links to recommended services that help meet their goals.

- Improvements:

  - **Responsive Across Devices**: The UI now adapts fluidly to a broader range of screen sizes, ensuring users enjoy a consistent and functional experience on any device.

  - **Documentation**: Site redesigned to match the new Intel® Tiber™ AI Cloud 2.0.0.

    Navigation of the Table of Contents reflects the same colors and styling as that of the console.

    The general :ref:`FAQ` was redesigned and new services are referenced.

22 July 2024
**************

- Version information

  - Intel® Tiber™ AI Cloud 1.0.16

- Release information

  - Release Type: General Availability (GA)

  - Release Date: 07/22/2024

  - Publish date of release notes: 08/12/2024

- Summary

  - **Landing page changes (cloud.intel.com)**: There will be only one option to create an account.
    In the past, there were 3 different registration options for service tiers (Standard, Premium, and Enterprise).

  - **Cloud console changes (console.cloud.intel.com)**:
    - The account tier screen is removed from the console.
    - After account creation, every new user starts as a Standard tier user.
    - Standard users will be able to see all available SKUs and services.
    - Users can upgrade to the Premium tier in the cloud console.

  - In the JupyterLab environment, we added support for PyTorch 2.4 for Jupyter Notebooks.
    See also `PyTorch 2.4`_.

- New features

  - Standard-tier users can request Premium-tier compute instances with a $/hr cost when cloud credits are
    applied to the user account.

- Improvements

  - **Instance auto-termination for standard tier accounts**
    - Active compute service instance reservations are automatically deleted when the user account runs out
    of cloud credits.
    - To avoid service downtime, the user will receive warning email messages and be requested to add more
    cloud credits or add a payment method to the account.

03 June 2024
************

- Version information

  - Intel® Tiber™ AI Cloud 1.0.15

- Release information

  - Release Type: General Availability (GA)

  - Release Date: 05/29/2024

  - Publish date of release notes: 06/03/2024

- Summary

  - Controlled access to Object Store Service available in |ITAC| Region-2 (Please contact support team for enabling account access).

- New features

  - Object storage service was added in Region-2 to support need to store large amounts of unstructured data used in model training and inference workflows. 

- **New or revised guides**:
  - Adds :ref:`staas_object` to assist users in creating storage buckets for a training and inference workflow.
  - Adds :ref:`staas_overview` to explain storage options available: file storage, or object storage.
  - Adds :ref:`processor model matrix <model_matrix>` to recommend an |INTC| processor filtered by large language model (LLM) use case.
  - Modifies :ref:`k8s_guide`, adding three new sections, including how to configure ingress and expose cluster services and deploy AI/ML workloads.

- **New or revised tutorials**:
    - Modifies Public Articles, adding five new public AI tutorials and one :ref:`video tutorials <tutorials>`.

- Improvements

  - Intel® Tiber™ AI Cloud now offers a performance improvement for billing and usage calculations for services provided. When you view your usage, usage data is retrieved and displayed much faster upon request.

  - Enhanced SSL security applied for critical back-end services.

03 April 2024
*************

- Version information

  - Intel® Tiber™ AI Cloud 1.0.12

- Release information
   
  - Release Type: General Availability (GA)

  - Release Date: 04/02/2024
   
  - Publish date of release notes: 04/03/2024

- Summary

  - Intel® Tiber™ AI Cloud 1.0.12 is the general availability release for storage as a service (STaaS), which includes file storage. The new service allows 
    Standard, Premium, and Enterprise account holders to create a storage volume. Storage quotas vary by account type.

  - This release includes initial launch of the Preview Catalog, which gives customers access to pre-release and early-release |INTC| hardware. 

- New features

  - **New guide**: 
    - Adds :ref:`staas_file` guide to explain CRUD operations for storage and how to mount a storage volume on a compute instance.
    - Adds :ref:`preview_cat` guide to explain how to request an instance in the Preview Catalog.

  - **New tutorials**: Adds Public Articles to increase audience exposure to publicly available Intel® Tiber™ AI Cloud articles, focused on GenAI and Machine Learning Operations (MLOps). 


08 February 2024
****************

- Version information

  - Intel® Tiber™ AI Cloud 1.0.11

- Release information
   
  - Release Type: General Availability (GA)

  - Release Date: 02/08/2024
   
  - Publish date of release notes: 02/08/2024

- Summary

  - Intel® Tiber™ AI Cloud 1.0.11 is the General Availability release for multi-user accounts, a feature that enables Premium and Enterprise account holders
    to invite others to share infrastructure and services. This feature improves collaboration, and it simplifies billing and infrastructure management.

- New features

  - **Multi-user Accounts**:A **Premium Account** or **Enterprise Accounts** holder may now invite others to share infrastructure and services by following 
    a secure invitation process. Users may be invited to join a Premium or Enterprise account. See :ref:`accounts`. 

  - **New guide**: Adds guide :ref:`multi_user_accounts` to explain multi-user account features and functionality.

25 January 2024
****************

- Version information

  - Intel® Tiber™ AI Cloud 1.0.10 

- Release information
   
  - Release Type: General Availability (GA)

  - Release Date: 01/23/2024
   
  - Publish date of release notes: 01/25/2024

- Summary

  - Intel® Tiber™ AI Cloud 1.0.10 is the General Availability release for Intel Kubernetes\* Service. This release also includes
    improvements to account type upgrades, to coupon duration, and it increases availability of hardware resources.

- New features

  - **Account Type Upgrades**:  Customers can now upgrade from a Standard to a Premium account using a **coupon**.

  - **Intel Kubernetes Service**: Intel Kubernetes Service is a fully managed container service that helps customers run GPU-accelerated Kubernetes workloads 
    at scale using Intel® Max Series GPU and |INTG2| Deep Learning Server.

  - **Hardware - Bare Metal**: Customers now have the option of using two-, four-, or eight-node clusters. 
    Increased capacity in region 2 results in more availability of hardware resources.

  - **Hardware - Virtual Machine**: Vnet Internal optimization. Baremetal-enrollment-api internal optimization. 

- Dependencies/requirements

  - **Coupons**: Coupon credit expiration was adjusted so customers can fully utilize credit for the intended duration. 


18 December 2023
****************

- Version information

  - Intel® Tiber™ AI Cloud 1.0.8

-  Release information
   
   - Release Type: General Availability (GA)

   - Release Date: 12/12/2023
   
   - Publish date of release notes: 12/18/2023

-  New features

   - **Payment Methods**: Customers can now use credit cards as a mode of payment.

   - **Account Type Upgrades**:  Customers can now upgrade from a Standard to a Premium account (only using credit card). This enables a smoother transition to innovative tools and technology for users who start with a Standard account. 

   - **Enterprise Accounts**: An Enterprise account type is now available. This account type includes monthly billing and invoicing. 

   - **Multi-region Support**: A second region was added to provide additional capacity for compute resources. This increases the availability of Gaudi2® Deep Learning Server nodes. 

   - **Training and Workshops** Users can expect improved accessibility for loading and training models. This was achieved by adding more SLURM clusters, supporting access to AI/ML development stacks and Jupyter Notebooks. 
     
   - **Hardware - Core Compute**: The platform added support for the 5th Generation Intel® Xeon® Scalable Processor for additional compute resources. 

   - **Hardware - HPC**: The platform added support for the 4th Gen Intel® Xeon® processors with high bandwidth memory (HBM) for additional compute resources.

   - **Hardware - GPU**: The platform added support for the Intel® Max Series GPU to support increased power consumption when doing model training and inferencing. 

-  Improvements

   - Improved deployment methods and processes provide higher availability for compute and faster releases of new features. 

24 October 2023
***************

- Version information

  - Product

- Release information

  - Release Type:

  - Release Date:

   - **Training and Workshops**: Jupyter Notebooks are now available for AI, C++ SYCL, and Generative AI (see below). This feature especially assists students,
     professors, and researchers by providing pre-loaded dependencies and executable examples.  

   - **Generative AI trainings**: We added new trainings on generative AI to our console.  Take a look at what generative AI can do when powered by the 
     Intel® Data Center GPU Max Series. Designed for everyone: from AI creators and artists to engineers and the just-plain curious.

   - **New tutorials**: Adds tutorial **Habana Gaudi 2**.

   - **Analytics tracking**: We added analytics tracking to our web app to help us continuously improve it and make it more user-friendly.

-  Improvements

   - **Account Settings**:  Users can now view their *Cloud Account ID*, *Name*, and *E-mail*. In the Intel® Tiber™ AI Cloud console, click on the User Profile icon
     (upper right) and select Account Settings. 

   - **Invoices**: Cloud credits now determine the ability to provision and launch an instances.  In addition, to simplify customer experience with   
     invoicing, enhancements to the anniversary dates were made. For Premium users, in the Intel® Tiber™ AI Cloud console, click on the User Profile icon
     (upper right) and Invoices. 



03 October 2023
***************

- Version information

  - Intel® Tiber™ AI Cloud 1.0.3


-  Release information

   -  Release Type: General Availability (GA)

   -  Release Date: 10/03/2023

   -  Publish date of release notes: 10/03/2023

-  New features

   -  Availability of Gaudi2® clusters (instance groups) on Gaudi2® Deep Learning Server, 4th Generation Intel®
      Xeon® Scalable processors, and Intel® Max Series GPU.

- Improvements

  - Improved user sign up process. The user can now select a service tier if problems are encountered during first use.

19 September 2023
*****************

-  Version information

   -  Intel® Tiber™ AI Cloud 1.0.2

-  Release information

   -  Release Type: General Availability (GA)

   -  Release Date: 09/19/2023

   -  Publish date of release notes: 09/19/2023

-  Summary

   -  Intel® Tiber™ AI Cloud 1.0.2 is the General Availability release for Intel Developer
      Cloud. In this first release, there is one region available based
      in US. Stay tuned for more exciting announcements coming shortly.

-  New features

   -  Intel® Tiber™ AI Cloud 1.0.2 provides the ability to provision Bare Metal providing
      access to Gaudi2® Deep Learning Server, 4th Generation Intel®
      Xeon® Scalable processors, and Intel® Max Series GPU.

   -  Intel® Tiber™ AI Cloud 1.0.2 provides the ability to provision Virtual Machine providing
      access to 4th Generation Intel® Xeon® Scalable processors.

-  Dependencies/requirements

   -  Supported Browsers include Firefox, Chrome, Safari, and Edge.

   -  Credit Card and Coupon are currently supported payment methods.

-  Documentation

   -  Please click the “Getting Started” link on the Learning and Support
      section of the Home page to access product tutorials and guides.

-  Support

   -  Please click the “Getting Started” link on the Learning and Support
      section of the Home page to access the Support section.

.. meta::
   :description: View release notes for the Intel® Tiber™ AI Cloud service platform.
   :keywords: release notes, release information, release date, version information

.. _PyTorch 2.4: https://www.intel.com/content/www/us/en/developer/articles/technical/pytorch-2-4-supports-gpus-accelerate-ai-workloads.html
.. _cloud.intel.com: https://cloud.intel.com
.. _Meta Llama 3.2-Vision-Instruct Multimodal LLM on Intel Accelerators: https://huggingface.co/blog/bconsolvo/llama3-vision-instruct-fine-tuning

- Summary