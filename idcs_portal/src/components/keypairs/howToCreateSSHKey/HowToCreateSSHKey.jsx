// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState } from 'react'
import Form from 'react-bootstrap/Form'
import Accordion from 'react-bootstrap/Accordion'
import HowToCreateSSHKeyWindows from './HowToCreateSSHKeyWindows'
import HowToCreateSSHKeyLinux from './HowToCreateSSHKeyLinux'

const HowToCreateSSHKey = (props) => {
  const [selectedValue, setSelectedValue] = useState('windows')

  return (
    <>
      <Accordion intc-id="howToCreateSSHKeyAccordion">
        <Accordion.Item eventKey="0">
          <Accordion.Header>
            <h1 className="h4">How to create a SSH key</h1>
          </Accordion.Header>
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
            {selectedValue === 'windows' ? <HowToCreateSSHKeyWindows /> : null}
            {selectedValue === 'linux' ? <HowToCreateSSHKeyLinux /> : null}
          </Accordion.Body>
        </Accordion.Item>
      </Accordion>
    </>
  )
}

export default HowToCreateSSHKey
