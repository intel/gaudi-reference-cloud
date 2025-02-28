// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useNavigate, useParams } from 'react-router'
import BucketService from '../../services/BucketService'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useBucketUsersPermissionsStore from '../../store/bucketUsersPermissionsStore/BucketUsersPermissionsStore'
import useBucketStore from '../../store/bucketStore/BucketStore'
import ObjectStorageUserEdit from '../../components/objectStorage/objectStorageUserEdit/ObjectStorageUserEdit'
import { isErrorInAuthorization } from '../../utils/apiError/apiError'

const ObjectStorageUsersEditContainer = () => {
  const { param: name } = useParams()
  // variables
  const initialState = {
    mainTitle: `Edit principal ${name}`,
    servicePayload: {
      spec: []
    },
    showReservationModal: false,
    showErrorModal: false,
    errorMessage: '',
    errorTitleMessage: '',
    errorDescription: '',
    navigationBottom: [
      {
        buttonLabel: 'Save',
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
    subTitle: 'To edit a principal, please create a bucket first.',
    action: {
      type: 'redirect',
      href: '/buckets/reserve',
      label: 'Create bucket'
    }
  }

  const throwError = useErrorBoundary()
  const [searchParams] = useSearchParams()

  // local state
  const [state, setState] = useState(initialState)
  const [isPageReady, setIsPageReady] = useState(false)
  const [userPermission, setUserPermission] = useState([])

  // global state
  const objectStorages = useBucketStore((state) => state.objectStorages)
  const selectionType = useBucketUsersPermissionsStore((state) => state.selectionType)
  const setSelectionType = useBucketUsersPermissionsStore((state) => state.setSelectionType)
  const bucketsPermissions = useBucketUsersPermissionsStore((state) => state.bucketsPermissions)
  const setBucketsPermissions = useBucketUsersPermissionsStore((state) => state.setBucketsPermissions)
  const bucketUsers = useBucketStore((state) => state.bucketUsers)
  const loading = useBucketStore((state) => state.loading)
  const setCurrentSelectedBucketUser = useBucketStore((state) => state.setCurrentSelectedBucketUser)
  const setBucketUsers = useBucketStore((state) => state.setBucketUsers)
  const setObjectStorages = useBucketStore((state) => state.setObjectStorages)

  const refreshUsers = async (background) => {
    await setBucketUsers(background)
  }

  // Navigation
  const navigate = useNavigate()

  // Hooks
  useEffect(() => {
    setTitle()
    const fetch = async () => {
      setSelectionType('bucket')
      if (bucketUsers.length === 0) await refreshUsers(false)
      if (objectStorages.length === 0) await setObjectStorages(false)
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
    if (state.showReservationModal) {
      return
    }
    const user = bucketUsers.find((instance) => instance.name === name)
    if (user === undefined) {
      if (isPageReady) {
        navigate({
          pathname: '/buckets/users'
        })
      }
      return
    }
    const userDetail = { ...user }
    setBucketsPermissions(userDetail.spec)
  }, [bucketUsers, isPageReady, objectStorages])

  useEffect(() => {
    if (isPageReady) {
      updateDetails()
    }
  }, [bucketsPermissions, selectionType, isPageReady])

  // Functions

  function goBack() {
    const backTo = searchParams.get('backTo')
    switch (backTo) {
      case 'detail':
        navigate({
          pathname: `/buckets/users/d/${name}`
        })
        break
      default:
        navigate({
          pathname: '/buckets/users'
        })
        break
    }
  }

  function setTitle() {
    const stateUpdated = {
      ...state
    }

    if (name) {
      stateUpdated.mainTitle = 'Edit principal - ' + name
    }
    setState(stateUpdated)
  }

  const updateDetails = () => {
    const userPermission = []
    const buckets = objectStorages.map((x) => x.name)

    for (const i in bucketsPermissions) {
      const permissions = selectionType === 'All' ? bucketsPermissions.All : bucketsPermissions[i]
      if (buckets.includes(i)) {
        userPermission.push(permissions)
      }
    }
    setUserPermission(userPermission)
  }

  function onCancel() {
    // Navigates back to the page when this method triggers.
    goBack()
  }

  async function onSubmit(e) {
    try {
      const stateCopy = { ...state }

      const payloadCopy = { ...stateCopy.servicePayload }

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

      await editUser(payloadCopy)
    } catch (error) {
      const stateUpdated = { ...state }
      stateUpdated.showReservationModal = false
      if (error.response) {
        // TODO: keep auth check condition for future update, even though it currently produces identical output. This will make it easier to adjust permissions if requirements change.
        if (isErrorInAuthorization(error)) {
          stateUpdated.errorMessage = error.response.data.message
        } else stateUpdated.errorMessage = error.response.data.message
      } else {
        stateUpdated.errorMessage = error.message
      }
      stateUpdated.showErrorModal = true
      stateUpdated.errorTitleMessage = 'Could not update your principal'
      stateUpdated.errorDescription = 'There was an error while processing your principal.'
      setState(stateUpdated)
    }
  }

  function onClickCloseErrorModal() {
    const stateUpdated = { ...state }
    stateUpdated.showErrorModal = false
    setState(stateUpdated)
  }

  async function editUser(servicePayload) {
    const userName = name
    await BucketService.updateBucketUser(servicePayload, userName)
    setCurrentSelectedBucketUser(userName)
    refreshUsers(true)
    setTimeout(() => {
      const stateUpdated = { ...state }

      stateUpdated.showReservationModal = false

      setState(stateUpdated)

      goBack()
    }, state.timeoutMiliseconds)
  }

  return (
    <ObjectStorageUserEdit
      state={state}
      loading={loading || !isPageReady}
      onClickCloseErrorModal={onClickCloseErrorModal}
      onSubmit={onSubmit}
      objectStorages={objectStorages}
      emptyBuckets={emptyBuckets}
    />
  )
}

export default ObjectStorageUsersEditContainer
