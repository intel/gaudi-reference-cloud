import { useState, useEffect } from 'react'
import moment from 'moment'
import { getAccessToken, parseJwt } from '../../utility/axios/AxiosInstance'
import ApiKeysView from '../../components/profile/apiKeys/ApiKeysView'

const ApiKeysContainer = () => {
  const [token, setToken] = useState('')
  const [expirationDate, setExpirationDate] = useState('')

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

  useEffect(() => {
    refreshKey()
  }, [])

  return (
    <>
      <ApiKeysView
        token={token}
        expirationDate={expirationDate}
        refreshKey={refreshKey}
      />
    </>
  )
}

export default ApiKeysContainer
