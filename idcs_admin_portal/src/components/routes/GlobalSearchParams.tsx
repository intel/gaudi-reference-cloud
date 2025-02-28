// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import { useLocation, useSearchParams } from 'react-router-dom'
import idcConfig from '../../config/configurator'

const GlobalSearchParams = (): JSX.Element => {
  const { search } = useLocation()
  const [searchParams, setSearchParams] = useSearchParams()
  const [currentRegion, setCurrentRegion] = useState(searchParams.get('region'))

  useEffect(() => {
    if (currentRegion !== idcConfig.REACT_APP_SELECTED_REGION || !searchParams.get('region')) {
      setCurrentRegion(idcConfig.REACT_APP_SELECTED_REGION)
      setSearchParams(
        (params) => {
          params.set('region', idcConfig.REACT_APP_SELECTED_REGION)
          return params
        },
        { replace: true }
      )
    }
  }, [search, idcConfig.REACT_APP_SELECTED_REGION])

  return <></>
}
export default GlobalSearchParams
