// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import useUserStore from '../../store/userStore/UserStore'
import idcConfig from '../../config/configurator'
import loader from '../../assets/images/loader.svg'

const AuthenticationSpinner = (props) => {
  const isLogoutInProgress = useUserStore((state) => state.isLogoutInProgress)

  const Message = () => {
    if (isLogoutInProgress) {
      return <p>Logout in progress</p>
    }
    return null
  }

  return (
    <div className="d-flex flex-column align-items-center justify-content-center" style={{ height: '50vh' }}>
      <div className="box">
        <img src={loader} alt="" className="loader" />
      </div>
      <h3 className="p-3">{`${idcConfig.REACT_APP_CONSOLE_LONG_NAME}`}</h3>
      <Message />
    </div>
  )
}

export default AuthenticationSpinner
