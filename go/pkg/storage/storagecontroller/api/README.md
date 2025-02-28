<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
# SDS-Controller API
## Why APIs are vendored?
API is vendored from `https://github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/api/storagecontroller/v1` 

Reasons why it was vendored is:
- Project with remote repository dependency to internal intel gitlab, will stop building without VPN
- Remote repo uses different version proto validation library https://github.com/bufbuild/protovalidate, bringing it up will force update of the common lib, and can introduce breaking changes
- Build can be complicated and require builder/ci access to private intel source github repo

After all above problems are fixed, or at least no longer valid (e.g. with VPN, if internal connection is required) then vendored files can be replaced with `git_repository` rule.

## How to generate pb files for IDE support?
Run
`bazel run //go/pkg/storage/storagecontroller:update_gen`