<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
# Release Log
Weka Storage fix for GB to TB conversion & Disable for VMs
## [Release Branch] - 2025-29-01
- [prod-012925](https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/commits/prod-012925/)
### Added
### Changed
- [IDCK8S-995](https://jira.devtools.intel.com/browse/IDCK8S-995)
  Solution: Making sure we are trying to handling the GB to TB conversion properly. Also Disable weka for VMs.
### Fixed

# Release Log
Security finding P1 fixes for kubernetes operator
## [Release Branch] - 2025-06-01
- [iks-prod-121924-tmp](https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/tree/iks-prod-121924-tmp)
### Added
### Changed
- [IDCK8S-1007](https://internal-placeholder.com/browse/IDCK8S-1007)
Solution: Fix firewall statuses reconciliation logic.
  The original problem wasn't related to firewall rules, but to wrong status reflection of the rule.
  When user was creating or updating the rule, the status was updated to Active, while in fact the rule didn't become active yet.
  This fix apply required changes to reflect the real status of FW rule.
### Fixed
IKS P1 Security Fixes for IKS