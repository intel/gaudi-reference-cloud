// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import { BsPencilFill, BsTrash3 } from 'react-icons/bs'
import BucketService from '../../services/BucketService'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useToastStore from '../../store/toastStore/ToastStore'
import { useNavigate, useParams } from 'react-router'
import ObjectStorageUsersReservationsDetails from '../../components/objectStorage/objectStorageUsersDetails/ObjectStorageUsersReservationsDetails'
import { ArrowCounterclockwise } from 'react-bootstrap-icons'
import { useCopy } from '../../hooks/useCopy'
import useBucketUsersPermissionsStore from '../../store/bucketUsersPermissionsStore/BucketUsersPermissionsStore'
import useBucketStore from '../../store/bucketStore/BucketStore'
import SpinnerIcon from '../../utils/spinner/SpinnerIcon'
import { isErrorInAuthorization } from '../../utils/apiError/apiError'

const getActionItemLabel = (text: string, statusStep: string | null = null): JSX.Element => {
  let message = null

  switch (text) {
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
    default:
      message = <> {text} </>
      break
  }

  return message
}

const ObjectStorageUsersReservationsDetailsContainer = (): JSX.Element => {
  // local state
  const tabsInitial = [
    {
      label: 'Credentials',
      id: 'credentials',
      show: true
    },
    {
      label: 'Permissions',
      id: 'permissions',
      show: true
    }
  ]

  const tabDetailsInitial: any = [
    {
      tapTitle: 'Bucket credentials',
      tapConfig: { type: 'custom' },
      fields: [
        {
          label: 'Principal:',
          field: 'name',
          value: '',
          action: [{ func: copyItem, label: 'Copy', icon: 'Copy' }]
        }
      ]
    },
    {
      tapTitle: 'Buckets permissions',
      tapConfig: { type: 'custom' },
      customContent: null
    }
  ]

  const actionsOptions: any = [
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

  const showGeneratePwdInitial = {
    show: true,
    label: 'Generate password',
    icon: <ArrowCounterclockwise />
  }

  const errorModal = {
    titleMessage: '',
    description: '',
    message: ''
  }

  const throwError = useErrorBoundary()

  const [reserveDetails, setReserveDetails] = useState<any>(null)
  const [actionsReserveDetails, setActionsReserveDetails] = useState<any[]>([])
  const [activeTab, setActiveTab] = useState(0)
  const [showActionModal, setShowActionModal] = useState(false)
  const [actionModalContent, setActionModalContent] = useState(modalContent)
  const [tabDetails, setTabDetails] = useState(tabDetailsInitial)
  const [showGeneratePwd, setShowGenerate] = useState(showGeneratePwdInitial)
  const [showAccessKey, setShowAccessKey] = useState<string | null>(null)
  const [showSecretKey, setShowSecretKey] = useState<string | null>(null)
  const [userBuckets, setUserBuckets] = useState<any[]>([])
  const [isPageReady, setIsPageReady] = useState(false)
  const [showErrorModal, setShowErrorModal] = useState(false)
  const [errorModalContent, setErrorModalContent] = useState(errorModal)
  // Global State
  const bucketUsers = useBucketStore((state) => state.bucketUsers)
  const loading = useBucketStore((state) => state.loading)
  const setBucketUsers = useBucketStore((state) => state.setBucketUsers)
  const setShouldRefreshBucketUsers = useBucketStore((state) => state.setShouldRefreshBucketUsers)
  const showError = useToastStore((state) => state.showError)
  const { copyToClipboard } = useCopy()
  const { showSuccess } = useToastStore((state) => state)
  const setBucketsPermissions: any = useBucketUsersPermissionsStore((state) => state.setBucketsPermissions)
  const setCurrentSelectedBucketUser = useBucketStore((state) => state.setCurrentSelectedBucketUser)

  const refreshUsers = async (background: boolean): Promise<void> => {
    try {
      await setBucketUsers(background)
    } catch (error: any) {
      if (isErrorInAuthorization(error)) {
        setShowErrorModal(true)
        const content = { ...errorModal }
        content.message = error.response.data.message
        setErrorModalContent(content)
      } else throwError(error)
    }
  }

  // Navigation
  const navigate = useNavigate()
  const { param: name } = useParams()

  // Hooks
  useEffect(() => {
    const fetch = async (): Promise<void> => {
      if (bucketUsers.length === 0) await refreshUsers(false)
      setIsPageReady(true)
    }
    fetch().catch((error) => {
      throwError(error)
    })
    setShouldRefreshBucketUsers(true)
    return () => {
      setShouldRefreshBucketUsers(false)
    }
  }, [])

  useEffect(() => {
    updateDetails()
  }, [bucketUsers, isPageReady])

  function copyItem(value: string): void {
    copyToClipboard(value)
  }

  // functions
  const updateDetails = (): void => {
    setShowAccessKey(null)
    setShowSecretKey(null)
    setBucketsPermissions(null)
    setUserBuckets([])

    const user = bucketUsers.find((instance) => instance.name === name)
    if (user === undefined) {
      if (isPageReady) navigate('/buckets/users')
      setActionsReserveDetails([])
      setReserveDetails(null)
      return
    }

    const userDetail: any = { ...user }

    const tabDetailsCopy = []
    for (const tapIndex in tabDetails) {
      const tapDetail = { ...tabDetails[tapIndex] }
      const updateFields = []
      for (const index in tapDetail.fields) {
        const field = { ...tapDetail.fields[index] }
        field.value = userDetail[field.field]
        updateFields.push(field)
      }
      tapDetail.fields = updateFields
      tabDetailsCopy.push(tapDetail)
    }

    setTabDetails(tabDetailsCopy)
    setActionsReserveDetails(getActionsByStatus(userDetail.status))
    setReserveDetails(userDetail)
    setBucketsPermissions(userDetail.spec)
    const userSpec = Object.keys(userDetail.spec).map((x) => {
      return {
        name: x
      }
    })
    setUserBuckets(userSpec)
  }

  function setAction(action: any, item: any): void {
    switch (action.id) {
      case 'edit':
        navigate({
          pathname: `/buckets/users/d/${item.name}/edit`,
          search: '?backTo=detail'
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

  function getActionsByStatus(status: string): any[] {
    const result = []

    for (const index in actionsOptions) {
      const option = { ...actionsOptions[index] }
      if (option.status.find((item: string) => item === status)) {
        result.push(option)
      }
    }

    return result
  }

  const deleteUser = async (userName: string): Promise<void> => {
    try {
      await BucketService.deleteBucketUserByCloudAccount(userName)

      setTimeout(() => {
        refreshUsers(true).catch((error) => {
          throwError(error)
        })
      }, 1000)
      showSuccess('Principal deleted successfully.', false)
      setTimeout(() => {
        navigate('/buckets/users')
      }, 2000)
    } catch (error) {
      showError('Could not able to delete principal. Please try again later.', false)
    }
  }

  function actionOnModal(result: boolean): void {
    if (result) {
      deleteUser(actionModalContent.name)
        .then(() => {
          if (reserveDetails.name === actionModalContent.name) {
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

  async function generatePwd(): Promise<void> {
    setShowGenerate({
      ...showGeneratePwdInitial,
      label: 'Generating password',
      icon: <SpinnerIcon />
    })

    try {
      const { data } = await BucketService.updateBucketUserCredentialsByCloudAccount(reserveDetails.name)

      const accessKey = data?.status?.principal?.credentials?.accessKey
      const secretKey = data?.status?.principal?.credentials?.secretKey

      setShowAccessKey(accessKey)
      setShowSecretKey(secretKey)

      setShowGenerate({ ...showGeneratePwdInitial, show: false })
    } catch (error) {
      setShowGenerate(showGeneratePwdInitial)
      showError('Could not generate password', false)
    }
  }

  return (
    <ObjectStorageUsersReservationsDetails
      reserveDetails={reserveDetails}
      activeTab={activeTab}
      tabDetails={tabDetails}
      actionsReserveDetails={actionsReserveDetails}
      showActionModal={showActionModal}
      actionModalContent={actionModalContent}
      tabs={tabsInitial}
      loading={loading}
      showGeneratePwd={showGeneratePwd}
      showAccessKey={showAccessKey}
      showSecretKey={showSecretKey}
      userBuckets={userBuckets}
      errorModalContent={errorModalContent}
      showErrorModal={showErrorModal}
      setShowErrorModal={setShowErrorModal}
      setShowActionModal={actionOnModal}
      setActiveTab={setActiveTab}
      setAction={setAction}
      generatePwd={generatePwd}
      copyItem={copyItem}
    />
  )
}

export default ObjectStorageUsersReservationsDetailsContainer
