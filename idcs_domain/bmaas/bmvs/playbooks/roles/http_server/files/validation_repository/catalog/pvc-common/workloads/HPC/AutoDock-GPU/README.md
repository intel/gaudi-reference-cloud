## AutoDock-GPU

### Get source code and build

```
# Download source code
git clone https://github.com/emascarenhas/AutoDock-GPU.git

# Initialize oneAPI environment
source /opt/intel/oneapi/setvars.sh

# Build
cd AutoDock-GPU
make DEVICE=XeGPU NUMWI=64

# After the build, the autodock_xegpu_64wi binary will be generated in the bin directory
```
Please note that the best value for NUMWI (work-group size) depends on the workload. Accepted values are 1,2,4,8,16,32,64,128,256,512,1024.

### Run AUtoDock-GPU

```
# Initialize oneAPI environment 
source /opt/intel/oneapi/setvars.sh
```

#### Basic command
```
./bin/autodock_xegpu_64wi \
--ffile <protein>.maps.fld \
--lfile <ligand>.pdbqt \
--nrun <nruns>
```

| Mandatory options|   | Description   | Value                     |
|:----------------:|:-:|:-------------:|:-------------------------:|
|--ffile           |-M |Protein file   |&lt;protein&gt;.maps.fld   |
|--lfile           |-L |Ligand file    |&lt;ligand&gt;.pdbqt       |

Please see for additional supported arguments, see github.com/emascarenhas/AutoDock-GPU/tree/develop


#### Example command
```
./bin/autodock_xegpu_64wi \
--ffile ./input/1stp/derived/1stp_protein.maps.fld \
--lfile ./input/1stp/derived/1stp_ligand.pdbqt \
-nrun 100
```
