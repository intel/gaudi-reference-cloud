// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useNavigate } from 'react-router'
import AccountsContainer from './AccountsContainer'
import useAppStore from '../../store/appStore/AppStore'

const AccountsContainerInRoute = () => {
  const navigate = useNavigate()
  const firstLoadedPage = useAppStore((state) => state.firstLoadedPage)
  const setFirstLoadedPage = useAppStore((state) => state.setFirstLoadedPage)

  const goBack = () => {
    if (firstLoadedPage === window.location.pathname) {
      setFirstLoadedPage(null)
      navigate({ pathname: '/home' })
    } else {
      navigate({ pathname: '/home' })
    }
  }

  return <AccountsContainer goBack={goBack} />
}

export default AccountsContainerInRoute
