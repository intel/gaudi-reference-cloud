// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import GetStarted from '../../components/getStarted/GetStarted'
import CloudAccountService from '../../services/CloudAccountService'
import { type GetStartedCard, initialCardState } from '../homePage/Home.types'
import useTrainingStore from '../../store/trainingStore/TrainingStore'

const GetStartedContainer = (): JSX.Element => {
  // local variables
  const servicePayload: any = {
    accessType: 'ACCESS_TYPE_JUPYTER',
    trainingId: null
  }

  // Navigation and query params
  const [searchParams] = useSearchParams()

  const tab: string = searchParams.get('tab') ?? 'learn'

  const [errorModal, setErrorModal] = useState({
    show: false,
    message: ''
  })
  const [launchModal, setLaunchModal] = useState({
    show: false,
    response: ''
  })

  // Store
  const trainings = useTrainingStore((state) => state.trainings)
  const setTrainings = useTrainingStore((state) => state.setTrainings)
  const loading = useTrainingStore((state) => state.loading)

  // useEffect
  useEffect(() => {
    const fetchTrainings = async (): Promise<void> => {
      await setTrainings()
    }
    if (trainings.length === 0 && tab === 'learn') fetchTrainings().catch(() => {})
  }, [tab])

  // functions
  async function onClickLaunch(id: string | null): Promise<void> {
    try {
      setLaunchModal({
        show: true,
        response: ''
      })
      const payload = { ...servicePayload }
      payload.trainingId = id
      const response = await CloudAccountService.postEnrollTraining(payload)
      const data = response.data
      setLaunchModal({
        show: true,
        response: data.jupyterLoginInfo
      })
    } catch (error: any) {
      const response = error.response
      const data = response.data
      const message = data.message
      setLaunchModal({
        show: false,
        response: ''
      })
      setErrorModal({
        show: true,
        message: message || error.message
      })
    }
  }

  function getTrainingId(name: string | null): string | null {
    if (!name) return null
    const result = trainings.find((training) => training.name === name)
    return result?.id ?? null
  }

  const openJupiterLab = async (trainingName: string): Promise<void> => {
    const trainingId = getTrainingId(trainingName)
    await onClickLaunch(trainingId)
  }

  let details: GetStartedCard | undefined

  switch (tab) {
    case 'learn':
      details = initialCardState.learn
      break
    case 'deploy':
      details = initialCardState.deploy
      break
    default:
      break
  }

  return (
    <GetStarted
      tab={tab}
      loading={loading}
      errorModal={errorModal}
      launchModal={launchModal}
      getStartedDetails={details}
      openJupiterLab={openJupiterLab}
      setErrorModal={setErrorModal}
      setLaunchModal={setLaunchModal}
    />
  )
}

export default GetStartedContainer
