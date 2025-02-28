// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import React, { useEffect, useState } from 'react'
import useSuperComputerStore from '../../store/superComputer/SuperComputerStore'
import SuperComputerWorkerNodes from '../../components/superComputer/superComputerWorkerNodes/SuperComputerWorkerNodes'
import { BsNodePlusFill } from 'react-icons/bs'
import { useNavigate } from 'react-router'
import SuperComputerService from '../../services/SuperComputerService'
import { friendlyErrorMessages, isErrorInsufficientCapacity } from '../../utils/apiError/apiError'
import { superComputerNodeGroupTypes } from '../../utils/Enums'

const SuperComputerWorkerNodesContainer = () => {
  // *****
  // Global state
  // *****
  const clusterDetail = useSuperComputerStore((state) => state.clusterDetail)
  const setDebounceDetailRefresh = useSuperComputerStore((state) => state.setDebounceDetailRefresh)
  const nodeTabNumber = useSuperComputerStore((state) => state.nodeTabNumber)
  // *****
  // local state
  // *****

  const tabsInitial = [
    {
      id: 'ai',
      label: 'AI nodes',
      actions: [],
      show: true,
      content: []
    },
    {
      id: 'compute',
      label: 'Compute nodes',
      actions: [
        {
          name: (
            <>
              <BsNodePlusFill /> Add node group{' '}
            </>
          ),
          function: addNodeGroup
        }
      ],
      show: true,
      content: []
    }
  ]

  const nodeGroupInitial = {
    groupName: '',
    clusterUuid: '',
    nodegroupUuid: '',
    instanceType: '',
    userDataUrl: '',
    nodeGroupState: '',
    nodeStatus: '',
    upgradeAvailable: false,
    nodes: []
  }

  const modalContentInitial = {
    show: false,
    action: '',
    label: '',
    buttonLabel: '',
    buttonVariant: '',
    name: '',
    clusterUuid: '',
    nodeGroupUuid: '',
    question: '',
    feedback: '',
    nodeCount: ''
  }

  const modalContentOptions = [
    {
      id: 'deleteGroup',
      title: 'Delete node group',
      question: 'Are you sure you want to delete node group $<Name> ?',
      feedback: 'If the node group is running it will be stopped. All your information will be lost.',
      label: 'Delete node group',
      buttonLabel: 'Delete'
    },
    {
      id: 'deleteNode',
      title: 'Delete node',
      question: 'The node count of node group $<Name> will change from $<NodeLength> to $<NewNodeLength>. Continue?',
      feedback:
        'The deleted node will be selected automattically. If running, the node will stop and all your information will be lost.',
      label: 'Delete node',
      buttonLabel: 'Continue'
    },
    {
      id: 'addNode',
      title: 'Add node',
      question: 'The node count of node group $<Name> will change from $<NodeLength> to $<NewNodeLength>. Continue?',
      feedback: '',
      label: 'Add node',
      buttonLabel: 'Add',
      buttonVariant: 'primary'
    },
    {
      id: 'upgradeImage',
      title: 'Upgrade IMI for Nodegroup',
      question: 'Are you sure you want to update the image of node group $<Name>?',
      feedback: '',
      label: 'Upgrade IMI for Nodegroup',
      buttonLabel: 'Update',
      buttonVariant: 'primary'
    }
  ]

  const errorModalInit = {
    show: false,
    errorDescription: null,
    errorMessage: null,
    errorHideRetryMessage: true
  }

  const nodeGroupSelectionInitial = {
    label: 'Nodegroup type:',
    type: 'radio',
    options: [
      {
        name: 'AI nodes',
        value: 0,
        onChanged: () => {
          setActiveTab(0)
        }
      },
      {
        name: 'Compute nodes',
        value: 1,
        onChanged: () => {
          setActiveTab(1)
        }
      }
    ]
  }

  const navigate = useNavigate()
  const [activeTab, setActiveTab] = useState(nodeTabNumber)
  const [isClusterInActiveState, setIsClusterInActiveState] = useState(true)
  const [aiWorkerGroupItems, setAiWorkerGroupItems] = useState([])
  const [computeWorkerGroupItems, setComputeWorkerGroupItems] = useState([])
  const [actionModal, setActionModal] = useState(modalContentInitial)
  const [errorModal, setErrorModal] = useState(errorModalInit)

  // *****
  // Hooks
  // *****
  useEffect(() => {
    getWorkGroupInfo()
  }, [clusterDetail])

  // *****
  // functions
  // *****

  const getWorkGroupInfo = () => {
    setIsClusterInActiveState(clusterDetail.clusterstate === 'Active')
    const nodeGroups = clusterDetail.nodegroups
    if (nodeGroups.length > 0) {
      // AI
      const aiWorkerGroups = nodeGroups.filter(
        (nodeGroup) => nodeGroup.nodeGroupType === superComputerNodeGroupTypes.aiCompute
      )
      const aiNodeGroups = getNodeGroupsItems(aiWorkerGroups)
      const computeWorkerGroups = nodeGroups.filter(
        (nodeGroup) => nodeGroup.nodeGroupType === superComputerNodeGroupTypes.gpCompute
      )
      const computeNodeGroups = getNodeGroupsItems(computeWorkerGroups)
      setComputeWorkerGroupItems(computeNodeGroups)
      setAiWorkerGroupItems(aiNodeGroups)
    }
  }

  const getNodeGroupsItems = (nodeGroupsItems) => {
    const nodeGroups = []
    for (const index in nodeGroupsItems) {
      const workerGroupItem = { ...nodeGroupsItems[index] }
      const nodegroupUuid = workerGroupItem.nodegroupuuid
      const nodeGroup = { ...nodeGroupInitial }
      nodeGroup.nodegroupUuid = nodegroupUuid
      nodeGroup.groupName = workerGroupItem.name
      nodeGroup.clusterUuid = workerGroupItem.clusteruuid
      nodeGroup.userDataUrl = workerGroupItem.userdataurl
      nodeGroup.upgradeAvailable = workerGroupItem.upgradeavailable
      nodeGroup.nodeGroupState = workerGroupItem.nodegroupstate
      nodeGroup.instanceType = workerGroupItem.instanceTypeDetails
        ? workerGroupItem.instanceTypeDetails.displayName
        : ''
      nodeGroup.nodeStatus = `${workerGroupItem.nodegroupstate} - ${
        workerGroupItem.nodegroupstatus.message || 'No Status'
      }`
      const nodeItems = [...workerGroupItem.nodes]
      if (nodeItems.length > 0) {
        const nodes = []
        for (const nodeIndex in nodeItems) {
          const node = { ...nodeItems[nodeIndex] }
          nodes.push({
            ...node
          })
        }
        nodeGroup.nodes = nodes
      }
      nodeGroups.push(nodeGroup)
    }
    return nodeGroups
  }

  function setGroupAction(action, index) {
    const modalContent = modalContentOptions.find((item) => item.id === action.id)
    if (modalContent) {
      const actionModalUpdated = { ...actionModal }
      const groupNode = computeWorkerGroupItems[index]
      actionModalUpdated.show = true
      actionModalUpdated.action = action.id
      actionModalUpdated.clusterUuid = groupNode.clusterUuid
      actionModalUpdated.nodeGroupUuid = groupNode.nodegroupUuid
      actionModalUpdated.feedback = modalContent.feedback
      actionModalUpdated.title = modalContent.title
      actionModalUpdated.name = groupNode.groupName
      actionModalUpdated.label = modalContent.label
      actionModalUpdated.buttonLabel = modalContent.buttonLabel
      actionModalUpdated.buttonVariant = modalContent.buttonVariant
      switch (action.id) {
        case 'addNode': {
          const newCount = groupNode.nodes.length + 1
          let question = modalContent.question.replace('$<Name>', groupNode.groupName)
          question = question.replace('$<NodeLength>', groupNode.nodes.length)
          question = question.replace('$<NewNodeLength>', newCount)
          actionModalUpdated.nodeCount = newCount
          actionModalUpdated.question = question
          break
        }
        case 'deleteNode': {
          const newCount = groupNode.nodes.length - 1
          let question = modalContent.question.replace('$<Name>', groupNode.groupName)
          question = question.replace('$<NodeLength>', groupNode.nodes.length)
          question = question.replace('$<NewNodeLength>', newCount)
          actionModalUpdated.nodeCount = newCount
          actionModalUpdated.question = question
          break
        }
        case 'upgradeImage':
        case 'deleteGroup': {
          const question = modalContent.question.replace('$<Name>', groupNode.groupName)
          actionModalUpdated.question = question
          break
        }
        default:
          break
      }
      setActionModal(actionModalUpdated)
    }
  }

  function addNodeGroup() {
    navigate(`/supercomputer/d/${clusterDetail.name}/addnodegroup`)
  }

  function onActionModal(result) {
    if (!result) {
      setActionModal({ ...actionModal, show: false })
      return
    }
    switch (actionModal.action) {
      case 'deleteNode':
      case 'addNode':
        updateNodetoNodeGroup(actionModal.clusterUuid, actionModal.nodeGroupUuid, actionModal.nodeCount)
        break
      case 'deleteGroup':
        deleteNodeGroup(actionModal.clusterUuid, actionModal.nodeGroupUuid)
        break
      case 'upgradeImage':
        upgradeNodeGroup(actionModal.clusterUuid, actionModal.nodeGroupUuid)
        break
      default:
        break
    }
  }

  async function updateNodetoNodeGroup(clusterUuid, nodeGroupUuid, count) {
    try {
      await SuperComputerService.updateNodeGroupNodeCount(count, clusterUuid, nodeGroupUuid)
      setDebounceDetailRefresh(true)
      setActionModal({ ...actionModal, show: false })
    } catch (error) {
      setActionModal({ ...actionModal, show: false })
      createErrorModal(error, 'Could not update your cluster')
    }
  }

  const deleteNodeGroup = async (clusterUuid, nodeGroupUuid) => {
    try {
      await SuperComputerService.deleteNodeGroup(clusterUuid, nodeGroupUuid)
      setActionModal({ ...actionModal, show: false })
      setDebounceDetailRefresh(true)
    } catch (error) {
      setActionModal({ ...actionModal, show: false })
      createErrorModal(error, 'Could not delete your node group')
    }
  }

  const upgradeNodeGroup = async (clusterUuid, nodeGroupUuid) => {
    try {
      await SuperComputerService.upgradeNodeGroup(clusterUuid, nodeGroupUuid)
      setActionModal({ ...actionModal, show: false })
      setDebounceDetailRefresh(true)
    } catch (error) {
      setActionModal({ ...actionModal, show: false })
      createErrorModal(error, 'Could not upgrade your node group')
    }
  }

  const createErrorModal = (error, errorTitle) => {
    const newErrorModal = { ...errorModalInit }
    newErrorModal.errorHideRetryMessage = false
    newErrorModal.errorDescription = ''
    newErrorModal.titleMessage = errorTitle
    newErrorModal.show = true
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

  const onClickCloseErrorModal = () => {
    setErrorModal(errorModalInit)
  }

  return (
    <SuperComputerWorkerNodes
      aiWorkerGroupItems={aiWorkerGroupItems}
      computeWorkerGroupItems={computeWorkerGroupItems}
      setActiveTab={setActiveTab}
      activeTab={activeTab}
      tabs={tabsInitial}
      addNodeGroup={addNodeGroup}
      actionModal={actionModal}
      setGroupAction={setGroupAction}
      onActionModal={onActionModal}
      isClusterInActiveState={isClusterInActiveState}
      errorModal={errorModal}
      onClickCloseErrorModal={onClickCloseErrorModal}
      nodeGroupSelection={nodeGroupSelectionInitial}
    />
  )
}

export default SuperComputerWorkerNodesContainer
