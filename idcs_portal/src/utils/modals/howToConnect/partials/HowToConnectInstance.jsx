// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState } from 'react'
import Form from 'react-bootstrap/Form'
import useUserStore from '../../../../store/userStore/UserStore'
import CodeLine from '../../../CodeLine'
import idcConfig from '../../../../config/configurator'

const HowToConnectInstance = ({ data }) => {
  const isIntelUser = useUserStore((state) => state.isIntelUser)

  const [selectedValue, setSelectedValue] = useState('windows')

  if (!data) {
    return null
  }

  const status = data.status

  const getHost = () => {
    const defaultHost = '146.152.*.*'

    const address = data?.sshProxyAddress

    if (!address) return defaultHost

    const octets = address.split('.')

    if (octets.length !== 4) {
      return defaultHost
    }

    const maskedIP = `${octets[0]}.${octets[1]}.*.*`

    return maskedIP
  }

  const hosts = getHost()

  const getSshConfigProxyMessage = () => {
    return selectedValue === 'windows'
      ? `Host ${hosts}\r\nProxyCommand "C:\\Program Files\\Git\\mingw64\\bin\\connect.exe" -S internal-placeholder.com:1080 %h %p`
      : selectedValue === 'linux'
        ? `Host ${hosts}\nProxyCommand /usr/bin/nc -x internal-placeholder.com:1080 %h %p`
        : `Host ${hosts}\nProxyCommand /usr/bin/nc -X internal-placeholder.com:1080 %h %p`
  }

  const sshPermissonCommand = 'chmod 400 my-key.ssh'
  let sshConnectCommand = ''
  sshConnectCommand = `ssh -J ${data?.sshProxyUser}@${data?.sshProxyAddress} ${
    data?.userName ? data.userName : 'sdp'
  }@${data?.ipNbr}`

  return (
    <div className="section">
      {status === 'Provisioning' || status === 'Updating' || status === 'Starting' ? (
        <>
          We are currently awaiting the completion of instance initialization. Once the instance is ready, we will show
          you how to connect to your instance.
        </>
      ) : null}
      {status === 'Pending Review' ? (
        <>
          We are currently waiting for the instance request to be approved. Once the instance is ready, we will show you
          how to connect to your instance.
        </>
      ) : null}
      {status === 'Failed' ? <>The instance could not be initialized. Please try creating a new one.</> : null}
      {status === 'Stopped' || status === 'Stopping' ? (
        <>Currently this instance is not running. Start the instance to connect.</>
      ) : null}
      {status === 'Ready' || status === 'Active' ? (
        <>
          {isIntelUser() ? (
            <>
              <span className="lead">Select your OS:</span>
              <div className="d-flex flex-column gap-s4">
                <Form.Check
                  id="WindowsRadioButton"
                  intc-id="WindowsRadioButton"
                  label="Windows"
                  aria-label="Select Windows how to copy keys file to instance instructions"
                  type="radio"
                  value="Windows"
                  onChange={() => setSelectedValue('windows')}
                  checked={selectedValue === 'windows'}
                />
                <Form.Check
                  id="LinuxRadioButton"
                  intc-id="LinuxRadioButton"
                  label="Linux / macOS"
                  aria-label="Select Linux and Mac OS how to copy keys file to instance instructions"
                  type="radio"
                  value="Linux"
                  onChange={() => setSelectedValue('linux')}
                  checked={selectedValue === 'linux'}
                />
              </div>
            </>
          ) : null}
          <span>To access your instance with an SSH client:</span>
          <ol>
            {isIntelUser() ? (
              <li>
                {`From the Intel network: to connect to the ${idcConfig.REACT_APP_CONSOLE_SHORT_NAME} tenant SSH proxy, you need to configure the
                  SSH client to use the Intel SOCKS proxy. You can do this by adding the following configuration to
                  ~/.ssh/config.`}
                <CodeLine codeline={getSshConfigProxyMessage()} />
              </li>
            ) : null}
            <li>Open an SSH client.</li>
            <li>
              Locate your public key file (my-key.ssh). The wizard automatically detects the key you used to launch the
              instance.{' '}
            </li>
            <li>
              Your key must not be publicly visible for SSH to work. Use this command:
              <CodeLine codeline={sshPermissonCommand} />
            </li>
            <li>
              SSH command to connect to instance
              <CodeLine codeline={sshConnectCommand} />
            </li>
          </ol>
          <span className="valid-feedback w-100">
            Please note that in most cases the username above will be correct, however please ensure that you read your{' '}
            instance usage instructions to ensure that the instance owner has not changed the default instance username.
          </span>
        </>
      ) : null}
    </div>
  )
}

export default HowToConnectInstance
