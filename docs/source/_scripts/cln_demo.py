# !/usr/bin/env python3
# Author: michael.vincerra@intel.com

from git.repo import Repo
import git
import os, sys
import subprocess as sps
from subprocess import check_output
from pathlib import Path
from os.path import normpath
from time import sleep

def demo_files_2rm():
    '''Create a list of files to be removed in Git'''
    tmp_files = [ ]
    public_tut_path = os.path.join("docs","source","public" ,"tutorials")
    tut_template = Path(__file__).resolve().parents[2].joinpath(public_tut_path,"tutorial_tmp.rst")

    tmp_files.append('/'.join(tut_template.parts[-4:]))
    tut_idx  = Path(__file__).resolve().parents[2].joinpath(public_tut_path,"index.rst")
    
    tmp_files.append('/'.join(tut_idx.parts[-4:]))
    # print("\n",tmp_files,"\n")
    return tmp_files

def git_reset():
    '''Unstage and reset demo files in Git repo without commit'''    
    doc_root = Path(__file__).resolve().parents[1] # run from docs dir
    filepaths = demo_files_2rm()
    files_2rm = [f for f in filepaths]
    git_tut_tmp = os.path.join("docs",files_2rm[0]) # prepend docs for Git
    git_tut_idx = os.path.join("docs",files_2rm[1]) # prepend docs for Git 

    repo=git.Repo(doc_root, search_parent_directories=True)

    untracked = repo.untracked_files
    unt = [u for u in untracked]

    modified = [m.a_path for m in repo.index.diff(None)]
    mod = [i for i in modified]

    if git_tut_tmp in unt and git_tut_idx in mod:
        print(f"Demo files reset: \n{git_tut_tmp}\n{git_tut_idx}")
        repo.git.clean("-f", git_tut_tmp)
        repo.git.checkout("--", git_tut_idx)
        print(f"\nDemo files removed successfully from: {repo.active_branch}")
        return
    else:
        print("\nReturn to docs directory and start over.\n")

def clean_html():
    '''Based on host OS, make clean accordingly and return to pwd '''
    grand_parent = Path(__file__).resolve().parents[2]
    curr_dir = os.getcwd()
    os.chdir(grand_parent)
    if os.name == "nt":
        print(f"Operating System: {os.name}")
        win = sps.run([".\make.bat", "clean"])
        os.chdir(curr_dir)
        return win
    else:
        print(f"Operating System: {os.name}")
        lnx = sps.run(["make", "-f", "Makefile", "clean"])
        os.chdir(curr_dir)
        return lnx

def main():
    '''Orchestrate make clean and git reset'''
    clean_html()
    clean_tut_demo = git_reset()
    return clean_tut_demo

if __name__ == "__main__":
    main()
