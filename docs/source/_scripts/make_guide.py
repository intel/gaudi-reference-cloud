# !/usr/bin/env python3
# Author: michael.vincerra@intel.com
import os, sys
import subprocess as sps
from pathlib import Path
import shutil
import re
from tut_demo import cp_tut
import subprocess as sps

def exiter():
    sys.exit("\nExiting.")

def cp_guide(): 
    '''Copy guide_tmp.rst to source/public/guides by default'''
    try: 
        guide_template = Path(__file__).resolve().parents[1].joinpath(".rst-templates", "guide_tmp.rst")
        public_guides = os.path.join("public" ,"guides")
        guide_dest = Path(__file__).resolve().parents[1].joinpath(public_guides,"guide_tmp.rst") 
        cp_tmp_2src = shutil.copyfile(guide_template, guide_dest)
        return cp_tmp_2src
    except Exception as e:
        print(e)

def main():
    '''Copy guide_tmp.rst to source/public/guides; else exit'''
    try: 
        copy_guide_tmp= cp_guide()
        print("\nFollow template instructions.") 
        print("\nSave file to appropriate directory and manage in Git.")
        return copy_guide_tmp
    except Exception as e:
        print(e)
        exiter()

if __name__ == "__main__":
    main()
