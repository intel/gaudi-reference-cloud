// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import { BsPencilFill, BsTrash3, BsCopy } from 'react-icons/bs'
import BucketService from '../../services/BucketService'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useToastStore from '../../store/toastStore/ToastStore'
import ObjectStorageReservationsDetails from '../../components/objectStorage/objectStorageDetails/ObjectStorageReservationsDetails'
import { NavLink, useNavigate, useParams } from 'react-router-dom'
import useBucketStore from '../../store/bucketStore/BucketStore'
import { capitalizeString } from '../../utils/stringFormatHelper/StringFormatHelper'
import { useCopy } from '../../hooks/useCopy'
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

const getTabTitle = (text: string): JSX.Element => {
  let title = null
  switch (text) {
    case 'Users':
      title = (
        <>
          <NavLink
            to="/buckets/users"
            className="btn btn-outline-primary"
            intc-id={'btn-navigate-action-Manage-users-and-permissions'}
          >
            Manage buckets principals and permissions
          </NavLink>
        </>
      )
      break
    default:
      title = <> {text} </>
      break
  }

  return title
}

const getItemList = (items: string[]): JSX.Element[] => {
  return items.map((item: string, index: number) => {
    return (
      <div className="mt-1" key={index}>
        <span>{item}</span>
      </div>
    )
  })
}

