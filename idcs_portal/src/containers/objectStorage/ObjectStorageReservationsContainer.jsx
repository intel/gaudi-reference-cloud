// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useEffect, useState } from 'react'
import { BsDashCircle, BsCheckCircle, BsTrash3, BsStopCircle, BsPlayCircle, BsPencilFill } from 'react-icons/bs'
import BucketService from '../../services/BucketService'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useToastStore from '../../store/toastStore/ToastStore'
import ObjectStorageReservations from '../../components/objectStorage/objectStorageReservations/ObjectStorageReservations'
import { useNavigate } from 'react-router'
import useBucketStore from '../../store/bucketStore/BucketStore'
import StateTooltipCell from '../../utils/gridPagination/cellRender/StateTooltipCell'
import { isErrorInAuthorization } from '../../utils/apiError/apiError'

const getActionItemLabel = (text, statusStep = null) => {
  let message = null

  switch (text) {
    case 'Start':
      message = (
        <>
          {' '}
          <BsPlayCircle /> {text}{' '}
        </>
      )
      break
    case 'Stop':
      message = (
        <>
          {' '}
          <BsStopCircle /> {text}{' '}
        </>
      )
      break
    case 'Delete':
      message = (
        <>
          {' '}
          <BsTrash3 /> {text}{' '}
        </>
      )
      break
    case 'Edit':
      message = (
        <>
          {' '}
          <BsPencilFill /> {text}{' '}
        </>
      )
      break
    case 'Provisioning':
      message = <StateTooltipCell statusStep={statusStep} text={text} />
      break
    case 'Deleting':
      message = (
        <>
          {' '}
          <BsDashCircle /> {text}{' '}
        </>
      )
      break
    case 'Ready':
      message = (
        <>
          {' '}
          <BsCheckCircle /> {text}{' '}
        </>
      )
      break
    default:
      message = <> {text} </>
      break
  }

  return message
}

