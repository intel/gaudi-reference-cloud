# !/usr/bin/env python3
# Author: michael.vincerra@intel.com
from tut_demo import cp_tut
import subprocess as sps

def main():
    '''From tut_demo.py, copy tutorial_tmp.rst to source/public/tutorials'''
    try: 
        copy_tut_tmp= cp_tut()
        print("\nFollow template instructions.") 
        print("\nSave file to appropriate directory and manage in Git.")        
        return copy_tut_tmp
    except Exception as e:
        print(e)

if __name__ == "__main__":
    main()
