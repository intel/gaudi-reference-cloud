// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import TrainingAndWorkshops from '../../components/trainingAndWorkshops/catalog/TrainingAndWorkshops'
import useTrainingStore from '../../store/trainingStore/TrainingStore'
import CloudAccountService from '../../services/CloudAccountService'
import useErrorBoundary from '../../hooks/useErrorBoundary'

const TrainingAndWorkshopsContainer = () => {
  // Local state
  const [trainingTextFilter, setTrainingTextFilter] = useState('')
  const [isTrainingServiceAvailable, setIsTrainingServiceAvailable] = useState(true)
  const [searchFilter, setSearchFilter] = useState([])
  const [enrolling, setEnrolling] = useState(false)
  const servicePayload = {
    accessType: 'ACCESS_TYPE_JUPYTER',
    trainingId: null
  }
  const [showErrorModal, setShowErrorModal] = useState(false)
  const [errorMessage, setErrorMessage] = useState(null)
  const [jupyterRes, setJupyterRes] = useState(null)
  // Store
  const trainings = useTrainingStore((state) => state.trainings)
  const setTrainings = useTrainingStore((state) => state.setTrainings)
  const loading = useTrainingStore((state) => state.loading)

  // Error handle
  const throwError = useErrorBoundary()

  // Hooks
  useEffect(() => {
    const fetch = async () => {
      try {
        await setTrainings(trainings.length > 0)
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
        setIsTrainingServiceAvailable(false)
      }
    }
    fetch()
  }, [])

  // functions
  function setTagFilter(key, values) {
    const updatedFilter = { ...searchFilter }
    updatedFilter[key] = values
    setSearchFilter(updatedFilter)
  }

  async function onClickLaunch(e, trainingId) {
    try {
      e.preventDefault()
      setEnrolling(true)
      setJupyterRes(null)
      const payload = { ...servicePayload }
      payload.trainingId = trainingId
      const response = await CloudAccountService.postEnrollTraining(payload)
      const data = response.data
      setJupyterRes(data.jupyterLoginInfo)
    } catch (error) {
      setEnrolling(false)
      setShowErrorModal(true)
      const response = error.response
      const data = response.data
      const message = data.message
      if (message) {
        setErrorMessage(message)
      } else {
        setErrorMessage(error.message)
      }
    }
  }

  return (
    <TrainingAndWorkshops
      trainings={trainings}
      isTrainingServiceAvailable={isTrainingServiceAvailable}
      loading={loading}
      trainingTextFilter={trainingTextFilter}
      searchFilter={searchFilter}
      enrolling={enrolling}
      jupyterRes={jupyterRes}
      setEnrolling={setEnrolling}
      onClickLaunch={onClickLaunch}
      errorMessage={errorMessage}
      showErrorModal={showErrorModal}
      setShowErrorModal={setShowErrorModal}
      setTagFilter={setTagFilter}
      setTrainingTextFilter={setTrainingTextFilter}
    />
  )
}

export default TrainingAndWorkshopsContainer
