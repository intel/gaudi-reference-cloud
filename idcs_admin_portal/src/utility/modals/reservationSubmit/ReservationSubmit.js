import React from 'react'
import Modal from 'react-bootstrap/Modal'
import { Link } from 'react-router-dom'

const ReservationSubmit = (props) => {
  const instanceType = props?.instanceType ? props?.instanceType : 'other'

  const messages = {
    storage: {
      m1: 'Storage Volume',
      m2: 'Storage'
    },
    other: {
      m1: 'instance',
      m2: 'instance'
    }
  }

  return (
        <Modal
            show={props.showReservationCreateModal}
            backdrop="static"
            keyboard={false}
            size="lg">
                <Modal.Header></Modal.Header>
            <Modal.Body>
                <div className="modal-body row justify-content-center">
                    <div className="col-6 row ">
                        <div className="spinner-border text-primary center"></div>
                        {instanceType === 'storage'
                          ? <>
                            <span className="col-12 pl-0 pt-2"><strong>Working on your volume creation</strong></span>
                            <small className="col-12 pl-0 pt-1 ">&nbsp; Provisioning volume</small>
                            <small className="col-12 pl-0 pt-0 ">&nbsp; Generating mounting details</small>
                        </>
                          : null}

                        {instanceType === 'cluster'
                          ? <>
                            <span className="col-12 pl-0 pt-2"><strong>Working on your new Cluster</strong></span>
                            <small className="col-12 pl-0 pt-1 ">&nbsp; Creating Nodes</small>
                            <small className="col-12 pl-0 pt-0 ">&nbsp; Provisioning Cluster</small>
                            <small className="col-12 pl-0 pt-0 ">&nbsp; Tip: Check cluster status, wait for Resource is Ready.</small>
                        </>
                          : null}

                        {instanceType === 'other'
                          ? <>
                            <span className="col-12 pl-0 pt-2"><strong>Working on your reservation</strong></span>
                            <small className="col-12 pl-0 pt-1 ">&nbsp; Reserving {messages[instanceType].m1}</small>
                            <small className="col-12 pl-0 pt-0 ">&nbsp; Confirming keys</small>
                            <small className="col-12 pl-0 pt-0 ">&nbsp; Adding {messages[instanceType].m2} to your reservation</small>
                        </>
                          : null}
                    </div>
                </div>
            </Modal.Body>
            <Modal.Footer>
                <Link to={props.redirectTo} style={{ textDecoration: 'none' }}>
                </Link>
            </Modal.Footer>
        </Modal>
  )
}

export default ReservationSubmit
