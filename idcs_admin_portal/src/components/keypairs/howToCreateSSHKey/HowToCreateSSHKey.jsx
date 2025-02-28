import React, { useState } from 'react'
import Accordion from 'react-bootstrap/Accordion'
import { BsQuestionCircle } from 'react-icons/bs'
import HowToCreateSSHKeyWindows from './HowToCreateSSHKeyWindows'
import HowToCreateSSHKeyLinux from './HowToCreateSSHKeyLinux'

const HowToCreateSSHKey = (props) => {
  const [selectedValue, setSelectedValue] = useState('windows')

  return (
    <div className="row">
      <div className="col-md-12">
        <Accordion intc-id="howToCreateSSHKeyAccordion">
          <Accordion.Item eventKey="0">
            <Accordion.Header className=""><BsQuestionCircle className="me-2" size="16" /> How to create a SSH key</Accordion.Header>
            <Accordion.Body>
              <div className="border me-3 p-4" style={{ background: '#F8F9FA' }}>
                <div>
                  <h3 className="h4">Select your OS:</h3>
                  <input intc-id="WindowsRadioButton" aria-label="Select Windows how to create ssh instructions" type="radio" value="Windows" className="me-2" onChange={() => setSelectedValue('windows')} checked={selectedValue === 'windows'} /> Windows
                  <br/>
                  <input intc-id="LinuxRadioButton" aria-label="Select Linux and Mac OS how to create ssh instructions" type="radio" value="Linux" className="me-2" onChange={() => setSelectedValue('linux')} checked={selectedValue === 'linux'} /> Linux/MacOS
                  <br/>
                  { selectedValue === 'windows' ? <HowToCreateSSHKeyWindows/> : null }
                  { selectedValue === 'linux' ? <HowToCreateSSHKeyLinux/> : null }
                </div>
              </div>
            </Accordion.Body>
          </Accordion.Item>
        </Accordion>
      </div>
    </div>
  )
}

export default HowToCreateSSHKey
