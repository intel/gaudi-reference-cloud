// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import UserBadge from '../../utility/UserBadge/UserBadge'
import LabelValuePair from '../../utility/labelValuePair/LabelValuePair'
import Spinner from '../../utility/modals/spinner/Spinner'

const AccountSettingsView = (props: any): JSX.Element => {
  const user = props.user
  const email = user?.name
  const cloudAccountId = user?.id
  const userType = user?.type
  const loading = props.loading

  const spinner = <Spinner />
  const accountContent = (
    <>
      <div className="section">
        <span className="h4">Selected AI Cloud Account</span>
        <div className="d-flex flex-row flex-wrap gap-s8">
          <LabelValuePair label="Cloud Account ID">{cloudAccountId}</LabelValuePair>
          <div className="d-inline">
            <label className="fw-semibold" htmlFor="planTypeInfo">
              Tier:
            </label>
            <span className="fw-normal ms-s4" id="planTypeInfo">
              <UserBadge userType={userType} />
            </span>
          </div>
        </div>
      </div>
      <div className="section">
        <span className="h4">Selected intel.com Account</span>
        <LabelValuePair label="Email">{email}</LabelValuePair>
      </div>
    </>
  )
  return <>{loading ? spinner : accountContent}</>
}

export default AccountSettingsView
