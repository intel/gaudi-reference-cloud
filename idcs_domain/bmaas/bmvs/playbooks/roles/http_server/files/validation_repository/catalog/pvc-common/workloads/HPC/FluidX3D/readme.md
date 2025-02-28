# FluidX3D
source $ONE_API_DIR
## How to build
```
git clone https://github.com/ProjectPhysX/FluidX3D.git

Change config file, by default, it is FP32 with D3Q19 SRT
Change src/info.hpp could change it to FP16S or FP16C
Change src/defines.hpp could change D3Q19 to D3Q15 D3Q27 or D2Q9 
Change src/setup.cpp line 18 for multiple GPU purpose

cd FluidX3D
sh make.sh 
```
## How to run

```
./bin/FluidX3D
```

## Sample out
```
.-----------------------------------------------------------------------------.
|                       ______________   ______________                       |
|                       \   ________  | |  ________   /                       |
|                        \  \       | | | |       /  /                        |
|                         \  \      | | | |      /  /                         |
|                          \  \     | | | |     /  /                          |
|                           \  \_.-"  | |  "-._/  /                           |
|                            \    _.-" _ "-._    /                            |
|                             \.-" _.-" "-._ "-./                             |
|                               .-"  .-"-.  "-.                               |
|                               \  v"     "v  /                               |
|                                \  \     /  /                                |
|                                 \  \   /  /                                 |
|                                  \  \ /  /                                  |
|                                   \  '  /                                   |
|                                    \   /                                    |
|                                     \ /                FluidX3D Version 2.9 |
|                                      '     Copyright (c) Dr. Moritz Lehmann |
|-----------------------------------------------------------------------------|
|----------------.------------------------------------------------------------|
| Device ID    0 | Intel(R) Data Center GPU Max 1550                          |
|----------------'------------------------------------------------------------|
|----------------.------------------------------------------------------------|
| Device ID      | 0                                                          |
| Device Name    | Intel(R) Data Center GPU Max 1550                          |
| Device Vendor  | Intel(R) Corporation                                       |
| Device Driver  | 23.22.26516.25                                             |
| OpenCL Version | OpenCL C 1.2                                               |
| Compute Units  | 512 at 1600 MHz (8192 cores, 26.214 TFLOPs/s)              |
| Memory, Cache  | 62244 MB, 196608 KB global / 128 KB local                  |
| Buffer Limits  | 62244 MB global, 63737856 KB constant                      |
|----------------'------------------------------------------------------------|
| Info: OpenCL C code successfully compiled.                                  |
| Info: Allocating memory. This may take a few seconds.                       |
|-----------------.-----------------------------------------------------------|
| Grid Resolution |                                256 x 256 x 256 = 16777216 |
| Grid Domains    |                                             1 x 1 x 1 = 1 |
| LBM Type        |                                     D3Q19 SRT (FP32/FP32) |
| Memory Usage    |                                CPU 272 MB, GPU 1x 1488 MB |
| Max Alloc Size  |                                                   1216 MB |
| Time Steps      |                                                        10 |
| Kin. Viscosity  |                                                1.00000000 |
| Relaxation Time |                                                3.50000000 |
| Reynolds Number |                                                  Re < 148 |
|---------.-------'-----.-----------.-------------------.---------------------|
| MLUPs   | Bandwidth   | Steps/s   | Current Step      | Time Remaining      |
|    3941 |    603 GB/s |       235 |         9988  80% |                  0s |
|---------'-------------'-----------'-------------------'---------------------|
| Info: Peak MLUPs/s = 3946                                                   |
```
