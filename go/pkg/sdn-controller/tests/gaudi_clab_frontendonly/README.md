
# deploy the containerlab
go to "<IDC_Project_Root>/networking/containerlab/frontendonly" and run 
```
sudo containerlab deploy
```

# Create the switch & switchport CRs
go to "<IDC_Project_Root>/go/pkg/sdn-controller/tests/gaudi_clab_frontendonly", run
```
kubectl apply -f switches.yml
kubectl apply -f provider-switchports.yml
```

# run SDN-Controller
go to "<IDC_Project_Root>/go/pkg/sdn-controller", run
```
go run main.go --config="tests/gaudi_clab_frontendonly/controller_manager_config.yaml"
```
