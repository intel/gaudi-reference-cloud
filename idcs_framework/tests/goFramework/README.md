*********************************************
REST API Automation Framework - GO LANG
*********************************************

This is a REST API based Automation framework. The framework aims at testing all the possible components of a REST based product. All type of Functional, Sanity, Regression, scenarios and negative testing are covered in this project.

TOC
"""""""""""""""""
1. Pre-Requisites & Installation
2. Configuration & Execution Steps
3. Documentation
4. Usage for Unit Testing on Local Setup
5. Points to be noted
6. Error Scenarios
7. Bugs & Contributions
8. References and Links

1. Pre-Requisites & Installation
---------------------------------
   1.1 Check go is installed.

   1.2 Check whether the Test Servers have required internet connectivity.

   1.3 Setup Git Bash in your system to clone this project.

   1.4 Git clone the project from the repo https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc.git using -       

       # With HTTPS : git clone https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc.git

   1.5 Run the go.mod file under idcs_framework/tests/goFramework to install all packages required by the project to run
       flawlessly.

       PS:
           - This is a one time activity, unless you are using a different system/env or have uninstalled the packages manually.
             command : go mod download            


2. Configuration & Execution Steps
------------------------------------
   2.1 There are two go frameworks available in the repo, testify and ginkGO. Both frameworks follow same directory structure
   After all package installations , start editing the input config file in respective framework folders
       i.e. /test_config/config.json file.   
   2.2 The config.json file, present in /test_config folder needs to be updated
       manually with the proper details of machine under test as well as payloads inputs.
   2.3 The values fed in the 'config.json' will lead to the  input payloads
       for the REST-API calls, and hence they should be filled carefully.   

       - PS:
              - Mention the cluster name and token under cluster_info section, framework will pick cluster details as per the name mentioned
              - Update Report portal details in the report portal section
              - Replace all "namespace": "Test namespace" vaues by providing proper namespace in config.json
              - Replace all "sshPublicKeyNames": ["ssh Public Key"] with proper public key value          

3. Documentation
------------------
   3.1 Link to the framework design document is mentioned below -
    { Document link need to be added here}

4. Unit Testing on Local Setup
-----------------------------------
   4.1 Git Clone the project repo as mentioned in above steps.

   4.2 Follow all the steps as mentioned in 'Installations' section.   

   4.3 To execute the complete Test-Automation, run the following command from the path
   idcs_framework/tests/goFramework/(testify or ginkGo)/test_cases/APITest where the project is cloned -
       # go test -v  -timeout 99999s 
   4.4 To execute the specific test cases in Test-Automation, run the following command from the path
       idcs_framework/tests/goFramework/(testify or ginkGo)/test_cases/APITest where the project is cloned -
      # go test -v --tags=<test_tag> -timeout 99999s 

     - PS:
            - test tags are mentioned in test files in the build section
            - Ex : In test_cases/APITest/large_vm_api_test.go file tags are mentioned 
                    // +build Functional LargeVM VMaaS
               In the above example Functional, LargeVM, VMaaS are test tags
               go test -v --tags=Functional -timeout 99999s         
 

5. Points to be noted
-----------------------
   5.1 Please make sure to feed correct values in the config.json file to avoid test
       failures due to incorrect/invalid inputs.

   5.2 If there is any change by the developers or any Api modifications related to
       keys or values in the schema/input files, then config.json needs to be updated
       again with the proper payloads.   


6. Error Scenarios
------------------

   6.1 Check Proper auth token is being updated in config.json file, invalid auth token may lead to test failures 
   6.2 Report portal related inputs should be properly updated in config.json
   6.3 REST Endpoint needs to be reachable from the system where the tests are being executed


7. Bugs & Contributions
--------------------------

   7.1 All the bugs, if any, should be raised on below mentioned JIRA board.
       https://internal-placeholder.com


8. References and Links
--------------------------

   8.1 All the framework supporting documents are uploaded at the below mentioned
       Confluence link-
       https://internal-placeholder.com/display/INTCS/IDC+Quality+Engineering
   