const ObjectStorageReservationsContainer = () => {
  // local state
  const columns = [
    {
      columnName: 'Name',
      targetColumn: 'name'
    },
    {
      columnName: 'Size',
      targetColumn: 'size'
    },
    {
      columnName: 'State',
      targetColumn: 'status'
    },
    {
      columnName: 'Created at',
      targetColumn: 'creationTimestamp'
    },
    {
      columnName: 'Actions',
      targetColumn: 'actions'
    }
  ]

  const actionsOptions = [
    {
      id: 'terminate',
      name: getActionItemLabel('Delete'),
      status: ['Ready', 'Provisioning', 'Failed'],
      label: 'Delete bucket',
      buttonLabel: 'Delete'
    }
  ]

  const modalContent = {
    label: '',
    buttonLabel: '',
    instanceName: '',
    resourceId: '',
    question: '',
    feedback: '',
    name: ''
  }

  const emptyGrid = {
    title: 'No buckets found',
    subTitle: 'Your account currently has no storage buckets.',
    action: {
      type: 'redirect',
      href: '/buckets/reserve',
      label: 'Create bucket'
    }
  }

  const emptyGridByFilter = {
    title: 'No buckets found',
    subTitle: 'The applied filter criteria did not match any items.',
    action: {
      type: 'function',
      href: () => setFilter('', true),
      label: 'Clear filters'
    }
  }

  const errorModal = {
    titleMessage: '',
    description: '',
    message: ''
  }

  const throwError = useErrorBoundary()

  const [myreservations, setMyreservations] = useState(null)
  const [showActionModal, setShowActionModal] = useState(false)
  const [actionModalContent, setActionModalContent] = useState(modalContent)
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)
  const [afterActionShowModal, setAfterActionShowModal] = useState(false)
  const [afterActionModalContent, setAfterActionModalContent] = useState(modalContent)
  const [showErrorModal, setShowErrorModal] = useState(false)
  const [errorModalContent, setErrorModalContent] = useState(errorModal)
  const [isPageReady, setIsPageReady] = useState(false)

  // Global State
  const objectStorages = useBucketStore((state) => state.objectStorages)
  const loading = useBucketStore((state) => state.loading)
  const setObjectStorages = useBucketStore((state) => state.setObjectStorages)
  const setShouldRefreshObjectStorages = useBucketStore((state) => state.setShouldRefreshObjectStorages)
  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)
  const currentSelectedBucket = useBucketStore((state) => state.currentSelectedBucket)
  const setCurrentSelectedBucket = useBucketStore((state) => state.setCurrentSelectedBucket)

  const refreshStorages = async (background) => {
    try {
      await setObjectStorages(background)
      setIsPageReady(true)
    } catch (error) {
      if (isErrorInAuthorization(error)) {
        setIsPageReady(true)
        setShowErrorModal(true)
        const content = { ...errorModal }
        content.message = error.response.data.message
        setErrorModalContent(content)
      } else throwError(error)
    }
  }

  // Navigation
  const navigate = useNavigate()

  // Hooks
  useEffect(() => {
    refreshStorages(false)
    setShouldRefreshObjectStorages(true)
    return () => {
      setShouldRefreshObjectStorages(false)
    }
  }, [])

  useEffect(() => {
    if (!isPageReady) {
      return
    }
    setGridInfo()
  }, [objectStorages, isPageReady])

  // functions
  function setGridInfo() {
    const gridInfo = []

    for (const index in objectStorages) {
      const storage = { ...objectStorages[index] }

      gridInfo.push({
        name:
          storage.status !== 'FSDeleting'
            ? {
                showField: true,
                type: 'hyperlink',
                value: storage.name,
                function: () => {
                  setDetails(storage)
                }
              }
            : storage.name,
        size: storage.storage,
        status: {
          showField: true,
          type: 'function',
          value: storage,
          sortValue: storage.status,
          function: getStatusInfo
        },
        creationTimestamp: {
          showField: true,
          type: 'date',
          toLocalTime: true,
          value: storage.creationTimestamp,
          format: 'MM/DD/YYYY h:mm a'
        },
        actions: {
          showField: true,
          type: 'Buttons',
          value: storage,
          selectableValues: getActionsByStatus(storage.status, actionsOptions),
          function: setAction
        }
      })
    }
    setMyreservations(gridInfo)
  }

  function getStatusInfo(instance) {
    return getActionItemLabel(instance.status, instance.message)
  }

  const setDetails = (bucket = null) => {
    setCurrentSelectedBucket(bucket)
    if (bucket) navigate(`/buckets/d/${bucket.name}`)
  }

  function setAction(action, item) {
    switch (action.id) {
      case 'terminate': {
        const copyModalContent = { ...modalContent }
        copyModalContent.label = action.label
        copyModalContent.instanceName = item.name
        copyModalContent.buttonLabel = action.buttonLabel
        copyModalContent.resourceId = item.resourceId
        copyModalContent.actionType = 'terminateBucket'
        copyModalContent.name = item.name
        setActionModalContent(copyModalContent)
        setShowActionModal(true)
        break
      }
      default:
        break
    }
  }

  function getActionsByStatus(status, options) {
    const result = []

    for (const index in options) {
      const option = { ...options[index] }
      if (option.status.find((item) => item === status)) {
        result.push(option)
      }
    }

    return result
  }

  const deleteBucket = async (resourceId) => {
    try {
      await BucketService.deleteObjectBucketByCloudAccount(resourceId)
      setTimeout(() => {
        refreshStorages(true)
      }, 1000)
      showSuccess('Bucket deleted successfully.')
      const copyModalContent = { ...modalContent }
      copyModalContent.label = 'Deleted bucket principals'
      copyModalContent.feedback =
        'Your bucket was deleted. The associated principals may still have policies and permissions associated with the deleted bucket name.'
      copyModalContent.buttonLabel = 'Ok'
      setAfterActionModalContent(copyModalContent)
      setAfterActionShowModal(true)
      // In case terminating state does not show inmmediately
      setTimeout(() => {
        refreshStorages(true)
      }, 5000)
    } catch (error) {
      if (isErrorInAuthorization(error)) {
        setShowErrorModal(true)
        const content = { ...errorModal }
        content.message = error.response.data.message
        setErrorModalContent(content)
      } else showError('Error while deleting the bucket. Please try again later.')
    }
  }

  const deleteRule = async (resourceId, bucketId) => {
    try {
      await BucketService.deleteObjectBucketRule(resourceId, bucketId)
      if (currentSelectedBucket.name === actionModalContent.instanceName) {
        setCurrentSelectedBucket(null)
      }
      setTimeout(() => {
        refreshStorages(true)
      }, 1000)
      showSuccess('Rule deleted successfully.')
      // In case terminating state does not show inmmediately
      setTimeout(() => {
        refreshStorages(true)
      }, 5000)
    } catch (error) {
      if (isErrorInAuthorization(error)) {
        setShowErrorModal(true)
        const content = { ...errorModal }
        content.message = error.response.data.message
        setErrorModalContent(content)
      } else showError('Error while deleting the rule. Please try again later.')
    }
  }

  function actionOnModal(result) {
    setShowActionModal(result)
    if (result) {
      if (actionModalContent.actionType === 'terminateBucket') {
        deleteBucket(actionModalContent.resourceId)
      } else {
        deleteRule(actionModalContent.resourceId, currentSelectedBucket.resourceId)
      }
      setShowActionModal(false)
    }
  }

  function setFilter(event, clear) {
    if (clear) {
      setEmptyGridObject(emptyGrid)
      setFilterText('')
    } else {
      setEmptyGridObject(emptyGridByFilter)
      setFilterText(event.target.value)
    }
  }

  return (
    <ObjectStorageReservations
      myreservations={myreservations ?? []}
      columns={columns}
      showActionModal={showActionModal}
      actionModalContent={actionModalContent}
      emptyGrid={emptyGridObject}
      loading={loading || myreservations === null}
      filterText={filterText}
      setFilter={setFilter}
      setShowActionModal={actionOnModal}
      afterActionModalContent={afterActionModalContent}
      afterActionShowModal={afterActionShowModal}
      setAfterActionShowModal={setAfterActionShowModal}
      errorModalContent={errorModalContent}
      showErrorModal={showErrorModal}
      setShowErrorModal={setShowErrorModal}
    />
  )
}

export default ObjectStorageReservationsContainer
