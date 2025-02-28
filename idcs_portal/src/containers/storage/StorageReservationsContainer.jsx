// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useEffect, useState } from 'react'
import useCloudAccountStore from '../../store/cloudAccountStore/CloudAccountStore'
import { BsDashCircle, BsCheckCircle, BsTrash3, BsStopCircle, BsPlayCircle, BsPencilFill } from 'react-icons/bs'
import CloudAccountService from '../../services/CloudAccountService'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import StorageReservations from '../../components/storage/storageReservations/StorageReservations'
import { useNavigate } from 'react-router'
import { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'
import StateTooltipCell from '../../utils/gridPagination/cellRender/StateTooltipCell'
import { isErrorInAuthorization } from '../../utils/apiError/apiError'
import { capitalizeString } from '../../utils/stringFormatHelper/StringFormatHelper'
import useToastStore from '../../store/toastStore/ToastStore'

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

const StorageReservationsContainer = () => {
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
      id: 'edit',
      name: getActionItemLabel('Edit'),
      status: isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_STORAGE_EDIT) ? ['Ready'] : [],
      label: 'Edit Storage'
    },
    {
      id: 'terminate',
      name: getActionItemLabel('Delete'),
      status: ['Ready', 'Provisioning', 'Failed'],
      label: 'Delete storage',
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
    title: 'No volumes found',
    subTitle: 'Your account currently has no storage volumes',
    action: {
      type: 'redirect',
      href: '/storage/reserve',
      label: 'Create volume'
    }
  }

  const emptyGridByFilter = {
    title: 'No volumes found',
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

  // Navigation
  const navigate = useNavigate()

  const [myreservations, setMyreservations] = useState(null)
  const [showActionModal, setShowActionModal] = useState(false)
  const [actionModalContent, setActionModalContent] = useState(modalContent)
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)
  const [showErrorModal, setShowErrorModal] = useState(false)
  const [errorModalContent, setErrorModalContent] = useState(errorModal)
  const [isPageReady, setIsPageReady] = useState(false)
  // Global State
  const storages = useCloudAccountStore((state) => state.storages)
  const loading = useCloudAccountStore((state) => state.loading)
  const setStorages = useCloudAccountStore((state) => state.setStorages)
  const setShouldRefreshStorages = useCloudAccountStore((state) => state.setShouldRefreshStorages)
  const { setInstances, setInstanceGroups } = useCloudAccountStore((state) => state)
  const showError = useToastStore((state) => state.showError)

  const refreshStorages = async (background) => {
    await setStorages(background)
  }

  // Hooks
  useEffect(() => {
    const fetch = async () => {
      const promises = [refreshStorages(false), setInstances(true), setInstanceGroups(true)]
      await Promise.all(promises)
      setIsPageReady(true)
    }
    fetch().catch((error) => {
      if (isErrorInAuthorization(error)) {
        setIsPageReady(true)
        setShowErrorModal(true)
        const content = { ...errorModal }
        content.message = error.response.data.message
        setErrorModalContent(content)
      } else throwError(error)
    })
    setShouldRefreshStorages(true)
    return () => {
      setShouldRefreshStorages(false)
    }
  }, [])

  useEffect(() => {
    if (!isPageReady) {
      return
    }
    setGridInfo()
  }, [storages, isPageReady])

  // functions
  function setGridInfo() {
    const gridInfo = []

    for (const index in storages) {
      const storage = { ...storages[index] }
      gridInfo.push({
        name:
          storage.status !== 'FSDeleting' && storage.status !== 'Deleting'
            ? {
                showField: true,
                type: 'hyperlink',
                value: storage.name,
                function: () => {
                  setDetails(storage.name)
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
          selectableValues: getActionsByStatus(storage.status),
          function: setAction
        }
      })
    }
    setMyreservations(gridInfo)
  }

  function getStatusInfo(instance) {
    return getActionItemLabel(instance.status, instance.message)
  }

  const setDetails = (name) => {
    if (name) navigate(`/storage/d/${name}`)
  }

  function setAction(action, item) {
    switch (action.id) {
      case 'edit':
        navigate({
          pathname: `/storage/d/${item.name}/edit`,
          search: '?backTo=grid'
        })
        break
      case 'terminate': {
        const copyModalContent = { ...modalContent }
        copyModalContent.label = action.label
        copyModalContent.instanceName = item.name
        copyModalContent.buttonLabel = action.buttonLabel
        copyModalContent.resourceId = item.resourceId
        copyModalContent.name = item.name
        setActionModalContent(copyModalContent)
        setShowActionModal(true)
        break
      }
      default:
        break
    }
  }

  function getActionsByStatus(status) {
    const result = []

    for (const index in actionsOptions) {
      const option = { ...actionsOptions[index] }
      if (option.status.find((item) => item === status)) {
        result.push(option)
      }
    }

    return result
  }

  const deleteStorage = async (resourceId) => {
    try {
      await CloudAccountService.deleteStorageByCloudAccount(resourceId)
      setTimeout(() => {
        refreshStorages(true)
      }, 1000)
      // In case terminating state does dot show inmmediately
      setTimeout(() => {
        refreshStorages(true)
      }, 5000)
    } catch (error) {
      let errorMessage = ''
      if (error.response) {
        errorMessage = error.response.data.message
        showError(capitalizeString(errorMessage), false)
      } else {
        throwError(error)
      }
    }
  }

  function actionOnModal(result) {
    if (result) {
      deleteStorage(actionModalContent.resourceId)
        .then(() => {
          setShowActionModal(false)
        })
        .catch((error) => {
          setShowActionModal(false)
          if (isErrorInAuthorization(error)) {
            setShowErrorModal(true)
            const content = { ...errorModal }
            content.message = error.response.data.message
            setErrorModalContent(content)
          } else throwError(error)
        })
    } else {
      setShowActionModal(result)
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
    <StorageReservations
      myreservations={myreservations ?? []}
      columns={columns}
      showActionModal={showActionModal}
      actionModalContent={actionModalContent}
      emptyGrid={emptyGridObject}
      loading={loading || myreservations === null}
      filterText={filterText}
      errorModalContent={errorModalContent}
      showErrorModal={showErrorModal}
      setShowErrorModal={setShowErrorModal}
      setFilter={setFilter}
      setShowActionModal={actionOnModal}
    />
  )
}

export default StorageReservationsContainer
