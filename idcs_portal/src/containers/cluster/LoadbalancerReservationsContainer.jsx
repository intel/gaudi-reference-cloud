// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState, useEffect } from 'react'
import LoadbalancerReservations from '../../components/cluster/loadbalancerReservations/LoadbalancerReservations'
import { UpdateFormHelper, isValidForm, showFormRequiredFields } from '../../utils/updateFormHelper/UpdateFormHelper'
import { useNavigate, useParams } from 'react-router'
import ClusterService from '../../services/ClusterService'
import useClusterStore from '../../store/clusterStore/ClusterStore'
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

const LoadbalancerReservationsContainer = () => {
  // Global state
  const loading = useClusterStore((state) => state.loading)
  const clusters = useClusterStore((state) => state.clustersData)
  const setClusters = useClusterStore((state) => state.setClustersData)
  const showError = useToastStore((state) => state.showError)
  const navigate = useNavigate()
  const { param: name } = useParams()
  // local state
  const initialState = {
    mainTitle: `Add load balancer to cluster ${name}`,
    mainSubtitle: 'Specify the needed Information',
    errorDescription: 'There was an error while processing your Load Balancer',
    titleMessage: 'Could not launch your Load Balancer',
    instanceType: 'clusterLoadBalancer',
    timeoutMiliseconds: 4000,
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
    ],
    initialForm: {
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
        helperMessage: getCustomMessage('name')
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
        helperMessage: getCustomMessage('port')
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
        helperMessage: getCustomMessage('type')
      }
    }
  }

  const [isValid, setIsValid] = useState(false)

  const [form, setForm] = useState(initialState.initialForm)
  const [showReservationModal, setShowReservationModal] = useState(false)
  const [showErrorModal, setShowErrorModal] = useState(false)
  const [errorMessage, setErrorMessage] = useState('')
  const [clusterToEdit, setClusterToEdit] = useState(null)
  const [isPageReady, setIsPageReady] = useState(false)

  const throwError = useErrorBoundary()

  // functions
  const fetchClusters = async (isBackground) => {
    try {
      await setClusters(isBackground)
      setIsPageReady(true)
    } catch (error) {
      throwError(error)
    }
  }

  function goBack() {
    navigate({
      pathname: `/cluster/d/${name}`,
      search: 'tab=loadBalancers'
    })
  }

  function onCancel() {
    goBack()
  }

  function onChangeInput(event, formInputName) {
    let value = null
    value = event.target.value
    const formCopy = {
      ...form
    }

    const updatedForm = UpdateFormHelper(value, formInputName, formCopy)

    const isValid = isValidForm(updatedForm)

    setForm(updatedForm)

    setIsValid(isValid)
  }

  function showRequiredFields() {
    let formCopy = { ...form }
    // Mark regular Inputs
    const updatedForm = showFormRequiredFields(formCopy)
    // Create toast
    showError(toastMessageEnum.formValidationError, false)
    formCopy = updatedForm
    setForm(formCopy)
  }

  async function onSubmit() {
    try {
      if (!isValid) {
        showRequiredFields()
        return
      }
      await createILB()
    } catch (error) {
      setShowErrorModal(true)
      setShowReservationModal(false)

      if (error.response) {
        setErrorMessage(error.response.data.message)
      } else {
        setErrorMessage(error.message)
      }
    }
  }
  async function createILB() {
    try {
      const ok = [200, 201]
      const response = await ClusterService.createLoadBalancer(form, clusterToEdit.uuid)
      setTimeout(() => {
        fetchClusters(false)
        goBack()
      }, initialState.timeoutMiliseconds)
      if (ok.includes(response.status)) {
        setShowErrorModal(false)
        setShowReservationModal(true)

        return response
      }
      if (response.code) {
        const message = response.message ? response.message : ''
        setErrorMessage(message)
        setShowErrorModal(true)
        setShowReservationModal(false)
      }
      return response
    } catch (error) {
      setShowErrorModal(true)
      setShowReservationModal(false)
      if (error.response) {
        setErrorMessage(error.response.data.message)
      } else {
        setErrorMessage(error.message)
      }
    }
  }

  function onClickCloseErrorModal() {
    setShowErrorModal(false)
  }

  // *****
  // use effect
  // *****

  useEffect(() => {
    if (clusters.length === 0) {
      fetchClusters(false)
    } else {
      setIsPageReady(true)
    }
  }, [])

  useEffect(() => {
    let shouldExit = false
    if (isPageReady) {
      if (clusters.length > 0) {
        const cluster = clusters.find((item) => item.name === name)
        if (cluster !== undefined && cluster.vips.length < 2) {
          setClusterToEdit(cluster)
        } else {
          shouldExit = true
        }
      } else {
        shouldExit = true
      }
    }

    if (shouldExit) {
      navigate({
        pathname: '/cluster'
      })
    }
  }, [isPageReady])

  return (
    <LoadbalancerReservations
      state={initialState}
      form={form}
      loading={loading || !isPageReady || !clusterToEdit}
      onSubmit={onSubmit}
      onChangeInput={onChangeInput}
      errorMessage={errorMessage}
      showReservationModal={showReservationModal}
      showErrorModal={showErrorModal}
      onClickCloseErrorModal={onClickCloseErrorModal}
    />
  )
}

export default LoadbalancerReservationsContainer
