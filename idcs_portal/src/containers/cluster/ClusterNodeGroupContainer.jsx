// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import ClusterNodeGroup from '../../components/cluster/clusterMyReservations/ClusterNodeGroup'
import ClusterService from '../../services/ClusterService'
import { iksNodeGroupActionsEnum } from '../../utils/Enums'

import {
  friendlyErrorMessages,
  getErrorMessageFromCodeAndMessage,
  isErrorInsufficientCapacity
} from '../../utils/apiError/apiError'

const ClusterNodeGroupContainer = ({
  nodeGroup,
  getActionItemLabel,
  debounceClusterRefresh,
  isClusterInActiveState
}) => {
  const errorModalInit = {
    showErrorModal: false,
    errorDescription: null,
    errorMessage: null,
    errorHideRetryMessage: true
  }

  const [nodes, setNodes] = useState([])
  const [selectedNodeDetails, setSelectedNodeDetails] = useState(null)
  const [showHowToConnectModal, setShowHowToConnectModal] = useState(false)
  const [showConfirmationModal, setShowConfirmationModal] = useState(false)
  const [confirmationModalAction, setConfirmationModalAction] = useState(null)
  const [errorModal, setErrorModal] = useState(errorModalInit)
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(undefined)

  const columns = [
    {
      columnName: 'Node name',
      targetColumn: 'name'
    },
    {
      columnName: 'State',
      targetColumn: 'status'
    },
    {
      columnName: 'IP',
      targetColumn: 'ipNbr',
      className: 'text-end'
    },

    {
      columnName: 'IMI',
      targetColumn: 'imi'
    },
    {
      columnName: 'Storage Status',
      targetColumn: 'storagestatus'
    },
    {
      columnName: 'Created at',
      targetColumn: 'creationTimestamp'
    },
    {
      columnName: 'Actions',
      hideField: true,
      targetColumn: 'actions'
    }
  ]

  const actionsOptions = [
    {
      id: 'sshconnect',
      name: getActionItemLabel('Connect via SSH'),
      status: ['Active', 'Updating', 'Failed'],
      label: 'Connect via SSH',
      buttonLabel: 'Connect via SSH'
    }
  ]

  const emptyGridByFilter = {
    title: 'No nodes found',
    subTitle: 'The applied filter criteria did not match any items.',
    action: {
      type: 'function',
      href: () => setFilter('', true),
      label: 'Clear filters'
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

  function getNodeStatusMessage(node) {
    let message = 'No Status'
    if (node) {
      if (node.errorcode) {
        message = node.message ? getErrorMessageFromCodeAndMessage(node.errorcode, node.message) : 'No Status'
      } else {
        message = node.message ? node.message : 'No Status'
      }
    }
    return message
  }

  function getStatusInfo(node) {
    return getActionItemLabel(node.state, getNodeStatusMessage(node))
  }

  function setAction(action, item) {
    switch (action.id) {
      case 'sshconnect': {
        const instanceDetail = {
          status: item.state,
          ipNbr: item.ipaddress,
          sshProxyUser: '',
          sshProxyAddress: '',
          instanceCategory: nodeGroup.instanceTypeDetails.instanceCategory
        }
        setSelectedNodeDetails(instanceDetail)
        setShowHowToConnectModal(true)
        break
      }
      default:
        break
    }
  }

  const getNodesGridInfo = () => {
    if (!nodeGroup.nodes || nodeGroup.nodes.length === 0) {
      setNodes([])
      return
    }
    const gridInfo = []

    for (const index in nodeGroup.nodes) {
      const instance = { ...nodeGroup.nodes[index] }

      gridInfo.push({
        name: instance.name,
        status: {
          showField: true,
          type: 'function',
          value: instance,
          sortValue: instance.state,
          function: getStatusInfo
        },
        ipNbr: instance.ipaddress,
        imi: instance.imi,
        storagestatus: instance.wekaStorage
          ? instance.wekaStorage.status === ''
            ? 'No Status'
            : instance.wekaStorage.status
          : 'No Status',
        creationTimestamp: {
          showField: true,
          type: 'date',
          toLocalTime: true,
          value: instance.createddate,
          format: 'MM/DD/YYYY h:mm a'
        },
        actions: {
          showField: false,
          type: 'Buttons',
          value: instance,
          selectableValues: getActionsByStatus(instance.state, actionsOptions),
          function: setAction
        }
      })
    }
    setNodes(gridInfo)
  }

  const getNodegroupStatusMessage = (nodegroupstatus) => {
    let message = 'No Status'
    if (nodegroupstatus) {
      if (nodegroupstatus.state === 'Error' && nodegroupstatus.errorcode) {
        message = nodegroupstatus.message
          ? getErrorMessageFromCodeAndMessage(nodegroupstatus.errorcode, nodegroupstatus.message)
          : 'No Status'
      } else {
        message = nodegroupstatus.message ? nodegroupstatus.message : 'No Status'
      }
    }
    return message
  }

  const canUpdateNodeGroup = () => {
    return (
      isClusterInActiveState &&
      (nodeGroup.nodegroupstate === 'Active' ||
        nodeGroup.nodegroupstate === 'Error' ||
        nodeGroup.nodegroupstate === 'Updating')
    )
  }

  const openDeleteNodeGroup = (event) => {
    event.preventDefault()
    event.stopPropagation()
    setConfirmationModalAction(iksNodeGroupActionsEnum.deleteNodeGroup)
    setShowConfirmationModal(true)
    return false
  }

  const openUpgradeNodeImage = () => {
    setConfirmationModalAction(iksNodeGroupActionsEnum.upgradeImage)
    setShowConfirmationModal(true)
  }

  const openAddNode = () => {
    setConfirmationModalAction(iksNodeGroupActionsEnum.addNode)
    setShowConfirmationModal(true)
  }

  const openRemoveNode = () => {
    setConfirmationModalAction(iksNodeGroupActionsEnum.removeNode)
    setShowConfirmationModal(true)
  }

  const deleteNodeGroup = async () => {
    try {
      await ClusterService.deleteNodeGroup(nodeGroup.clusteruuid, nodeGroup.nodegroupuuid)
      setShowConfirmationModal(false)
      debounceClusterRefresh()
    } catch (error) {
      setShowConfirmationModal(false)
      createErrorModal(error, 'Could not delete your node group')
    }
  }

  const upgradeNodeGroup = async () => {
    try {
      await ClusterService.upgradeNodeGroup(nodeGroup.nodegroupuuid, nodeGroup.clusteruuid)
      setShowConfirmationModal(false)
      debounceClusterRefresh()
    } catch (error) {
      setShowConfirmationModal(false)
      createErrorModal(error, 'Could not upgrade your node group')
    }
  }

  const createErrorModal = (error, errorTitle) => {
    const newErrorModal = { ...errorModalInit }
    newErrorModal.errorHideRetryMessage = false
    newErrorModal.errorDescription = ''
    newErrorModal.titleMessage = errorTitle
    newErrorModal.showErrorModal = true
    newErrorModal.errorMessage = ''

    if (error.response) {
      const errData = error.response.data
      const errCode = errData.code
      if (errCode === 11) {
        // No Quota
        newErrorModal.errorMessage = error.response.data.message
        newErrorModal.errorHideRetryMessage = true
      } else if (isErrorInsufficientCapacity(error.response.data.message)) {
        newErrorModal.errorDescription = friendlyErrorMessages.insufficientCapacity
        newErrorModal.errorHideRetryMessage = true
      } else {
        newErrorModal.errorMessage = error.response.data.message
      }
    } else {
      newErrorModal.errorMessage = error.message
    }

    setErrorModal(newErrorModal)
  }

  const UpdateNodeGroupCount = async (count) => {
    try {
      await ClusterService.updateNodeGroupNodeCount(count, nodeGroup.clusteruuid, nodeGroup.nodegroupuuid)
      debounceClusterRefresh()
    } catch (error) {
      createErrorModal(error, 'Could not update your cluster')
    }
  }

  const getAction = async (actionConfirm) => {
    if (!actionConfirm || !canUpdateNodeGroup()) {
      setShowConfirmationModal(false)
      return
    }
    switch (confirmationModalAction) {
      case iksNodeGroupActionsEnum.deleteNodeGroup:
        await deleteNodeGroup()
        setShowConfirmationModal(false)
        break
      case iksNodeGroupActionsEnum.upgradeImage:
        await upgradeNodeGroup()
        setShowConfirmationModal(false)
        break
      case iksNodeGroupActionsEnum.addNode: {
        const newCount = nodeGroup.count + 1
        await UpdateNodeGroupCount(newCount)
        setShowConfirmationModal(false)
        break
      }
      case iksNodeGroupActionsEnum.removeNode: {
        const newCount = nodeGroup.count - 1
        await UpdateNodeGroupCount(newCount)
        setShowConfirmationModal(false)
        break
      }
      default:
        break
    }
  }

  const onClickCloseErrorModal = () => {
    const newErrorModal = { ...errorModalInit }
    setErrorModal(newErrorModal)
  }

  function setFilter(event, clear) {
    if (clear) {
      setEmptyGridObject(undefined)
      setFilterText('')
    } else {
      setEmptyGridObject(emptyGridByFilter)
      setFilterText(event.target.value)
    }
  }

  useEffect(() => {
    getNodesGridInfo()
  }, [nodeGroup.nodes])

  return (
    <ClusterNodeGroup
      nodeGroup={nodeGroup}
      columns={columns}
      canUpdateNodeGroup={canUpdateNodeGroup}
      showHowToConnectModal={showHowToConnectModal}
      setShowHowToConnectModal={setShowHowToConnectModal}
      selectedNodeDetails={selectedNodeDetails}
      nodes={nodes}
      getNodegroupStatusMessage={getNodegroupStatusMessage}
      showConfirmationModal={showConfirmationModal}
      openDeleteNodeGroup={openDeleteNodeGroup}
      openUpgradeNodeImage={openUpgradeNodeImage}
      openAddNode={openAddNode}
      openRemoveNode={openRemoveNode}
      getAction={getAction}
      confirmationModalAction={confirmationModalAction}
      errorModal={errorModal}
      onClickCloseErrorModal={onClickCloseErrorModal}
      filterText={filterText}
      setFilter={setFilter}
      emptyGrid={emptyGridObject}
    />
  )
}

export default ClusterNodeGroupContainer
