# Hashcat
[hashcat](https://github.com/hashcat/hashcat/) is a password recovery utility with highly-optimized hashing algorithms support.   
The guide provides the instructions to build and benchmarks on Intel Data Center GPU.   
Please go to [hashcat](https://github.com/hashcat/hashcat/) github repo for futher details.

## Hashcat Build
```
git clone https://github.com/hashcat/hashcat/
cd hashcat
git checkout 
make -j
```

## Hashcat Benchmark
```
# List the supported hash mode
$ ./hashcat -hh

- [ Hash Modes ] -

      # | Name                                                       | Category
  ======+============================================================+======================================
    900 | MD4                                                        | Raw Hash
      0 | MD5                                                        | Raw Hash
    100 | SHA1                                                       | Raw Hash
   1300 | SHA2-224                                                   | Raw Hash
   1400 | SHA2-256                                                   | Raw Hash

# Run benchmark for a given hash mode, e.g. MD5
$ ./hashcat -b -m 0

# Run benchmark for all hash mode. This may take several hours.
$ ./hashcat -b

# Run test on specific GPU device, use ZE_AFFINITY_MASK for device filter, e.g. on device 0
$ ZE_AFFINITY_MASK=0 ./hashcat -b -m 0

```