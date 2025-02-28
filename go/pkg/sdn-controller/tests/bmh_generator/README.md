This tool help generates up to 8 NodeGroups and 256 Gaudi BMH nodes(8 nodes in a NodeGroup). 

run `go run generate.go` to generate 256 BMHs and the Switch CRs. When SDN Controller is running, it will create the NodeGroups, NetworkNodes and SwitchPorts. 

run `go run generate.go --n=1` to generate BMHs for one NodeGroup

run `go run generate.go --action=delete` to cleanup the generated resources.
