// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import useCloudAccountStore from '../../store/cloudAccountStore/CloudAccountStore'
import { BsPencilFill, BsTrash3 } from 'react-icons/bs'
import CloudAccountService from '../../services/CloudAccountService'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import { useCopy } from '../../hooks/useCopy'
import useToastStore from '../../store/toastStore/ToastStore'
import { useNavigate, useParams } from 'react-router'
import { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'
import StorageReservationsDetails from '../../components/storage/StorageReservationsDetails/StorageReservationsDetails'
import SpinnerIcon from '../../utils/spinner/SpinnerIcon'
import { isErrorInAuthorization } from '../../utils/apiError/apiError'
import { capitalizeString } from '../../utils/stringFormatHelper/StringFormatHelper'

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

const StorageReservationsDetailsContainer = (): JSX.Element => {
  // local state
  const tabsInitial: any = [
    {
      label: 'Details',
      id: 'details',
      show: true
    },
    {
      label: 'Security',
      id: 'security',
      show: false
    },
    {
      label: 'Security',
      id: 'securityVast',
      show: false
    }
  ]

  const tabDetailsInitial: any = [
    {
      tapTitle: 'Volume information',
      tapConfig: { type: 'columns', columnCount: 3 },
      fields: [
        { label: 'Volume ID:', field: 'resourceId', value: '' },
        { label: 'Size:', field: 'storage', value: '' },
        { label: 'State:', field: 'status', value: '' },
        { label: 'Cluster', field: 'mountClusterName', value: '' },
        { label: 'Encryption', field: 'encryptedText', value: '' },
        { label: 'Namespace:', field: 'mountNamespace', value: '' }
      ],
      show: true
    },
    {
      tapTitle: 'Volume credentials',
      tapConfig: { type: 'custom' },
      fields: [
        {
          label: 'User:',
          field: 'user',
          value: '',
          maske: false,
          action: [{ func: copyItem, label: 'Copy', icon: 'Copy' }]
        },
        {
          label: 'Password:',
          field: 'password',
          value: '',
          mask: true,
          action: [{ func: copyItem, label: 'Copy', icon: 'Copy' }]
        }
      ],
      show: !isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_STORAGE_VAST)
    },
    {
      tapTitle: 'Volume IP range',
      tapConfig: { type: 'columns', columnCount: 3 },
      fields: [{ label: 'Allowed IP Range:', field: 'securityGroup', value: '' }],
      show: isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_STORAGE_VAST)
    }
  ]

  const actionsOptions: any = [
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

  const showGeneratePwdInitial: any = {
    show: true,
    label: 'Generate password',
    icon: null
  }

  const errorModal = {
    titleMessage: '',
    description: '',
    message: ''
  }

  const throwError = useErrorBoundary()

  // Navigation
  const navigate = useNavigate()
  const { param: name } = useParams()

  const [reserveDetails, setReserveDetails] = useState<any>(null)
  const [actionsReserveDetails, setActionsReserveDetails] = useState<any[]>([])
  const [activeTab, setActiveTab] = useState(0)
  const [showHowToMountModal, setShowHowToMountModal] = useState(false)
  const [showHowToUnmountModal, setShowHowToUnmountModal] = useState(false)
  const [showActionModal, setShowActionModal] = useState(false)
  const [actionModalContent, setActionModalContent] = useState(modalContent)
  const [tabDetails, setTabDetails] = useState(tabDetailsInitial)
  const [tabs, setTabs] = useState(tabsInitial)
  const [showGeneratePwd, setShowGenerate] = useState(showGeneratePwdInitial)
  const [userProfile, setUserProfile] = useState(null)
  const [isPageReady, setIsPageReady] = useState(false)
  const [showErrorModal, setShowErrorModal] = useState(false)
  const [errorModalContent, setErrorModalContent] = useState(errorModal)

  // Global State
  const storages = useCloudAccountStore((state) => state.storages)
  const loading = useCloudAccountStore((state) => state.loading)
  const setStorages = useCloudAccountStore((state) => state.setStorages)
  const setShouldRefreshStorages = useCloudAccountStore((state) => state.setShouldRefreshStorages)
  const { setInstances, setInstanceGroups } = useCloudAccountStore((state) => state)
  const showError = useToastStore((state) => state.showError)
  const { copyToClipboard } = useCopy()

  const refreshStorages = async (background: boolean): Promise<void> => {
    try {
      await setStorages(background)
    } catch (error) {
      throwError(error)
    }
  }

  // Hooks
  useEffect(() => {
    const fetch = async (): Promise<void> => {
      if (!storages || storages.length === 0) await refreshStorages(false)
      await setInstances(true)
      await setInstanceGroups(true)
      setIsPageReady(true)
    }
    fetch().catch((error) => {
      throwError(error)
    })

    setShouldRefreshStorages(true)
    return () => {
      setShouldRefreshStorages(false)
    }
  }, [])

  useEffect(() => {
    updateDetails()
  }, [storages, isPageReady])

  useEffect(() => {
    UpdateTapDetails()
  }, [userProfile])

  const UpdateTapDetails = (): void => {
    if (userProfile != null) {
      const tabDetailsCopy = []
      for (const tapIndex in tabDetails) {
        const tapDetail = { ...tabDetails[tapIndex] }
        const updateFields = []
        if (tapDetail.tapTitle === 'Volume credentials') {
          for (const index in tapDetail.fields) {
            const field = { ...tapDetail.fields[index] }
            field.value = userProfile[field.field]
            updateFields.push(field)
          }
          tapDetail.fields = updateFields
        }
        tabDetailsCopy.push(tapDetail)
      }
      setTabDetails(tabDetailsCopy)
    }
  }

  // functions
  const updateDetails = (): void => {
    const storage = storages.find((instance) => instance.name === name)
    if (storage === undefined) {
      if (isPageReady) navigate('/storage')
      setActionsReserveDetails([])
      setReserveDetails(null)
      return
    }

    const storageDetail: any = {
      ...storage,
      encryptedText: storage.encrypted ? 'Enabled' : 'Disabled'
    }

    const tabDetailsCopy = []
    for (const tapIndex in tabDetails) {
      const tapDetail = { ...tabDetails[tapIndex] }
      const updateFields = []
      for (const index in tapDetail.fields) {
        const field = { ...tapDetail.fields[index] }
        if (field.field === 'securityGroup') {
          const securityGroups = storageDetail.securityGroup
          let gateways = ''
          if (securityGroups.length > 0) {
            gateways = securityGroups
              .map((group: any) => String(group.subnet) + '/' + String(group.prefixLength))
              .join('- ')
          }
          field.value = gateways
        } else {
          field.value = storageDetail[field.field]
        }
        updateFields.push(field)
      }
      tapDetail.fields = updateFields

      if (tapDetail.show) {
        tabDetailsCopy.push(tapDetail)
      }
    }

    // Hide security tab when the status is failed
    const tabsUpdated = []
    for (const index in tabs) {
      const tab = tabs[index]
      if (tab.label === 'Security') {
        const isVastEnabled = isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_STORAGE_VAST)
        tab.show =
          ((tab.id === 'security' && !isVastEnabled) || (tab.id === 'securityVast' && isVastEnabled)) &&
          !(
            storageDetail.status === 'Failed' ||
            storageDetail.status === 'Provisioning' ||
            storageDetail.status === 'Deleting'
          )
      }
      if (!tab.show) {
        continue
      }
      tabsUpdated.push(tab)
    }

    setTabs(tabsUpdated)
    setTabDetails(tabDetailsCopy)
    setActionsReserveDetails(getActionsByStatus(storageDetail.status))
    setReserveDetails(storageDetail)
  }

  function setAction(action: any, item: any): void {
    switch (action.id) {
      case 'edit':
        navigate({
          pathname: `/storage/d/${item.name}/edit`,
          search: '?backTo=detail'
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

  const deleteStorage = async (resourceId: string): Promise<void> => {
    try {
      await CloudAccountService.deleteStorageByCloudAccount(resourceId)
      setTimeout(() => {
        refreshStorages(true).catch((error) => {
          throwError(error)
        })
      }, 1000)
      setTimeout(() => {
        navigate('/storage')
      }, 2000)
    } catch (error: any) {
      let errorMessage = ''
      if (error.response) {
        errorMessage = error.response.data.message
        showError(capitalizeString(errorMessage), false)
      } else {
        throwError(error)
      }
    }
  }

  function actionOnModal(result: boolean): void {
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

  function copyItem(value: string): void {
    copyToClipboard(value)
  }

  async function generatePwd(): Promise<void> {
    setShowGenerate({
      ...showGeneratePwdInitial,
      label: 'Generating password',
      icon: <SpinnerIcon />
    })

    try {
      const response = await CloudAccountService.getStorageCredentialsByCloudAccount(reserveDetails.name)
      const data = response.data
      setUserProfile(data)
      setShowGenerate({ ...showGeneratePwdInitial, show: false })
    } catch (error) {
      setShowGenerate(showGeneratePwdInitial)
      showError('Could not generate password', false)
    }
  }

  return (
    <StorageReservationsDetails
      reserveDetails={reserveDetails}
      activeTab={activeTab}
      tabDetails={tabDetails}
      actionsReserveDetails={actionsReserveDetails}
      showHowToMountModal={showHowToMountModal}
      showHowToUnmountModal={showHowToUnmountModal}
      showActionModal={showActionModal}
      actionModalContent={actionModalContent}
      tabs={tabs}
      loading={loading}
      showGeneratePwd={showGeneratePwd}
      userProfile={userProfile}
      errorModalContent={errorModalContent}
      showErrorModal={showErrorModal}
      setShowErrorModal={setShowErrorModal}
      setShowHowToMountModal={setShowHowToMountModal}
      setShowHowToUnmountModal={setShowHowToUnmountModal}
      setShowActionModal={actionOnModal}
      setActiveTab={setActiveTab}
      setAction={setAction}
      generatePwd={generatePwd}
    />
  )
}

export default StorageReservationsDetailsContainer
