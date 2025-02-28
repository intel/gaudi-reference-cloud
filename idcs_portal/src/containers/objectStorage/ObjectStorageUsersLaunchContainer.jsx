// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router'
import BucketService from '../../services/BucketService'
import {
  UpdateFormHelper,
  isValidForm,
  getFormValue,
  showFormRequiredFields
} from '../../utils/updateFormHelper/UpdateFormHelper'
import ObjectStorageUserLaunch from '../../components/objectStorage/objectStorageUserLaunch/ObjectStorageUserLaunch'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useBucketUsersPermissionsStore from '../../store/bucketUsersPermissionsStore/BucketUsersPermissionsStore'
import useBucketStore from '../../store/bucketStore/BucketStore'
import EmptyView from '../../utils/emptyView/EmptyView'
import useToastStore from '../../store/toastStore/ToastStore'
import Spinner from '../../utils/spinner/Spinner'
import { isErrorInAuthorization } from '../../utils/apiError/apiError'

const getCustomMessage = (messageType) => {
  let message = null

  switch (messageType) {
    case 'name':
      message = (
        <div className="valid-feedback" intc-id={'ObjectStorageUsersLaunchNameValidMessage'}>
          Max length 12 characters.
        </div>
      )
      break
    default:
      break
  }

  return message
}

const ObjectStorageUsersLaunchContainer = () => {
  // local state

  const initialState = {
    mainTitle: 'Create principal',
    form: {
      name: {
        sectionGroup: 'configuration',
        type: 'text', // options = 'text ,'textArea'
        label: 'Name:',
        placeholder: 'Name',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 12,
        validationRules: {
          isRequired: true,
          onlyAlphaNumLower: true,
          checkMaxLength: true
        },
        validationMessage: '', // Errror message to display to the user
        helperMessage: getCustomMessage('name')
      }
    },
    servicePayload: {
      metadata: {
        name: ''
      },
      spec: []
    },
    isValidForm: false,
    showReservationModal: false,
    showErrorModal: false,
    errorMessage: '',
    errorTitleMessage: '',
    errorDescription: '',
    navigationBottom: [
      {
        buttonLabel: 'Create',
        buttonVariant: 'primary'
      },
      {
        buttonLabel: 'Cancel',
        buttonVariant: 'link',
        buttonFunction: () => onCancel()
      }
    ],
    timeoutMiliseconds: 4000
  }

  const emptyBuckets = {
    title: 'No buckets found',
    subTitle: 'To create a principal, please create a bucket first.',
    action: {
      type: 'redirect',
      href: '/buckets/reserve',
      label: 'Create bucket'
    }
  }

  const throwError = useErrorBoundary()

  const [state, setState] = useState(initialState)
  const [isPageReady, setIsPageReady] = useState(false)
  const [userPermission, setUserPermission] = useState([])
  const objectStorages = useBucketStore((state) => state.objectStorages)
  const setObjectStorages = useBucketStore((state) => state.setObjectStorages)
  const selectionType = useBucketUsersPermissionsStore((state) => state.selectionType)
  const setSelectionType = useBucketUsersPermissionsStore((state) => state.setSelectionType)
  const bucketsPermissions = useBucketUsersPermissionsStore((state) => state.bucketsPermissions)
  const setCurrentSelectedBucketUser = useBucketStore((state) => state.setCurrentSelectedBucketUser)
  const showError = useToastStore((state) => state.showError)

  // Hooks
  useEffect(() => {
    async function fetch() {
      setSelectionType('All')
      await setObjectStorages(false)
      setIsPageReady(true)
    }
    fetch().catch((error) => {
      if (isErrorInAuthorization(error)) {
        const stateUpdated = { ...state }
        stateUpdated.showErrorModal = true
        stateUpdated.errorMessage = error.response.data.message
        setState(stateUpdated)
      } else throwError(error)
    })
  }, [])

  useEffect(() => {
    setBucketsPermissions()
  }, [bucketsPermissions, selectionType])

  // Navigation
  const navigate = useNavigate()

  // Functions

  function setBucketsPermissions() {
    const userPermission = []
    for (const i in bucketsPermissions) {
      const permissions = bucketsPermissions[i]
      userPermission.push(permissions)
    }
    setUserPermission(userPermission)
  }

  function onCancel() {
    // Navigates back to the page when this method triggers.
    navigate('/buckets/users')
  }

  function onChangeInput(event, formInputName) {
    const value = event.target.value
    const updatedState = {
      ...state
    }

    const updatedForm = UpdateFormHelper(value, formInputName, updatedState.form)

    updatedState.isValidForm = isValidForm(updatedForm)

    updatedState.form = updatedForm

    setState(updatedState)
  }

  function showRequiredFields() {
    const stateCopy = { ...state }
    // Mark regular Inputs
    const updatedForm = showFormRequiredFields(stateCopy.form)
    // Create toast
    showError('Please complete required fields and permissions.', false)
    stateCopy.form = updatedForm
    setState(stateCopy)
  }

  async function onSubmit(e) {
    try {
      const stateCopy = { ...state }
      const isValidForm = stateCopy.isValidForm

      if (!isValidForm) {
        showRequiredFields()
        return
      }

      const payloadCopy = { ...stateCopy.servicePayload }
      payloadCopy.metadata.name = getFormValue('name', stateCopy.form)

      if (selectionType === 'All' && userPermission?.find((x) => x.bucketId === 'All')) {
        const allSelection = userPermission?.find((x) => x.bucketId === 'All')
        payloadCopy.spec = objectStorages.map((bucket) => {
          const userObj = { ...allSelection }
          userObj.bucketId = bucket.name
          return userObj
        })
      } else {
        payloadCopy.spec = userPermission
      }

      stateCopy.showReservationModal = true
      setState(stateCopy)

      await createUser(payloadCopy)
    } catch (error) {
      const stateUpdated = { ...state }
      stateUpdated.showReservationModal = false
      if (error.response) {
        if (isErrorInAuthorization(error)) {
          stateUpdated.errorMessage = error.response.data.message
        } else {
          stateUpdated.errorMessage =
            error.response.data.message === 'missing spec'
              ? 'Your account currently has no active buckets.'
              : error.response.data.message
        }
      } else {
        stateUpdated.errorMessage = error.message
      }
      stateUpdated.showErrorModal = true
      stateUpdated.errorTitleMessage = 'Could not create your principal'
      stateUpdated.errorDescription = 'There was an error while processing your principal.'
      setState(stateUpdated)
    }
  }

  function onClickCloseErrorModal() {
    const stateUpdated = { ...state }
    stateUpdated.showErrorModal = false
    setState(stateUpdated)
  }

  async function createUser(servicePayload) {
    await BucketService.postBucketUsers(servicePayload)
    setCurrentSelectedBucketUser(servicePayload.metadata.name)
    setTimeout(() => {
      const stateUpdated = { ...state }

      stateUpdated.showReservationModal = false

      setState(stateUpdated)

      navigate({
        pathname: '/buckets/users'
      })
    }, state.timeoutMiliseconds)
  }
  return objectStorages.length > 0 ? (
    <ObjectStorageUserLaunch
      state={state}
      onClickCloseErrorModal={onClickCloseErrorModal}
      onSubmit={onSubmit}
      onChangeInput={onChangeInput}
      objectStorages={objectStorages}
    />
  ) : isPageReady ? (
    <EmptyView title={emptyBuckets.title} subTitle={emptyBuckets.subTitle} action={emptyBuckets.action} />
  ) : (
    <Spinner />
  )
}

export default ObjectStorageUsersLaunchContainer
