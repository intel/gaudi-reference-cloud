// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useNavigate } from 'react-router'
import AccountSettingsView from '../../components/profile/accountSettings/AccountSettingsView'
import useUserStore from '../../store/userStore/UserStore'
import { EnrollAccountType } from '../../utils/Enums'
import { useCopy } from '../../hooks/useCopy'

const AccountSettingsContainer = () => {
  // Global State
  const user = useUserStore((state) => state.user)
  const isOwnCloudAccount = useUserStore((state) => state.isOwnCloudAccount)

  // Navigation && copy
  const navigate = useNavigate()
  const { copyToClipboard } = useCopy()

  const upgrade = () => {
    if (user.ownCloudAccountType === EnrollAccountType.standard) {
      navigate({
        pathname: '/upgradeaccount'
      })
    }
  }

  return (
    <AccountSettingsView
      displayName={`${user.lastName} ${user.firstName}`}
      cloudAccountId={user.cloudAccountNumber}
      email={user.email}
      cloudAccountType={user.cloudAccountType}
      upgrade={upgrade}
      isOwnCloudAccount={isOwnCloudAccount}
      copyToClipboard={copyToClipboard}
    />
  )
}

export default AccountSettingsContainer
