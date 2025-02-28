// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import {
  UpdateFormHelper,
  isValidForm,
  getFormValue,
  setFormValue,
  setSelectOptions
} from '../../utility/updateFormHelper/UpdateFormHelper'
import CloudAccountService from '../../services/CloudAccountService'
import useNodePoolStore from '../../store/nodePoolStore/NodePoolStore'
import useUserStore from '../../store/userStore/UserStore'
import AddCloudAccount from '../../components/nodePoolManagement/AddCloudAccount'
import useToastStore from '../../store/toastStore/ToastStore'
import NodePoolService from '../../services/NodePoolService'

const AddCloudAccountContainer = () => {
  const navigate = useNavigate()

  const { poolId } = useParams()
  if (!poolId) navigate('/npm/pools')

  // Initial state for form and validation
  const initialState = {
    mainSubtitle: 'Specify the needed Information',
    form: {
      pool: {
        sectionGroup: 'pool',
        type: 'select', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Compute Node Pool:',
        placeholder: 'Please select',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [],
        validationMessage: ''
      },
      cloudAccount: {
        sectionGroup: 'cloudAccount',
        type: 'text', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Cloud Account:',
        placeholder: 'Enter Cloud Account Email or ID',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 63,
        validationRules: {
          isRequired: true,
          checkMaxLength: true
        },
        validationMessage: '' // Errror message to display to the user
      }
    },
    isValidForm: false,
    timeoutMilliseconds: 5000
  }

  // Initial state for loader
  const initialLoaderData = {
    isShow: false,
    message: ''
  }

  // Error Boundary
  const throwError = useErrorBoundary()

  const [state, setState] = useState(initialState)
  const [showLoader, setShowLoader] = useState(initialLoaderData)
  const [selectedPool, setSelectedPool] = useState(null)
  const [selectedCloudAccount, setSelectedCloudAccount] = useState(null)

  // Global States)
  const poolList = useNodePoolStore((state) => state.poolList)
  const user = useUserStore((state) => state.user)
  const setPoolList = useNodePoolStore((state) => state.setPoolList)
  const showError = useToastStore((state) => state.showError)

  const backButtonLabel = '⟵ Back to Cloud Accounts of ' + poolId

  // Hooks
  useEffect(() => {
    if (poolList.length === 0) {
      const fetchProducts = async () => {
        try {
          await setPoolList()
        } catch (error) {
          throwError(error)
        }
      }

      fetchProducts()
    }
  }, [])

  useEffect(() => {
    setForm()
  }, [poolList])

  // functions
  function setForm() {
    const stateUpdated = {
      ...state
    }

    // Load Pools
    const selectablePools = getPoolOptions()
    if (selectablePools.length > 0) {
      stateUpdated.form = setSelectOptions('pool', selectablePools, stateUpdated.form)

      const findPool = getSelectedPool(poolId)

      const selectedPool = findPool ? poolId : selectablePools[0].value
      stateUpdated.form = setFormValue('pool', selectedPool, stateUpdated.form)

      setSelectedPool(getSelectedPool(selectedPool))
    }

    setState(stateUpdated)
  }

  function getSelectedPool(poolId) {
    return poolList.find((pool) => pool.poolId === poolId)
  }

  function getPoolOptions() {
    return poolList.map((pool) => {
      return {
        name: pool.poolId,
        value: pool.poolId
      }
    })
  }

  function onChangeInput(event, formInputName) {
    const updatedState = {
      ...state
    }

    const updatedForm = UpdateFormHelper(event.target.value, formInputName, updatedState.form)

    updatedState.isValidForm = isValidForm(updatedForm)
    updatedState.form = updatedForm

    if (formInputName === 'cloudAccount') {
      setSelectedCloudAccount(null)
      updatedState.isValidForm = false
    }

    if (formInputName === 'pool') {
      setSelectedPool(getSelectedPool(event.target.value))
    }

    if (!selectedCloudAccount) {
      updatedState.isValidForm = false
    }

    setState(updatedState)
  }

  async function onSearchCloudAccount() {
    const cloudAccount = getFormValue('cloudAccount', state.form)
    setCloudAccountError('')
    setSelectedCloudAccount(null)
    if (cloudAccount !== '') {
      try {
        let data
        if (cloudAccount.includes('@') || /[a-zA-Z]/.test(cloudAccount)) {
          data = await CloudAccountService.getCloudAccountDetailsByName(cloudAccount)
        } else {
          data = await CloudAccountService.getCloudAccountDetailsById(cloudAccount)
        }

        setSelectedCloudAccount(data?.data)
        const updatedState = { ...state }
        updatedState.form.cloudAccount.validationMessage = ''
        updatedState.isValidForm = isValidForm(updatedState.form)
        setState(updatedState)
      } catch (e) {
        const code = e.response.data?.code
        const errorMsg = e.response.data?.message
        const message =
          code && [3, 5].includes(code)
            ? String(errorMsg.charAt(0).toUpperCase()) + String(errorMsg.slice(1))
            : 'Cloud Account ID is not found'
        setCloudAccountError(message)
        setSelectedCloudAccount(null)
      }
    } else {
      setCloudAccountError('Cloud Account Number is required')
    }
  }

  function setCloudAccountError(errorMessage) {
    const updatedState = { ...state }
    const updatedForm = { ...updatedState.form }
    const updatedFormElement = { ...updatedForm.cloudAccount }
    updatedFormElement.isValid = false
    updatedFormElement.validationMessage = errorMessage
    updatedForm.cloudAccount = updatedFormElement
    updatedState.form = updatedForm
    updatedState.isValidForm = false
    setState(updatedState)
  }

  function onCancel() {
    navigate('/npm/pools/accounts/' + poolId)
  }

  function onSubmit() {
    setShowLoader({ isShow: true, message: 'Working on your request' })
    submitForm()
  }

  async function submitForm() {
    try {
      const payload = {
        cloudAccountId: selectedCloudAccount.id,
        createAdmin: user.email
      }

      const pId = selectedPool.poolId

      await NodePoolService.addCloudAccountsToPool(pId, payload)

      setTimeout(() => {
        navigate('/npm/pools/accounts/' + pId)
      }, state.timeoutMilliseconds)
    } catch (error) {
      let message = ''
      if (error.response) {
        if (error.response.data.message !== '') {
          message = error.response.data.message
        } else {
          message = error.message
        }
      } else {
        message = error.message
      }
      setShowLoader({ isShow: false })
      showError(message)
    }
  }

  return (
    <AddCloudAccount
      state={state}
      showLoader={showLoader}
      backButtonLabel={backButtonLabel}
      onCancel={onCancel}
      onSubmit={onSubmit}
      onChangeInput={onChangeInput}
      selectedCloudAccount={selectedCloudAccount}
      selectedPool={selectedPool}
      onSearchCloudAccount={onSearchCloudAccount}
    />
  )
}

export default AddCloudAccountContainer