const ObjectStorageReservationsDetailsContainer = (): JSX.Element => {
  // Column structure for security tab
  const userColumns = [
    {
      columnName: 'Login',
      targetColumn: 'name'
    },
    {
      columnName: 'Actions',
      targetColumn: 'actions',
      isSort: false
    },
    {
      columnName: 'Policies',
      targetColumn: 'permission',
      isSort: false
    },
    {
      columnName: 'Path',
      targetColumn: 'prefix',
      isSort: false
    }
  ]

  // Column structure for policy tab
  const policyColumns = [
    {
      columnName: 'Rule Name',
      targetColumn: 'ruleName'
    },
    {
      columnName: 'Prefix',
      targetColumn: 'prefix'
    },
    {
      columnName: 'Delete Marker',
      targetColumn: 'deleteMarker'
    },
    {
      columnName: 'Expiry Days',
      targetColumn: 'expireDays'
    },
    {
      columnName: 'Non Current Expiry Days',
      targetColumn: 'noncurrentExpireDays'
    },
    {
      columnName: 'Actions',
      targetColumn: 'actions'
    }
  ]

  const tabsInitial = [
    {
      label: 'Details',
      id: 'details',
      show: true
    },
    {
      label: 'Principals',
      id: 'principals',
      show: true
    },
    {
      label: 'Lifecycle Rules',
      id: 'lifecycleRules',
      show: true
    }
  ]

  const tabDetailsInitial: any = [
    {
      tapTitle: 'Bucket information',
      tapConfig: { type: 'columns', columnCount: 4 },
      fields: [
        { label: 'Bucket ID:', field: 'resourceId', value: '' },
        { label: 'Default Size:', field: 'storage', value: '' },
        { label: 'Description:', field: 'description', value: '' },
        { label: 'State:', field: 'status', value: '' },
        {
          label: 'Private Endpoint URL:',
          field: 'accessEndpoint',
          value: '',
          action: [{ func: copyItem, label: 'Copy', icon: 'Copy' }]
        },
        { label: 'Network Security Group', field: 'subnet', value: '' },
        { label: 'Versioning:', field: 'versioned', value: '' }
      ]
    },
    {
      tapTitle: getTabTitle('Users'),
      tapConfig: { type: 'custom' },
      customContent: null
    },
    {
      tapTitle: '',
      tapConfig: { type: 'custom' },
      customContent: null
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

  const ruleActionsOptions = [
    {
      id: 'edit',
      name: getActionItemLabel('Edit'),
      status: ['Ready'],
      label: 'Edit Rule'
    },
    {
      id: 'terminate',
      name: getActionItemLabel('Delete'),
      status: ['Ready', 'Provisioning', 'Failed'],
      label: 'Delete rule',
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
    actionType: '',
    name: ''
  }

  const errorModal = {
    titleMessage: '',
    description: '',
    message: ''
  }

  const throwError = useErrorBoundary()
  const [reserveDetails, setReserveDetails] = useState<any>(null)
  const [reserveUser, setReserveUser] = useState<any>(null)
  const [reserveLifeCyclePolicies, setReserveLifeCyclePolicies] = useState<any>(null)
  const [actionsReserveDetails, setActionsReserveDetails] = useState<any[]>([])
  const [showActionModal, setShowActionModal] = useState(false)
  const [actionModalContent, setActionModalContent] = useState(modalContent)
  const [tabDetails, setTabDetails] = useState(tabDetailsInitial)
  const [afterActionShowModal, setAfterActionShowModal] = useState(false)
  const [afterActionModalContent, setAfterActionModalContent] = useState(modalContent)
  const [isPageReady, setIsPageReady] = useState(false)
  const [showErrorModal, setShowErrorModal] = useState(false)
  const [errorModalContent, setErrorModalContent] = useState(errorModal)
  // Global State
  const objectStorages = useBucketStore((state) => state.objectStorages)
  const loading = useBucketStore((state) => state.loading)
  const setObjectStorages = useBucketStore((state) => state.setObjectStorages)
  const setShouldRefreshObjectStorages = useBucketStore((state) => state.setShouldRefreshObjectStorages)
  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)

  const setCurrentSelectedBucket = useBucketStore((state) => state.setCurrentSelectedBucket)
  const bucketActiveTab = useBucketStore((state) => state.bucketActiveTab)
  const setBucketActiveTab = useBucketStore((state) => state.setBucketActiveTab)
  const { copyToClipboard } = useCopy()

  const refreshStorages = async (background: boolean): Promise<void> => {
    try {
      await setObjectStorages(background)
    } catch (error: any) {
      if (isErrorInAuthorization(error)) {
        setShowErrorModal(true)
        const content = { ...errorModal }
        content.message = error.response.data.message
        setErrorModalContent(content)
      } else throwError(error)
    }
  }

  function copyItem(value: string): void {
    copyToClipboard(value)
  }

  // Navigation
  const navigate = useNavigate()
  const { param: name } = useParams()

  // Hooks
  useEffect(() => {
    const fetch = async (): Promise<void> => {
      if (objectStorages?.length === 0) await refreshStorages(false)
      setIsPageReady(true)
    }
    fetch().catch((error) => {
      throwError(error)
    })
    setShouldRefreshObjectStorages(true)
    return () => {
      setShouldRefreshObjectStorages(false)
    }
  }, [])

  useEffect(() => {
    updateDetails()
  }, [objectStorages, isPageReady])

  // functions
  const updateDetails = (): void => {
    const storage = objectStorages.find((instance) => instance.name === name)
    if (storage === undefined) {
      if (isPageReady) navigate('/buckets')
      setActionsReserveDetails([])
      setReserveUser(null)
      setReserveLifeCyclePolicies(null)
      setCurrentSelectedBucket(null)
      setReserveDetails(null)
      return
    }

    const storageDetail: any = { ...storage }

    const tabDetailsCopy = []
    for (const tapIndex in tabDetails) {
      const tapDetail = { ...tabDetails[tapIndex] }
      const updateFields = []
      for (const index in tapDetail.fields) {
        const field = { ...tapDetail.fields[index] }
        field.value = String(storageDetail[field.field])
        updateFields.push(field)
      }
      tapDetail.fields = updateFields
      tabDetailsCopy.push(tapDetail)
    }

    setTabDetails(tabDetailsCopy)
    setActionsReserveDetails(getActionsByStatus(storageDetail.status, actionsOptions))
    setCurrentSelectedBucket(storageDetail)
    setReserveDetails(storageDetail)
    const userGrid = []
    const userAccess = storageDetail.userAccessPolicies
    for (const index in userAccess) {
      const user = { ...userAccess[index] }
      userGrid.push({
        name: {
          showField: true,
          type: 'hyperlink',
          noHyperLinkValue: user.name,
          value: '',
          icon: <BsCopy className="text-primary" size="20" />,
          function: () => {
            copyItem(user.name)
          }
        },
        actions: getItemList(user.spec.actions),
        permission: getItemList(user.spec.permission),
        path: user.spec.prefix ? user.spec.prefix : '/'
      })
    }
    setReserveUser(userGrid)

    const ruleGrid = []
    const lifecycleRules = storageDetail.lifecycleRulePolicies
    for (const index in lifecycleRules) {
      const rule = { ...lifecycleRules[index] }
      ruleGrid.push({
        ruleName: rule.ruleName,
        prefix: rule.prefix ? rule.prefix : '/',
        deleteMarker: capitalizeString(rule.deleteMarker.toString()),
        expireDays: rule.expireDays,
        noncurrentExpireDays: rule.noncurrentExpireDays,
        actions: {
          showField: true,
          type: 'Buttons',
          value: rule,
          selectableValues: getActionsByStatus(rule.status, ruleActionsOptions),
          function: setRuleAction
        }
      })
    }
    setReserveLifeCyclePolicies(ruleGrid)
  }

  function setAction(action: any, item: any): void {
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

  function setRuleAction(action: any, item: any): void {
    switch (action.id) {
      case 'edit':
        navigate({
          pathname: `/buckets/d/${name}/lifecyclerule/e/${item.ruleName}`
        })
        break
      case 'terminate': {
        const copyModalContent = { ...modalContent }
        copyModalContent.label = action.label
        copyModalContent.instanceName = item.ruleName
        copyModalContent.buttonLabel = action.buttonLabel
        copyModalContent.resourceId = item.resourceId
        copyModalContent.actionType = 'terminateRule'
        copyModalContent.name = item.ruleName
        setActionModalContent(copyModalContent)
        setShowActionModal(true)
        break
      }
      default:
        break
    }
  }

  function getActionsByStatus(status: string, options: any[]): any[] {
    const result = []

    for (const index of options) {
      const option = { ...index }
      if (option.status.find((item: string) => item === status)) {
        result.push(option)
      }
    }

    return result
  }

  const deleteBucket = async (resourceId: string): Promise<void> => {
    try {
      await BucketService.deleteObjectBucketByCloudAccount(resourceId)
      setTimeout(() => {
        refreshStorages(true).catch((error) => {
          throwError(error)
        })
      }, 1000)
      showSuccess('Bucket deleted successfully.', false)
      const copyModalContent = { ...modalContent }
      copyModalContent.label = 'Deleted bucket principals'
      copyModalContent.feedback =
        'Your bucket was deleted. The associated principals may still have policies and permissions associated with the deleted bucket name.'
      copyModalContent.buttonLabel = 'Ok'
      setAfterActionModalContent(copyModalContent)
      setAfterActionShowModal(true)
      setTimeout(() => {
        navigate('/buckets')
      }, 2000)
    } catch (error: any) {
      if (isErrorInAuthorization(error)) {
        setShowErrorModal(true)
        const content = { ...errorModal }
        content.message = error.response.data.message
        setErrorModalContent(content)
      } else showError('Error while deleting the bucket. Please try again later.', false)
    }
  }

  const deleteRule = async (resourceId: string, bucketId: string): Promise<void> => {
    try {
      await BucketService.deleteObjectBucketRule(resourceId, bucketId)
      if (reserveDetails.name === actionModalContent.instanceName) {
        setCurrentSelectedBucket(null)
        setBucketActiveTab(0)
      }
      setTimeout(() => {
        refreshStorages(true).catch((error) => {
          throwError(error)
        })
      }, 1000)
      showSuccess('Rule deleted successfully.', false)
      // In case terminating state does not show inmmediately
      setTimeout(() => {
        refreshStorages(true).catch((error) => {
          throwError(error)
        })
      }, 5000)
    } catch (error: any) {
      if (isErrorInAuthorization(error)) {
        setShowErrorModal(true)
        const content = { ...errorModal }
        content.message = error.response.data.message
        setErrorModalContent(content)
      } else showError('Error while deleting the rule. Please try again later.', false)
    }
  }

  async function actionOnModal(result: boolean): Promise<void> {
    setShowActionModal(result)
    if (result) {
      if (actionModalContent.actionType === 'terminateBucket') {
        await deleteBucket(actionModalContent.resourceId)
      } else {
        await deleteRule(actionModalContent.resourceId, reserveDetails.resourceId)
      }
      setShowActionModal(false)
    }
  }

  return (
    <ObjectStorageReservationsDetails
      reserveDetails={reserveDetails}
      bucketActiveTab={bucketActiveTab}
      tabDetails={tabDetails}
      actionsReserveDetails={actionsReserveDetails}
      showActionModal={showActionModal}
      actionModalContent={actionModalContent}
      tabs={tabsInitial}
      loading={loading}
      setShowActionModal={actionOnModal}
      setBucketActiveTab={setBucketActiveTab}
      setAction={setAction}
      userColumns={userColumns}
      reserveUser={reserveUser}
      policyColumns={policyColumns}
      reserveLifeCyclePolicies={reserveLifeCyclePolicies}
      afterActionModalContent={afterActionModalContent}
      afterActionShowModal={afterActionShowModal}
      setAfterActionShowModal={setAfterActionShowModal}
      errorModalContent={errorModalContent}
      showErrorModal={showErrorModal}
      setShowErrorModal={setShowErrorModal}
    />
  )
}

export default ObjectStorageReservationsDetailsContainer
