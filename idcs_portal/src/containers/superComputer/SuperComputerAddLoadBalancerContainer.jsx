// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import SuperComputerAddLoadBalancer from '../../components/superComputer/superComputerAddLoadBalancer/SuperComputerAddLoadBalancer'
import { useNavigate, useParams } from 'react-router'
import useSuperComputerStore from '../../store/superComputer/SuperComputerStore'
import {
  UpdateFormHelper,
  isValidForm,
  getFormValue,
  showFormRequiredFields
} from '../../utils/updateFormHelper/UpdateFormHelper'
import SuperComputerService from '../../services/SuperComputerService'
import useToastStore from '../../store/toastStore/ToastStore'
import { toastMessageEnum } from '../../utils/Enums'
import useErrorBoundary from '../../hooks/useErrorBoundary'

const getCustomMessage = (messageType) => {
  let message = null

  switch (messageType) {
    case 'name':
      message = (
        <div className="valid-feedback" intc-id={'LoadBalancerDetailsValidMessage'}>
          Max lenght 63 characters. Letters, numbers and ‘- ‘ accepted. Name should start and end with an alphanumeric
          character.
          <br />
        </div>
      )
      break
    case 'port':
      message = (
        <div className="valid-feedback" intc-id={'InstanceDetailsValidMessage'}>
          Available options: 80(HTTP), 443(HTTPS).
          <br />
        </div>
      )
      break
    default:
      break
  }

  return message
}

const SuperComputerAddLoadBalancerContainer = () => {
  // *****
  // Global state
  // *****
  const loadingDetail = useSuperComputerStore((state) => state.loadingDetail)
  const clusterDetail = useSuperComputerStore((state) => state.clusterDetail)
  const setClusterDetail = useSuperComputerStore((state) => state.setClusterDetail)
  const setDebounceDetailRefresh = useSuperComputerStore((state) => state.setDebounceDetailRefresh)
  const showError = useToastStore((state) => state.showError)
  // *****
  // local state
  // *****
  const navigate = useNavigate()
  const { param: name } = useParams()

  const initialState = {
    mainTitle: `Add load balancer to cluster ${name}`,
    mainSubtitle: 'Specify the needed Information',
    form: {
      loadbalancerName: {
        sectionGroup: 'lb',
        type: 'text', // options = 'text ,'textArea'
        label: 'Name:',
        placeholder: 'Load Balancer Name',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 63,
        validationRules: {
          isRequired: true,
          onlyAlphaNumLower: true,
          checkMaxLength: true
        },
        validationMessage: '', // Errror message to display to the user
        helperMessage: getCustomMessage('name'),
        columnSize: '7'
      },
      loadbalancerPort: {
        sectionGroup: 'lb',
        type: 'dropdown', // options = 'text ,'textArea'
        label: 'Port:',
        placeholder: 'Please select load balancer port',
        value: '80',
        isValid: true,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [
          { key: 1, name: '80', value: '80' },
          { key: 2, name: '443', value: '443' }
        ],
        validationMessage: '',
        helperMessage: getCustomMessage('port'),
        columnSize: '6'
      },
      loadbalancerType: {
        sectionGroup: 'lb',
        type: 'dropdown', // options = 'text ,'textArea'
        label: 'Type:',
        placeholder: 'Please select load balancer type',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [
          { key: 1, name: 'private', value: 'private' },
          { key: 2, name: 'public', value: 'public' }
        ],
        validationMessage: '',
        helperMessage: getCustomMessage('type'),
        columnSize: '6'
      }
    },
    isValidForm: false,
    navigationBottom: [
      {
        buttonLabel: 'Launch',
        buttonVariant: 'primary'
      },
      {
        buttonLabel: 'Cancel',
        buttonVariant: 'link',
        buttonFunction: () => onCancel()
      }
    ]
  }

  const servicePayload = {
    name: '',
    description: '',
    port: '',
    viptype: ''
  }

  const reservationModalInitial = {
    show: false,
    instanceType: 'clusterLoadBalancer'
  }

  const errorModalInitial = {
    show: false,
    message: '',
    title: 'Could not launch your vip',
    description: 'There was an error while processing your request'
  }

  const [state, setState] = useState(initialState)
  const [reservationModal, setReservationModal] = useState(reservationModalInitial)
  const [errorModal, setErrorModal] = useState(errorModalInitial)
  const [isPageReady, setIsPageReady] = useState(clusterDetail && clusterDetail.name === name)

  const throwError = useErrorBoundary()

  const fetchClusterDetail = async (isBackground) => {
    try {
      await setClusterDetail(name, isBackground)
      setIsPageReady(true)
    } catch (error) {
      throwError(error)
    }
  }

  // *****
  // use effect
  // *****

  useEffect(() => {
    if (!isPageReady) {
      fetchClusterDetail(false)
    }
  }, [])

  useEffect(() => {
    if (isPageReady && clusterDetail === null) {
      navigate('/supercomputer')
    }
  }, [clusterDetail, isPageReady])

  // *****
  // functions
  // *****
  function goBack() {
    navigate({
      pathname: `/supercomputer/d/${name}`,
      search: 'tab=loadBalancers'
    })
  }

  function onCancel() {
    // Navigates back to the page when this method triggers.
    goBack()
  }

  function onChangeInput(event, formInputName) {
    const updatedState = {
      ...state
    }

    const updatedForm = UpdateFormHelper(event.target.value, formInputName, updatedState.form)

    updatedState.isValidForm = isValidForm(updatedForm)

    updatedState.form = updatedForm

    setState(updatedState)
  }

  function showRequiredFields() {
    const stateCopy = { ...state }
    // Mark regular Inputs
    const updatedForm = showFormRequiredFields(stateCopy.form)
    // Create toast
    showError(toastMessageEnum.formValidationError, false)
    stateCopy.form = updatedForm
    setState(stateCopy)
  }

  function onSubmit() {
    const isValidForm = state.isValidForm
    if (!isValidForm) {
      showRequiredFields()
      return
    }
    setReservationModal({ ...reservationModal, show: true })
    const payloadCopy = { ...servicePayload }
    payloadCopy.name = getFormValue('loadbalancerName', state.form)
    payloadCopy.port = getFormValue('loadbalancerPort', state.form)
    payloadCopy.viptype = getFormValue('loadbalancerType', state.form)
    createILB(payloadCopy)
  }
  async function createILB(payload) {
    try {
      await SuperComputerService.createLoadBalancer(payload, clusterDetail.uuid)
      setDebounceDetailRefresh(true)
      setReservationModal(reservationModalInitial)
      goBack()
    } catch (error) {
      setReservationModal(reservationModalInitial)
      let message = ''
      if (error.response) {
        message = error.response.data.message
      } else {
        message = error.message
      }
      setErrorModal({ ...errorModal, show: true, message })
    }
  }

  function onCloseErrorModal() {
    setErrorModal(errorModalInitial)
  }

  return (
    <SuperComputerAddLoadBalancer
      loading={loadingDetail || !isPageReady}
      navigationBottom={state.navigationBottom}
      setState={setState}
      form={state.form}
      onChangeInput={onChangeInput}
      isValidForm={state.isValidForm}
      mainTitle={state.mainTitle}
      mainSubtitle={state.mainSubtitle}
      onSubmit={onSubmit}
      reservationModal={reservationModal}
      errorModal={errorModal}
      onCloseErrorModal={onCloseErrorModal}
    />
  )
}

export default SuperComputerAddLoadBalancerContainer
