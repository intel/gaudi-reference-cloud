// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import CodeLine from '../../../utils/CodeLine'
import idcConfig from '../../../config/configurator'

const HowToCreateSSHkeyWindows = (props) => {
  return (
    <>
      <ol className="w-100 pe-s6">
        <li>Launch a new PowerShell window on your local system.</li>
        <li>
          Optional: if you haven&rsquo;t generated a key before, create an .ssh directory.
          <CodeLine codeline={'mkdir $env:UserProfile\\.ssh'} />
        </li>
        <li>
          Copy & paste the following to your terminal to generate SSH Keys.
          <CodeLine codeline={'ssh-keygen -t rsa -b 4096 -f $env:UserProfile\\.ssh\\id_rsa'} />
        </li>
        <li>
          If you are prompted to overwrite, select <strong className="text-danger">No</strong>.
        </li>
        <li>
          Copy & paste the following to your Powershell window to view your SSH key:
          <CodeLine codeline={'cat $env:UserProfile\\.ssh\\id_rsa.pub'} />
        </li>
        <li>Paste your key&rsquo;s contents in the field below.</li>
      </ol>
      <span className="valid-feedback">
        For more information go to{' '}
        <a rel="noreferrer" href={idcConfig.REACT_APP_SHH_KEYS} target="_blank">
          SSH key documentation
        </a>
        .
      </span>
    </>
  )
}

export default HowToCreateSSHkeyWindows
