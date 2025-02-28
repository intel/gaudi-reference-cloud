// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import InstancesView from '../../components/instanceManagement/InstancesView'
import useInstancesStore from '../../store/instancesStore/InstancesStore'
import useInstanceGroupsStore from '../../store/instancesStore/instanceGroupStore'
import CloudAccountService from '../../services/CloudAccountService'
import moment from 'moment'
import { BiXCircle } from 'react-icons/bi'
import useToastStore from '../../store/toastStore/ToastStore'

const dateFormat = 'MM/DD/YYYY hh:mm a'

const TerminateInstanceContainer = () => {
  // State to store the search query and results
  const [cloudAccount, setCloudAccount] = useState(null)
  const [cloudAccountId, setCloudAccountId] = useState(null)
  const [instances, setInstances] = useState([])
  const [instanceGroups, setInstanceGroups] = useState([])
  const [selectedCloudAccount, setSelectedCloudAccount] = useState(null)
  const [cloudAccountError, setCloudAccountError] = useState(false)

  const instancesColumns = [
    {
      columnName: 'Instance Name',
      targetColumn: 'instanceName'
    },
    {
      columnName: 'IP',
      targetColumn: 'ip'
    },
    {
      columnName: 'Instance Type',
      targetColumn: 'instanceType'
    },
    {
      columnName: 'Created at',
      targetColumn: 'createdAt'
    },
    {
      columnName: 'Status',
      targetColumn: 'status'
    },
    {
      columnName: 'Actions',
      targetColumn: 'actions',
      columnConfig: {
        behaviorType: 'buttons',
        behaviorFunction: null
      }
    }
  ]

  const instanceGroupsColumns = [
    {
      columnName: 'Instance Group Name',
      targetColumn: 'instanceGroupName'
    },
    {
      columnName: 'Instance Type',
      targetColumn: 'instanceType'
    },
    {
      columnName: 'Instance Count',
      targetColumn: 'instanceCount'
    },
    {
      columnName: 'Ready Count',
      targetColumn: 'readyCount'
    },
    {
      columnName: 'Actions',
      targetColumn: 'actions',
      columnConfig: {
        behaviorType: 'buttons',
        behaviorFunction: null
      }
    }
  ]

  // Initial state for confirm modal
  const initialInstanceConfirmData = {
    title: 'Terminate instance',
    instanceName: '',
    cloudAccount: '',
    resourceId: '',
    onClose: closeConfirmModal,
    isShow: false,
    action: 'Terminate',
    isValidInput: 'is-invalid'
  }

  const [showLoader, setShowLoader] = useState('')
  const [instancesModalData, setInstancesModalData] = useState(initialInstanceConfirmData)

  // Global State for Instances
  const loadingInstances = useInstancesStore((state) => state.loading)
  const instancesState = useInstancesStore((state) => state.instances)
  const setInstancesStore = useInstancesStore((state) => state.setInstances)
  const setShouldRefreshInstances = useInstancesStore((state) => state.setShouldRefreshInstances)

  // Global State for Instance Groups
  const loadingInstanceGroup = useInstanceGroupsStore((state) => state.loading)
  const instancesGroupState = useInstanceGroupsStore((state) => state.instanceGroups)
  const setInstanceGroupsStore = useInstanceGroupsStore((state) => state.setInstanceGroups)
  const setShouldRefreshInstanceGroups = useInstanceGroupsStore((state) => state.setShouldRefreshInstanceGroups)

  // global functions for toast
  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)

  useEffect(() => {
    if (cloudAccountId != null) {
      setInstancesGridInfo()
      setInstanceGroupsGridInfo()

      setShouldRefreshInstances(true)
      setShouldRefreshInstanceGroups(true)
      return () => {
        setShouldRefreshInstances(false)
        setShouldRefreshInstanceGroups(false)
      }
    }
  }, [instancesState, instancesGroupState])

  // Set instances grid.
  function setInstancesGridInfo() {
    const instancesGridInfo = []
    // Initializing the states to verify the status.
    for (const index in instancesState.items) {
      const inst = { ...instancesState.items }

      instancesGridInfo.push({
        instanceName: inst[index]?.metadata?.name,
        ip: inst[index]?.status?.interfaces[0]?.addresses[0],
        instanceType: inst[index]?.spec?.instanceType,
        createdAt: moment(inst[index]?.metadata?.creationTimestamp).format(dateFormat),
        status: inst[index]?.status?.phase,
        actions: {
          showField: true,
          type: 'Buttons',
          value: inst[index],
          selectableValues: [
            { name: getActionItemLabel('Delete'), label: 'Delete', action: 'Terminate' }
          ],
          function: setAction
        }
      })
    }
    setInstances(instancesGridInfo)
  }

  // Set instance groups grid.
  function setInstanceGroupsGridInfo() {
    const instanceGroupGridInfo = []
    for (const index in instancesGroupState.items) {
      const instG = { ...instancesGroupState.items }
      const instanceCountVar = instG[index]?.spec?.instanceCount
      const readyCountVar = instG[index]?.status?.readyCount

      instanceGroupGridInfo.push({
        instanceName: instG[index]?.metadata?.name,
        instanceType: instG[index]?.spec?.instanceSpec?.instanceType,
        instanceCount: instanceCountVar,
        status: readyCountVar,
        actions: {
          showField: true,
          type: 'Buttons',
          value: instG[index],
          selectableValues: [
            { name: getActionItemLabel('Delete'), label: 'Delete', action: 'Terminate' }
          ],
          function: setAction
        }
      })
    }
    setInstanceGroups(instanceGroupGridInfo)
  }

  const handleSearchInputChange = (e) => {
    // Update the state with the numeric value
    setCloudAccount(e.target.value)
  }

  const onClearSearchInput = () => {
    // Making cloud account null.
    setCloudAccount('')
    // Clearing / Making empty instances state.
    setInstances([])
    // Clearing / Making empty instance groups state.
    setInstanceGroups([])
    // Clearing / Making empty instance group state.
    setSelectedCloudAccount('')
    // Making refresh instance flag to FALSE.
    setShouldRefreshInstances(false)
    // Making refresh instance group to FALSE.
    setShouldRefreshInstanceGroups(false)
  }

  // Function to handle form submission
  const handleSubmit = async (e) => {
    setCloudAccountError('')
    setSelectedCloudAccount(null)
    if (cloudAccount !== '') {
      setShowLoader({ isShow: true, message: 'Searching for Details...' })
      try {
        let data = null
        if (cloudAccount.includes('@') || /[a-zA-Z]/.test(cloudAccount)) {
          data = await CloudAccountService.getCloudAccountDetailsByName(cloudAccount)
        } else {
          data = await CloudAccountService.getCloudAccountDetailsById(cloudAccount)
        }
        // Set Cloud Account ID.
        setCloudAccountId(data?.data?.id)
        // Calls setInstancesStore to fetch instances.
        await setInstancesStore(data?.data?.id)
        // Calls setInstanceGroupsStore to fetch instance groups.
        await setInstanceGroupsStore(data?.data?.id)
        // Picks the selected searched data.
        setSelectedCloudAccount(data?.data)
        // Making error state with false.
        setCloudAccountError(false)
      } catch (e) {
        const code = e.response.data?.code
        const errorMsg = e.response.data?.message
        const message = code && [3, 5].includes(code) ? errorMsg.charAt(0).toUpperCase() + errorMsg.slice(1) : 'Cloud Account ID is not found'
        // Assigning the error message.
        setCloudAccountError(message)
        // Clearing selected search data.
        setSelectedCloudAccount(false)
        // Clearing / Making empty instances state.
        setInstances([])
        // Clearing / Making empty instance group state.
        setInstanceGroups([])
        // Stops the instance calls.
        setShouldRefreshInstances(false)
        // Stops the instance group calls.
        setShouldRefreshInstanceGroups(false)
      }
      setShowLoader('')
    } else {
      // Assigning the error message.
      setCloudAccountError('Cloud Account is required')
      // Clearing / Making empty instances state.
      setInstances([])
      // Clearing / Making empty instance group state.
      setInstanceGroups([])
      // Making loader empty.
      setShowLoader('')
      // Stops the instance calls.
      setShouldRefreshInstances(false)
      // Stops the instance group calls.
      setShouldRefreshInstanceGroups(false)
    }
  }

  const getActionItemLabel = (text) => {
    return (
      <>
        {' '}
        <BiXCircle className="mb-1" /> {text}{' '}
      </>
    )
  }

  function setAction(item, value) {
    // Making sure there is no error while we open Terminate Instance modal.
    const instance = value
    setInstancesModalData({
      title: `Terminate Instance${!instance?.metadata?.resourceId ? ' Group' : ''}`,
      subtitle: `Are you sure you want to terminate this Instance${!instance?.metadata?.resourceId ? ' Group' : ''}?`,
      message: `To confirm termination, please enter the Instance${!instance?.metadata?.resourceId ? ' Group' : ''} Name in the input field.`,
      type: !instance?.metadata?.resourceId ? ' Group' : '',
      name: instance?.metadata?.name,
      cAccount: cloudAccount,
      cAccountId: cloudAccountId,
      resourceId: instance?.metadata?.resourceId,
      isShow: true,
      onClose: closeConfirmModal,
      action: 'Terminate',
      isValidInput: 'is-invalid'
    })
  }

  const instancesEmptyGrid = {
    title: 'No instances found',
    subTitle: 'Please search with Cloud Accound ID'
  }

  const instanceGroupsEmptyGrid = {
    title: 'No instance groups found',
    subTitle: 'Please search with Cloud Accound ID'
  }

  // Navigation
  const navigate = useNavigate()

  function backToHome() {
    // Clearing all the avaliable states.
    setCloudAccount('')
    setInstances([])
    setInstanceGroups([])
    setSelectedCloudAccount('')
    setShouldRefreshInstances(false)
    setShouldRefreshInstanceGroups(false)
    navigate('/')
  }

  function closeConfirmModal() {
    setInstancesModalData(initialInstanceConfirmData)
  }

  const onInstanceTerminateSubmit = async (name, resourceId, cloudAccount) => {
    setShowLoader({ isShow: true, message: 'Working on your request' })
    confirmOnInstanceTerminateSubmit(name, resourceId, cloudAccount)
  }

  const confirmOnInstanceTerminateSubmit = async (name, resourceId, cloudAccount) => {
    try {
      // Considering its and Instance when we have resoureceId available. Else, its an Instance Group.
      if (resourceId) {
        CloudAccountService.deleteComputeReservation(resourceId, cloudAccount).then(async () => {
          setShowLoader({ isShow: true, message: 'Working on your request' })
          // eslint-disable-next-line promise/param-names
          await new Promise(f => setTimeout(f, 6000))
          setShouldRefreshInstances(true)
          await setInstancesStore(cloudAccount)
          closeConfirmModal()
          setShowLoader({ isShow: false })
          showSuccess('Instance is going to terminate in few seconds.')
        }).catch(() => {
          showError('Service is unavailable, Please try after sometime.')
        })
      } else {
        CloudAccountService.deleteInstanceGroupByName(name, cloudAccount).then(async () => {
          setShowLoader({ isShow: true, message: 'Working on your request' })
          // eslint-disable-next-line promise/param-names
          await new Promise(f => setTimeout(f, 4000))
          setShouldRefreshInstances(true)
          await setInstanceGroupsStore(cloudAccount)
          closeConfirmModal()
          setShowLoader({ isShow: false })
          showSuccess('Instance group is going to terminate in few seconds.')
        }).catch(() => {
          showError('Service is unavailable, Please try after sometime.')
        })
      }
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
      showError(message)
    }
    setShowLoader({ isShow: false })
  }

  return (
    <InstancesView
      instances={instances}
      instanceGroups={instanceGroups}
      showLoader={showLoader}
      instancesColumns={instancesColumns}
      instanceGroupsColumns={instanceGroupsColumns}
      loadingInstances={loadingInstances}
      loadingInstanceGroup={loadingInstanceGroup}
      instancesEmptyGrid={instancesEmptyGrid}
      instanceGroupsEmptyGrid={instanceGroupsEmptyGrid}
      cloudAccount={cloudAccount}
      backToHome={backToHome}
      handleSearchInputChange={handleSearchInputChange}
      onClearSearchInput={onClearSearchInput}
      handleSubmit={handleSubmit}
      instancesModalData={instancesModalData}
      onInstanceTerminateSubmit={onInstanceTerminateSubmit}
      selectedCloudAccount={selectedCloudAccount}
      cloudAccountError={cloudAccountError}
    />
  )
}

export default TerminateInstanceContainer
