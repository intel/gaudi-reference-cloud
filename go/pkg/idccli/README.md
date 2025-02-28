<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
## Steps to Build the `idccli` Binary:

1. Install Visual Studio Code
2. Open the project
3. Open terminal and navigate to `$ ~/frameworks.cloud.devcloud.services.idc/go/pkg/idccli`
4. Run command `$ go build`
5. Command for cross-compiling the linux Go binary into Windows executable:
   `$ GOOS=windows GOARCH=amd64 go build -o idccli.exe`

## Supported Commands
```sh
$ ./idccli --help
$ ./idccli help-all
$ ./idccli ssh keygen --outputdir <output dir>
$ ./idccli ssh test-proxy --proxy-server <proxy server ip/name>
```
