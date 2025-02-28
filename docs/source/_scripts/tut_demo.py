# !/usr/bin/env python3
# Author: michael.vincerra@intel.com
import os, sys
import subprocess as sps
from pathlib import Path
import shutil
import re

def exiter():
    sys.exit("\nExiting.")

def cp_tut(): 
    '''Copy tutorial_tmp.rst to source/public/tutorials by default'''
    try: 
        tut_template = Path(__file__).resolve().parents[1].joinpath(".rst-templates", "tutorial_tmp.rst")
        public_tuts = os.path.join("public" ,"tutorials")
        tut_dest = Path(__file__).resolve().parents[1].joinpath(public_tuts,"tutorial_tmp.rst")
        print("\n",tut_dest)
        cp_tmp_2src = shutil.copyfile(tut_template, tut_dest)
        return cp_tmp_2src
    except Exception as e:
        print(e)

def validate_no_tut(tut_tmp):
    '''Ensure that tutorial_tmp.rst does not exist at source/public/tutorials.'''
    public_tut_path = os.path.join("public" ,"tutorials")
    tut_idx = Path(__file__).resolve().parents[1].joinpath(public_tut_path, "index.rst")
    tut_tmp_file = Path(__file__).resolve().parents[1].joinpath("tutorial_tmp.rst")
    # print(tut_idx)
    matched = re.compile(r"\b" + re.escape(tut_tmp) + r"\b").search
    with open(tut_idx) as file:
        valid = not any(matched(line) for line in file)
        if valid == True and not os.path.exists(tut_tmp_file):
            return valid
        else:   
            print('''
                  \nWARNING: Tutorial demo was not cleaned.
                  \nBefore proceeding, run per your OS:
                  \n"make tutclean" for Linux / MacOS.
                  \n".\make.bat tutclean" for PowerShell.
                  ''')
            exiter()
        
def add_tut_link_toctree(idx_file):
    '''Add tutorial_tmp.rst to toctree in index in source/public/tutorials'''
    t = "tutorial_tmp"
    t_pad = t.rjust(15) # 12 chars + 3 empty spaces for indent under doctree
    public_tut_path = os.path.join("public" ,"tutorials")
    tut_idx = Path(__file__).resolve().parents[1].joinpath(public_tut_path,idx_file)
    try:
        with open(tut_idx, 'a', encoding='utf-8') as output_file:
            output_file.write("\n")
            output_file.write(t_pad)
    except Exception as e:
        print(e)
  
def make_tut():
    '''Based on host OS, make html accordingly and return to pwd '''
    grand_parent = Path(__file__).resolve().parents[2]
    curr_dir = os.getcwd()
    os.chdir(grand_parent)
    if os.name == "nt":
       print(f"Operating System: {os.name}")
       win = sps.run([".\make.bat", "html"])
       os.chdir(curr_dir)
       return  win
    else:
       lnx = sps.run(["make", "-f", "Makefile", "html"])
       print(f"Operating System: {os.name}")
       os.chdir(curr_dir)
       return lnx
    
def main():
    '''If no tutorial_tmp exists, copy tutorial_tmp to source/public/tutorials and make html'''
    if validate_no_tut("tutorial_tmp") is True:
        cp_tut()
        add_tut_link_toctree("index.rst")
        make_tut()
    else:
        exiter()

if __name__ == "__main__":
    main()
