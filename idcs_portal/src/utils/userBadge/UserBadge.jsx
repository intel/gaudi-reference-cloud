// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import useUserStore from '../../store/userStore/UserStore'
import { EnrollAccountType } from '../Enums'

const UserBadge = () => {
  const user = useUserStore((state) => state.user)

  if (!user) {
    return null
  }

  const getUserLabel = () => {
    switch (user.cloudAccountType) {
      case EnrollAccountType.intel:
        return 'Intel'
      case EnrollAccountType.premium:
        return 'Premium'
      case EnrollAccountType.standard:
        return 'Standard'
      case EnrollAccountType.enterprise:
      case EnrollAccountType.enterprise_pending:
        return 'Enterprise'
      default:
        return ''
    }
  }

  if (
    user.cloudAccountType === EnrollAccountType.enterprise ||
    user.cloudAccountType === EnrollAccountType.enterprise_pending
  ) {
    return (
      <>
        <span className={'badge header-badge header-badge-bg-enterprise text-capitalize'}>{getUserLabel()}</span>
        {user.cloudAccountType === EnrollAccountType.enterprise_pending ? (
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
