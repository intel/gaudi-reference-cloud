// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import Accordion from 'react-bootstrap/Accordion'
import CodeLine from '../../../CodeLine'
import HowToConnectInstance from './HowToConnectInstance'
import HotToConnectSelectInstance from './HotToConnectSelectInstance'
import idcConfig from '../../../../config/configurator'
import { Link } from 'react-router-dom'

const HowToConnectVastStorage = (props: any): JSX.Element => {
  const data = props.data
  const mount = props.mount
  const [selectedInstance, setSelectedInstance] = useState<any>(null)
  const [showMountDetails, setShowMountDetails] = useState(false)
  const status = data.status
  const mountDirectoryCommand = 'sudo mkdir /mnt/test'
  const action = mount ? 'mount' : 'unmount'
  const actionCapital = mount ? 'Mount' : 'Unmount'
  const fileSystemName = data.name
  const mountCommand = `sudo mount -o noresvport,vers=4.1,nconnect=16 localhost:/${fileSystemName} /mnt/test`
  const unmountCommand = 'sudo umount /mnt/test'
  useEffect(() => {
    if (selectedInstance) {
      selectedInstance.ipNbr = selectedInstance.interfaces[0].addresses[0]
      setShowMountDetails(true)
    }
  }, [selectedInstance])

  return (
    <>
      {status === 'Provisioning' ? (
        <>
          We are currently awaiting the completion of storage initialization. Once the storage is ready, we will show
          you how to connect to your storage.
        </>
      ) : null}
      {status === 'Failed' ? <>The storage could not be initialized. Please try creating a new one.</> : null}
      {status === 'Ready' && !selectedInstance ? (
        <HotToConnectSelectInstance setSelectedInstance={setSelectedInstance} />
      ) : null}
      {status === 'Ready' && selectedInstance && showMountDetails ? (
        <>
          <Accordion intc-id="howToConnectInstanceAccordion">
            <Accordion.Item eventKey="0">
              <Accordion.Header>
                <h2 className="h6">1. Connect to your instance:</h2>
              </Accordion.Header>
              <Accordion.Body>
                <HowToConnectInstance data={selectedInstance} />
              </Accordion.Body>
            </Accordion.Item>
          </Accordion>
          <Accordion intc-id="howToConnectStorageAccordion">
            <Accordion.Item eventKey="1">
              <Accordion.Header>
                <h2 className="h6">2. {actionCapital} your volume:</h2>
              </Accordion.Header>
              <Accordion.Body>
                <div className="section">
                  <span className="h6">To {action} your storage with an SSH client:</span>
                  <ol className="w-100">
                    {action === 'mount' ? (
                      <>
                        <li>
                          Create a mount directory
                          <CodeLine codeline={mountDirectoryCommand} />
                        </li>
                        <li>
                          Mount the filesystem
                          <CodeLine codeline={mountCommand} />
                        </li>
                      </>
                    ) : (
                      <>
                        <li>
                          Unmount the filesystem
                          <CodeLine codeline={unmountCommand} />
                        </li>
                      </>
                    )}
                  </ol>
                  <span className="valid-feedback">
                    If you need any assistance connecting to your instance, please see our{' '}
                    <Link to={idcConfig.REACT_APP_GUIDES_STORAGE_FILE_URL} target="_blank">
                      documentation
                    </Link>
                    .
                  </span>
                </div>
              </Accordion.Body>
            </Accordion.Item>
          </Accordion>
        </>
      ) : null}
    </>
  )
}

export default HowToConnectVastStorage
