// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import SecondUseDashboard from '../../../components/homePage/secondUseDashboard/SecondUseDashboard'
import { checkRoles } from '../../../utils/accessControlWrapper/AccessControlWrapper'
import { AppRolesEnum } from '../../../utils/Enums'
import { useNavigate } from 'react-router'
import useUserStore from '../../../store/userStore/UserStore'

const SecondUseDashboardContainer = (): JSX.Element => {
  const showCloudCredits = !checkRoles([AppRolesEnum.EnterprisePending])
  const isOwnCloudAccount = useUserStore((state) => state.isOwnCloudAccount)
  const navigate = useNavigate()

  const navigateTo = (url: string): void => {
    if (url) navigate({ pathname: url })
  }

  return (
    <SecondUseDashboard
      showCloudCredits={showCloudCredits}
      isOwnCloudAccount={isOwnCloudAccount}
      navigateTo={navigateTo}
    />
  )
}

export default SecondUseDashboardContainer
