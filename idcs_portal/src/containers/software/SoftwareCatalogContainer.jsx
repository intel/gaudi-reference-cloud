// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import SoftwareCatalog from '../../components/software/catalog/SoftwareCatalog'
import useSoftwareStore from '../../store/SoftwareStore/SoftwareStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'

const SoftwareCatalogContainer = () => {
  // Local state
  const [softwareTextFilter, setSoftwareTextFilter] = useState('')
  const [searchFilter, setSearchFilter] = useState([])
  // Error handle
  const throwError = useErrorBoundary()

  // Local State
  const comingMessage =
    'Get ready for Intel-optimized software stacks hosted on performance optimized Intel compute platforms'
  const isAvailable = isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_SOFTWARE)
  // Store
  const softwareList = useSoftwareStore((state) => state.softwareList)
  const setSoftwareList = useSoftwareStore((state) => state.setSoftwareList)
  const loading = useSoftwareStore((state) => state.loading)

  // Hooks
  useEffect(() => {
    if (isAvailable) {
      getSoftwareInfo()
    }
  }, [])

  // functions
  function setTagFilter(key, values) {
    const updatedFilter = { ...searchFilter }
    updatedFilter[key] = values
    setSearchFilter(updatedFilter)
  }

  function getSoftwareInfo() {
    const fetch = async () => {
      try {
        await setSoftwareList(softwareList.length > 0)
      } catch (error) {
        let errorMessage = ''
        let errorCode = ''
        let errorStatus = -1
        const isApiErrorWithErrorMessage = Boolean(error.response && error.response.data && error.response.data.message)
        if (isApiErrorWithErrorMessage) {
          errorMessage = error.response.data.message
          errorCode = error.response.data.code
          errorStatus = error.response.status
        } else {
          errorMessage = error.toString()
        }

        if (errorStatus === 403 && errorCode === 7 && errorMessage.toLowerCase().indexOf('user is restricted') !== -1) {
          throwError(error)
        }
      }
    }
    fetch()
  }

  return (
    <SoftwareCatalog
      comingMessage={comingMessage}
      isAvailable={isAvailable}
      softwareList={softwareList}
      loading={loading}
      setTagFilter={setTagFilter}
      searchFilter={searchFilter}
      softwareTextFilter={softwareTextFilter}
      setSoftwareTextFilter={setSoftwareTextFilter}
    />
  )
}

export default SoftwareCatalogContainer
