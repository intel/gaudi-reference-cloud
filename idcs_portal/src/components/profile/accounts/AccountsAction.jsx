// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { X } from 'react-bootstrap-icons'
import { BsCheck2 } from 'react-icons/bs'

const AccountsAction = ({ onClickReject, onClickAccept }) => {
  return (
    <div className="d-flex flex-row justify-content-start">
      <a
        className="btn btn-success btn-sm me-1 shadow-none"
        intc-id="btn-accounts-accept"
        aria-label="Open accept invitation modal"
        onClick={onClickAccept}
      >
        <BsCheck2 />
        Accept
      </a>
      <a
        className="btn btn-danger btn-sm me-1 shadow-none"
        intc-id="btn-accounts-decline"
        aria-label="Open reject invitation modal"
        onClick={onClickReject}
      >
        <X />
        Decline
      </a>
    </div>
  )
}

export default AccountsAction
