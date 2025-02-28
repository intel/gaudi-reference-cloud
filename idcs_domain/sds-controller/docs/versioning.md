<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
# Versioning

This repository uses `SemVer` versioning scheme.

New version is resolved automatically using [GitVersion](https://gitversion.net/). Next tag is calculated based on the set of parameters such as current branch, existing tags and commit messages.
By default each merge to main branch updates minor version.

If one needs to introduce major change it must have "major" keyword in the beginning of the squashed commit message (pull request name)
