// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useState, useEffect } from 'react'
import ApiKeysView from '../../components/profile/apiKeys/ApiKeysView'
import Wrapper from '../../utils/Wrapper'
import { getAccessToken, parseJwt } from '../../utils/AxiosInstance'
import moment from 'moment'
import { useCopy } from '../../hooks/useCopy'

const ApiKeysContainer = () => {
  const [token, setToken] = useState('')
  const [expirationDate, setExpirationDate] = useState('')
  const { copyToClipboard } = useCopy()

  const currentDateTimeFormat = 'M/D/YYYY HH:mm:ss'

  const refreshKey = async () => {
    const accessToken = await getAccessToken({
      forceRefresh: true
    })
    const { exp } = parseJwt(accessToken)
    const expirationDate = new Date(0)
    expirationDate.setUTCSeconds(Number(exp))
    setToken(accessToken)
    setExpirationDate(moment(expirationDate).format(currentDateTimeFormat))
  }

  const copyKey = () => {
    copyToClipboard(token)
  }

  useEffect(() => {
    refreshKey()
  }, [])

  return (
    <Wrapper>
      <ApiKeysView token={token} expirationDate={expirationDate} refreshKey={refreshKey} copyKey={copyKey} />
    </Wrapper>
  )
}

export default ApiKeysContainer
