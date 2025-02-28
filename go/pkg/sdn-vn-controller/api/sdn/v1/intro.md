<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2024 Intel Corporation-->
# Note: This is actively being worked upon so consider this as WIP
# Purpose:

This API is for use by **NaaS** to communicate with SDN Controller for orchestrating Tenant Network resources.
<img src="interaction.png" width="800" height="400">

## Resources
The network resources available through this API are:

 - VPC
 - Subnet
 - Router
 - Port
 - Internal IP
 - Router Interface
 - Static Route
 - Security Group
 - Security Rule
 - Elastic IP

### Resource Model
There inter-relationship is expressed in this diagram below

![alt text](orm.png "")

## API Pattern
For most of the operations the following pattern of methods should be used.

 - Create : create a new resource.
 - Get : retrieve an existing resource.
 - Update : modify an existing resource.
 - Delete : destroy an existing resource.
 - List : get a collection of resources meeting the list criteria.
 - Replace : replace an entire resource as compared to updating a resource. This is likely to be not needed but documented here in case it is needed.

### Usage Notes
 - UUIDs are to be used as the primary identifier for all objects.
 - Names must be unique among all resources of the same type.
 - The API is declarative in its behavior.

#### Ordering of operations

 |Object|Create these first|Delete these first|
|:----|:----|:----|
|VPC|None|Router, Subnet, Port, Security Group, Security Rule|
|Router|VPC|Static Route, Router Interface|
|Subnet|VPC|Port, Router Interface, Assignment to any Security Group |
|Port|VPC, Subnet|Assignment to any Security Group|
|Security Group|VPC|Assignment to any Port or Subnet. Membership of Security Rule(s) to the Group|
|Security Rules|VPC, Security Group|Membership to any Security Group|
|Static Routes|Router| |
|Router Interfaces|Subnet, Router| |
|Address Translations|Port, Router Interface|(Address Translations get deleted implicitly when Port or Router Interface gets deleted)|
|

#### Known Limitations

 - Subnet CIDRs as well as Route prefixes used in Static routes must have prefix length less than 30. In other words /30, /31, and /32 should not be used in Subnet definition or Static routes.

 - A Security rule can be member of only one Security Group. 
 - A Security Group can be applied to either a set of Ports or a set of Subnets but not to both simultaneously.

### TODOs 
 -   any aggregate queries needed by NaaS? Typically NaaS should be interacting for specific objects because for normal lists etc.  it should be using it's own DB
 -   Add description for salient elements
 -   define pre-validation
 -   refine error reporting