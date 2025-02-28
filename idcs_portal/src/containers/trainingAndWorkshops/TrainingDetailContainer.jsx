// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState, useRef } from 'react'
import { useParams } from 'react-router-dom'
import useTrainingStore from '../../store/trainingStore/TrainingStore'
import TrainingDetail from '../../components/trainingAndWorkshops/trainingDetail/TrainingDetail'
import CloudAccountService from '../../services/CloudAccountService'
import moment from 'moment'
import useAppStore from '../../store/appStore/AppStore'

const TrainingDetailContainer = () => {
  // Params
  const { param } = useParams()

  // Local State
  const initialState = {
    servicePayloadSSH: {
      sshKeyNames: [],
      accessType: 'ACCESS_TYPE_SSH',
      trainingId: null
    },
    servicePayloadJupiter: {
      accessType: 'ACCESS_TYPE_JUPYTER',
      trainingId: null
    }
  }
  const noFoundTraining = {
    title: 'No learning found',
    subTitle: 'The page you are trying to access does not exist. \n You can go to any of the following links:',
    action: {
      type: 'redirect',
      btnType: 'link',
      href: '/learning/notebooks',
      label: 'Learning catalog'
    }
  }
  const stateRef = useRef(null)

  const [launchModal, setLaunchModal] = useState(false)
  const [jupiterLaunchModal, setJupiterLaunchModal] = useState(false)
  const [showErrorModal, setShowErrorModal] = useState(false)
  const [errorMessage, setErrorMessage] = useState(null)
  const [serviceResponse, setServiceResponse] = useState({
    jupyterLoginInfo: '',
    messge: '',
    expiryDate: ''
  })
  const [enrolling, setEnrolling] = useState(false)
  const [setIsSubmitted] = useState(false)
  const [enrollingJupiter, setEnrollingJupiter] = useState(false)
  const [jupiterExpiry, setJupiterExpiry] = useState('')
  const [jupyterRes, setJupyterRes] = useState(null)
  // Store
  const loading = useTrainingStore((state) => state.loading)
  const training = useTrainingStore((state) => state.trainingDetail)
  const getTraining = useTrainingStore((state) => state.getTraining)
  const addBreadcrumCustomTitle = useAppStore((state) => state.addBreadcrumCustomTitle)
  // Hooks
  useEffect(() => {
    const fetch = async () => {
      await getTraining(param)
    }
    fetch()
  }, [])

  useEffect(() => {
    if (training) {
      addBreadcrumCustomTitle(`/learning/notebooks/detail/${training.id}`, training.displayName)
    }
  }, [training])

  useEffect(() => {
    const fetch = async () => {
      try {
        const response = await CloudAccountService.getExpiry()
        const data = response.data
        setJupiterExpiry(moment(data.expiryDate).format('MM/DD/YYYY'))
      } catch (error) {
        setJupiterExpiry('Invalid date')
      }
    }
    fetch()
  }, [])

  async function onClickLaunch(show) {
    if (show) {
      try {
        if (stateRef.current.length > 0) {
          setLaunchModal(show)
          setErrorMessage(null)
          setEnrolling(true)
          const state = { ...initialState }
          const servicePayload = { ...state.servicePayloadSSH }
          servicePayload.sshKeyNames = stateRef.current.map((key) => key.value)
          servicePayload.trainingId = training.id
          const response = await CloudAccountService.postEnrollTraining(servicePayload)
          const data = response.data
          setServiceResponse({
            sshLoginInfo: data.sshLoginInfo,
            message: data.message,
            expiryDate: moment(data.expiryDate).format('MM/DD/YYYY')
          })
          setJupiterExpiry(moment(data.expiryDate).format('MM/DD/YYYY'))
          setEnrolling(false)
        } else {
          setIsSubmitted(true)
        }
      } catch (error) {
        setLaunchModal(false)
        setEnrolling(false)
        setIsSubmitted(false)
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
    } else {
      setLaunchModal(show)
    }
  }
  const onClickLaunchJupiter = async (e, show) => {
    try {
      if (!e) {
        return
      }
      e.preventDefault()
      if (show) {
        setJupiterLaunchModal(show)
        setEnrollingJupiter(true)
        const state = { ...initialState }
        const servicePayload = { ...state.servicePayloadJupiter }
        servicePayload.trainingId = training.id
        const response = await CloudAccountService.postEnrollTraining(servicePayload)
        const data = response.data
        setEnrollingJupiter(false)
        setJupyterRes(data.jupyterLoginInfo)
        setJupiterExpiry(moment(data.expiryDate).format('MM/DD/YYYY'))
      } else {
        setJupiterLaunchModal(show)
      }
    } catch (error) {
      setShowErrorModal(true)
      setEnrollingJupiter(false)
      setJupiterLaunchModal(false)
    }
  }

  const closeJupyterModal = (show) => {
    setJupiterLaunchModal(show)
  }

  return (
    <TrainingDetail
      loading={loading}
      launchModal={launchModal}
      showErrorModal={showErrorModal}
      enrolling={enrolling}
      errorMessage={errorMessage}
      serviceResponse={serviceResponse}
      jupiterExpiry={jupiterExpiry}
      enrollingJupiter={enrollingJupiter}
      setShowErrorModal={setShowErrorModal}
      onClickLaunch={onClickLaunch}
      closeJupyterModal={closeJupyterModal}
      onClickLaunchJupiter={onClickLaunchJupiter}
      jupyterRes={jupyterRes}
      training={training}
      jupiterLaunchModal={jupiterLaunchModal}
      noFoundTraining={noFoundTraining}
    />
  )
}

export default TrainingDetailContainer
