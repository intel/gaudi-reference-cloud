// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { CloudAccountType } from '../Enums'

const UserBadge = (props: any): JSX.Element => {
  const userType = props.userType
  if (!userType) {
    return <></>
  }

  const getUserLabel = (): string => {
    switch (userType) {
      case CloudAccountType.intel:
        return 'Intel'
      case CloudAccountType.premium:
        return 'Premium'
      case CloudAccountType.standard:
        return 'Standard'
      case CloudAccountType.enterprise:
      case CloudAccountType.enterprise_pending:
        return 'Enterprise'
      default:
        return ''
    }
  }

  if (userType === CloudAccountType.enterprise || userType === CloudAccountType.enterprise_pending) {
    return (
      <>
        <span className={'badge header-badge header-badge-bg-enterprise text-capitalize'}>{getUserLabel()}</span>
        {userType === CloudAccountType.enterprise_pending ? (
          <span className={'badge header-badge header-badge-bg-enterprise_pending text-capitalize'}>
            Pending confirmation
          </span>
        ) : null}
      </>
    )
  }

  return (
    <span className={`badge header-badge header-badge-bg-${getUserLabel().toLowerCase()} text-capitalize`}>
      {getUserLabel()}
    </span>
  )
}

export default UserBadge
