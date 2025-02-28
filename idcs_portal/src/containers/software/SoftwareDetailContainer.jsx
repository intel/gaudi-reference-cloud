// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import useSoftwareStore from '../../store/SoftwareStore/SoftwareStore'
import { useNavigate } from 'react-router'
import CloudAccountService from '../../services/CloudAccountService'
import { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'
import useAppStore from '../../store/appStore/AppStore'
import SoftwareDetailThirdParty from '../../components/software/softwareDetail/SoftwareDetailThirdParty'
import { ReactComponent as SeekerLogo } from '../../assets/images/SeekrLogo.svg'
import GetiLogoSource from '../../assets/images/geti-logo.png'
import PredictionGuardLogoSource from '../../assets/images/prediction-guard-logo.png'

const SoftwareDetailContainer = () => {
  // Params
  const { param: id } = useParams()

  // Navigation
  const navigate = useNavigate()

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
  const noFoundSoftware = {
    title: 'No software found',
    subTitle: 'The page you are trying to access does not exist. \n You can go to any of the following links:',
    action: {
      type: 'redirect',
      btnType: 'link',
      href: '/software',
      label: 'Software catalog'
    }
  }
  const comingMessage =
    'Get ready for Intel-optimized software stacks hosted on performance optimized Intel compute platforms'
  const softwareLogosAsTitle = ['sw-intc-geti', 'sw-prediction-guard']
  const isAvailable = isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_SOFTWARE)
  const [launchModal, setLaunchModal] = useState(false)
  const [jupyterRes, setJupyterRes] = useState(null)
  const [errorModal, setErrorModal] = useState({
    show: false,
    message: ''
  })
  const [activeTab, setActiveTab] = useState(0)
  // Store
  const loading = useSoftwareStore((state) => state.loading)
  const softwareDetail = useSoftwareStore((state) => state.softwareDetail)
  const getSoftware = useSoftwareStore((state) => state.getSoftware)
  const addBreadcrumCustomTitle = useAppStore((state) => state.addBreadcrumCustomTitle)

  // Hooks
  useEffect(() => {
    if (isAvailable) {
      getSoftwareInfo()
    }
  }, [])

  useEffect(() => {
    if (softwareDetail) {
      addBreadcrumCustomTitle(`/software/d/${id}`, softwareDetail.displayName)
    }
  }, [softwareDetail])

  function getSoftwareInfo() {
    const fetch = async () => {
      const fetch = async () => {
        await getSoftware(id)
      }
      fetch()
    }
    fetch()
  }

  function GetiLogo(props) {
    return <img src={GetiLogoSource} alt="Intel Geti logo" {...props} />
  }

  function PredictionGuardLogo(props) {
    return <img src={PredictionGuardLogoSource} alt="Prediction Guard logo" {...props} />
  }

  function getThirdPartyLogo(softwareName) {
    switch (softwareName) {
      case 'sw-seekr-flow':
        return <SeekerLogo />
      case 'sw-intc-geti':
        return <GetiLogo />
      case 'sw-prediction-guard':
        return <PredictionGuardLogo />
    }
  }

  async function onClickLaunch(type) {
    if (type === 'image') {
      navigate({
        pathname: `/software/d/${id}/launch`
      })
    } else {
      try {
        setLaunchModal(true)
        const state = { ...initialState }
        const servicePayload = { ...state.servicePayloadJupiter }
        servicePayload.trainingId = softwareDetail.id
        const response = await CloudAccountService.postEnrollTraining(servicePayload)
        const data = response.data
        setJupyterRes(data.jupyterLoginInfo)
        window.open(data.jupyterLoginInfo, '_blank', 'noreferrer')
      } catch (error) {
        setLaunchModal(false)
        const response = error.response
        const data = response.data
        const message = data.message
        if (message) {
          setErrorModal({
            show: true,
            message
          })
        } else {
          setErrorModal({
            show: true,
            message: error.message
          })
        }
      }
    }
  }

  return (
    <SoftwareDetailThirdParty
      noFoundSoftware={noFoundSoftware}
      softwareDetail={softwareDetail}
      comingMessage={comingMessage}
      isAvailable={isAvailable}
      loading={loading}
      launchModal={launchModal}
      jupyterRes={jupyterRes}
      errorModal={errorModal}
      setErrorModal={setErrorModal}
      setLaunchModal={setLaunchModal}
      activeTab={activeTab}
      setActiveTab={setActiveTab}
      softwareLogo={getThirdPartyLogo(softwareDetail?.name)}
      softwareLogoAsTitle={softwareLogosAsTitle.includes(softwareDetail?.name)}
      onClickLaunch={onClickLaunch}
    />
  )
}

export default SoftwareDetailContainer
