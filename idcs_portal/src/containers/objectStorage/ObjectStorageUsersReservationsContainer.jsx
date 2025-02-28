// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useEffect, useState } from 'react'
import { BsPencilFill, BsDashCircle, BsCheckCircle, BsTrash3, BsStopCircle, BsPlayCircle } from 'react-icons/bs'
import BucketService from '../../services/BucketService'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useToastStore from '../../store/toastStore/ToastStore'
import { useNavigate } from 'react-router'
import ObjectStorageUsersReservations from '../../components/objectStorage/objectStorageUsersReservations/ObjectStorageUsersReservations'
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

const ObjectStorageUsersReservationsContainer = () => {
  // local state
  const columns = [
    {
      columnName: 'Name',
      targetColumn: 'name'
    },
    {
      columnName: 'Status',
      targetColumn: 'status'
    },
    {
      columnName: 'Created at',
      targetColumn: 'creationTimestamp'
    },
    {
      columnName: 'Modified at',
      targetColumn: 'updateTimestamp'
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
      status: ['Ready'],
      label: 'Edit principal'
    },
    {
      id: 'terminate',
      name: getActionItemLabel('Delete'),
      status: ['Ready', 'Provisioning', 'Failed'],
      label: 'Delete principal',
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
    title: 'No principals found',
    subTitle: 'Your account currently has no principals.',
    action: {
      type: 'redirect',
      href: '/buckets/users/reserve',
      label: 'Create principal'
    }
  }

  const emptyGridByFilter = {
    title: 'No principals found',
    subTitle: 'The applied filter criteria did not match any items',
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

  const [myUsers, setMyUsers] = useState(null)
  const [showActionModal, setShowActionModal] = useState(false)
  const [actionModalContent, setActionModalContent] = useState(modalContent)
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)
  const [showErrorModal, setShowErrorModal] = useState(false)
  const [errorModalContent, setErrorModalContent] = useState(errorModal)
  const [isPageReady, setIsPageReady] = useState(false)

  // Global State
  const bucketUsers = useBucketStore((state) => state.bucketUsers)
  const loading = useBucketStore((state) => state.loading)
  const setBucketUsers = useBucketStore((state) => state.setBucketUsers)
  const setShouldRefreshBucketUsers = useBucketStore((state) => state.setShouldRefreshBucketUsers)
  const showError = useToastStore((state) => state.showError)
  const { showSuccess } = useToastStore((state) => state)
  const currentSelectedBucketUser = useBucketStore((state) => state.currentSelectedBucketUser)
  const setCurrentSelectedBucketUser = useBucketStore((state) => state.setCurrentSelectedBucketUser)

  const refreshUsers = async (background) => {
    try {
      await setBucketUsers(background)
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
  useEffect(() => {}, [])

  useEffect(() => {
    async function fetch() {
      await refreshUsers(false)
    }
    fetch()
    setShouldRefreshBucketUsers(true)
    return () => {
      setShouldRefreshBucketUsers(false)
    }
  }, [])

  useEffect(() => {
    if (!isPageReady) {
      return
    }
    setGridInfo()
  }, [bucketUsers, isPageReady])

  // functions
  function setGridInfo() {
    const gridInfo = []

    for (const index in bucketUsers) {
      const user = { ...bucketUsers[index] }

      gridInfo.push({
        name:
          user.status !== 'FSDeleting'
            ? {
                showField: true,
                type: 'hyperlink',
                value: user.name,
                function: () => {
                  setDetails(user.name)
                }
              }
            : user.name,
        status: {
          showField: true,
          type: 'function',
          value: user,
          sortValue: user.status,
          function: getStatusInfo
        },
        creationTimestamp: {
          showField: true,
          type: 'date',
          toLocalTime: true,
          value: user.creationTimestamp,
          format: 'MM/DD/YYYY h:mm a'
        },
        updateTimestamp: {
          showField: true,
          type: 'date',
          toLocalTime: true,
          value: user.updateTimestamp,
          format: 'MM/DD/YYYY h:mm a'
        },
        actions: {
          showField: true,
          type: 'Buttons',
          value: user,
          selectableValues: getActionsByStatus(user.status),
          function: setAction
        }
      })
    }
    setMyUsers(gridInfo)
  }

  function getStatusInfo(instance) {
    return getActionItemLabel(instance.status, instance.message)
  }

  const setDetails = (name = null) => {
    setCurrentSelectedBucketUser(name)
    if (name) navigate(`/buckets/users/d/${name}`)
  }

  function setAction(action, item) {
    switch (action.id) {
      case 'edit':
        navigate({
          pathname: `/buckets/users/d/${item.name}/edit`,
          search: '?backTo=grid'
        })
        break
      case 'terminate': {
        const copyModalContent = { ...modalContent }
        copyModalContent.label = action.label
        copyModalContent.name = item.name
        copyModalContent.buttonLabel = action.buttonLabel
        copyModalContent.resourceId = item.userId
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

  const deleteUser = async (userName) => {
    try {
      await BucketService.deleteBucketUserByCloudAccount(userName)

      setTimeout(() => {
        refreshUsers(true)
      }, 1000)
      showSuccess('Principal deleted successfully.')
      // In case terminating state does dot show inmmediately
      setTimeout(() => {
        refreshUsers(true)
      }, 5000)
    } catch (error) {
      if (isErrorInAuthorization(error)) {
        setShowErrorModal(true)
        const content = { ...errorModal }
        content.message = error.response.data.message
        setErrorModalContent(content)
      } else showError('Could not able to delete principal. Please try again later.')
    }
  }

  function actionOnModal(result) {
    if (result) {
      deleteUser(actionModalContent.name)
        .then(() => {
          if (currentSelectedBucketUser === actionModalContent.name) {
            setCurrentSelectedBucketUser(null)
          }
          setShowActionModal(false)
        })
        .catch((error) => {
          setShowActionModal(false)
          throwError(error)
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
    <ObjectStorageUsersReservations
      myUsers={myUsers ?? []}
      columns={columns}
      showActionModal={showActionModal}
      actionModalContent={actionModalContent}
      emptyGrid={emptyGridObject}
      loading={loading || myUsers === null}
      filterText={filterText}
      errorModalContent={errorModalContent}
      showErrorModal={showErrorModal}
      setShowErrorModal={setShowErrorModal}
      setShowActionModal={actionOnModal}
      setFilter={setFilter}
    />
  )
}

export default ObjectStorageUsersReservationsContainer
