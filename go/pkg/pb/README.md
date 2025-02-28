<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
# go code from protobuf compiler

To run the protobuf compiler after changing .proto files, run

`go generate ./..` in the top-level go directory

Run go generate in the top-level go directory rather than
because there is other auto-generated code that depends on
the .proto files.
