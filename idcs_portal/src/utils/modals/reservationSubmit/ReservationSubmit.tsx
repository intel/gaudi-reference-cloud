// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Modal from 'react-bootstrap/Modal'
import Spinner from '../../spinner/Spinner'

interface ReservationSubmitProps {
  showReservationCreateModal: boolean
}

const ReservationSubmit: React.FC<ReservationSubmitProps> = (props): JSX.Element => {
  return (
    <Modal
      show={props.showReservationCreateModal}
      backdrop="static"
      keyboard={false}
      size="sm"
      centered
      aria-label="Reservation modal"
    >
      <Modal.Body className="py-s7 px-s0">
        <div className="d-flex flex-column p-s6 justify-content-center align-items-center gap-s8 align-self-stretch">
          <Spinner className="p-0" />
          <div className="d-flex flex-column align-items-center justify-content-center gap-s3">
            <h2 className="h5">Please wait</h2>
            <span>Working on your reservation</span>
          </div>
        </div>
      </Modal.Body>
    </Modal>
  )
}

export default ReservationSubmit
