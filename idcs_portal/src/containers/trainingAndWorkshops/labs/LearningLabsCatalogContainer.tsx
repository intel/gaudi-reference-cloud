// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import LearningLabsCatalog from '../../../components/learningLabs/catalog/LearningLabsCatalog'
import useLearningLabsStore from '../../../store/learningLabsStore/LearningLabsStore'
import useErrorBoundary from '../../../hooks/useErrorBoundary'
import { appFeatureFlags, isFeatureFlagEnable } from '../../../config/configurator'

const LearningLabsCatalogContainer = (): JSX.Element => {
  // Error handle
  const throwError = useErrorBoundary()

  // Local State
  const comingMessage = 'Get ready for labs'
  const isAvailable = isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_LEARNING_LABS)
  // Store
  const learningLabsList = useLearningLabsStore((state) => state.learningLabsList)
  const setLearningLabsList = useLearningLabsStore((state) => state.setLearningLabsList)
  const loading = useLearningLabsStore((state) => state.loading)
  const [searchFilter, setSearchFilter] = useState({})
  const [textFilter, setTextFilter] = useState('')

  // Hooks
  useEffect(() => {
    if (isAvailable) {
      getLearningLabsInfo()
    }
  }, [])

  function getLearningLabsInfo(): void {
    const fetch = async (): Promise<void> => {
      try {
        await setLearningLabsList()
      } catch (error: any) {
        let errorMessage = ''
        let errorCode = -1
        let errorStatus = -1
        const isApiErrorWithErrorMessage = Boolean(error?.response?.data?.message)
        if (isApiErrorWithErrorMessage) {
          errorMessage = error.response.data.message
          errorCode = error.response.data.code
          errorStatus = error.response.status
        } else {
          errorMessage = error.toString()
        }

        if (errorStatus === 403 && errorCode === 7 && errorMessage.toLowerCase().includes('user is restricted')) {
          throwError(error)
        }
      }
    }
    fetch().catch((error) => {
      throwError(error)
    })
  }

  function setTagFilter(key: string, values: string[]): void {
    const updatedFilter: any = { ...searchFilter }
    updatedFilter[key] = values
    setSearchFilter(updatedFilter)
  }

  return (
    <LearningLabsCatalog
      loading={loading}
      isAvailable={isAvailable}
      comingMessage={comingMessage}
      learningLabsList={learningLabsList}
      setTagFilter={setTagFilter}
      textFilter={textFilter}
      setTextFilter={setTextFilter}
      searchFilter={searchFilter}
    />
  )
}

export default LearningLabsCatalogContainer
