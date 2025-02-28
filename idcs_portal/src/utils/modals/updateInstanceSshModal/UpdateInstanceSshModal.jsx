// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState } from 'react'
import Modal from 'react-bootstrap/Modal'
import Button from 'react-bootstrap/Button'
import Form from 'react-bootstrap/Form'
import ToastContainer from '../../toast/ToastContainer'
import Accordion from 'react-bootstrap/Accordion'
import Card from 'react-bootstrap/Card'
import CodeLine from '../../CodeLine'
import { BsDownload } from 'react-icons/bs'
import idcConfig from '../../../config/configurator'

const UpdateInstanceSsh = (props) => {
  const [selectedValue, setSelectedValue] = useState('windows')

  if (!props.data) {
    return null
  }

  const instanceData = props.data.instanceData || null
  const keysData = props.data.keysData || null

  const downloadKeysFile = () => {
    const fileData = new Blob([keysData.map((x) => x.sshPublicKey).join('\n')], { type: 'text/plain' })
    const url = URL.createObjectURL(fileData)
    const link = document.createElement('a')
    link.download = 'public_keys.pub'
    link.href = url
    link.click()
  }

  const getScpCommand = (os) => {
    const filePath = os === 'windows' ? '"$env:USERPROFILE\\Downloads\\public_keys.pub"' : '~/Downloads/public_keys.pub'
    let scpCommand = ''

    if (instanceData) {
      const ipNbr = instanceData.interfaces.length > 0 ? instanceData.interfaces[0].addresses[0] : null
      scpCommand = `scp -J ${instanceData?.sshProxyUser}@${instanceData?.sshProxyAddress} ${filePath} ${
        instanceData?.userName ? instanceData.userName : 'sdp'
      }@${ipNbr}:~/`
    }
    return scpCommand
  }

  const getLauchShellCommand = (os) => {
    return os === 'windows'
      ? '1. Launch a new PowerShell window on your local system.'
      : '1. Launch a terminal on your local system.'
  }

  return (
    <Modal
      show={props.showUpdateInstanceSshModal}
      onHide={() => props.onCloseUpdateInstanceSsh(false)}
      backdrop="static"
      size="xl"
      aria-labelledby="contained-modal-title-vcenter"
      centered
      aria-label="Update instance ssh modal"
    >
      <ToastContainer />
      <Modal.Header closeButton>
        <Modal.Title>
          <h1 className="h4">Edit keys</h1>
        </Modal.Title>
      </Modal.Header>
      <Modal.Body>
        Enable your new keys by completing the following steps
        <Accordion defaultActiveKey="1" intc-id="copySshToInstanceCommandAccordion">
          <Card className="accordion-card">
            <Card.Body className="d-flex py-2 justify-content-between">
              <Card.Text className="my-auto px-1">1. Download the new keys file</Card.Text>
              <Button variant="outline-primary" onClick={() => downloadKeysFile()} intc-id="downloadSshFileButton">
                <BsDownload />
                &nbsp;Download
              </Button>
            </Card.Body>
          </Card>
          <Accordion.Item eventKey="1">
            <Accordion.Header className="">2. Add the new keys to your instance</Accordion.Header>
            <Accordion.Body className="section">
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
              <span>{getLauchShellCommand(selectedValue)}</span>
              <span>
                2. Copy & paste the following to your terminal to transfer the downloaded file to your instance.
              </span>
              <CodeLine codeline={getScpCommand(selectedValue)} />
              <span>3. If prompted, enter your SSH passphrase.</span>
              <span>
                4. Connect to your instance using the original key. View{' '}
                <a
                  rel="noreferrer"
                  href={idcConfig.REACT_APP_INSTANCE_CONNECT}
                  className="text-decoration-none"
                  target="_blank"
                >
                  how to connect in our documentation
                </a>
                &nbsp;for guidance.
              </span>
              <span>5. Copy & paste the following to your terminal to add the new keys to your instance.</span>
              <CodeLine codeline={'cat public_keys.pub >> ~/.ssh/authorized_keys'} />
              <span className="mb-0">
                For more information go to{' '}
                <a
                  rel="noreferrer"
                  href={idcConfig.REACT_APP_SHH_KEYS}
                  className="text-decoration-none"
                  target="_blank"
                >
                  SSH key documentation
                </a>
                .
              </span>
            </Accordion.Body>
          </Accordion.Item>
        </Accordion>
      </Modal.Body>
      <Modal.Footer>
        <Button
          variant="outline-primary"
          onClick={() => props.onCloseUpdateInstanceSsh(false)}
          intc-id="UpdateInstanceSshClose"
        >
          Close
        </Button>
      </Modal.Footer>
    </Modal>
  )
}

export default UpdateInstanceSsh
