
# deploy the containerlab
go to "<IDC_Project_Root>/networking/containerlab/allscfabrics" and run 
```
sudo containerlab deploy
```

# generate the CRs
go to "<IDC_Project_Root>/go/pkg/sdn-controller/tests/gaudi_clab", run
```
go run generate.go
```

# run SDN-Controller
go to "<IDC_Project_Root>/go/pkg/sdn-controller", run
```
go run main.go --config="tests/gaudi_clab/controller_manager_config.yaml"
```

# cleanup the CRs
go to "<IDC_Project_Root>/go/pkg/sdn-controller/tests/gaudi_clab", run
```
go run generate.go --action=delete
```

