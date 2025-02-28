// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { BsCheckCircle, BsNodeMinusFill, BsNodePlusFill } from 'react-icons/bs'
import SuperComputerWorkerNodeInfo from '../../components/superComputer/superComputerWorkerNodes/SuperComputerWorkerNodeInfo'
import { useEffect, useState } from 'react'
import StateTooltipCell from '../../utils/gridPagination/cellRender/StateTooltipCell'
import { getErrorMessageFromCodeAndMessage } from '../../utils/apiError/apiError'
import useSuperComputerStore from '../../store/superComputer/SuperComputerStore'

const getActionItemLabel = (text, statusStep = null, option = null) => {
  let message = null
  switch (text) {
    case 'Pending':
    case 'Provisioning':
    case 'Error':
    case 'Updating':
      message = (
        <div className="d-flex flex-row bd-highlight">
          {option && (
            <>
              <div>{'State: '}</div>
              <StateTooltipCell statusStep={statusStep} text={text} spinnerAtTheEnd />
            </>
          )}
          {!option && <StateTooltipCell statusStep={statusStep} text={text} />}
        </div>
      )
      break
    case 'Active':
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

const SuperComputerWorkerNodeInfoContainer = ({
  nodeInfo,
  nodeGroupIndex,
  type,
  setGroupAction,
  isClusterInActiveState
}) => {
  // *****
  // local state
  // *****

  const groupActionsOptions = [
    {
      id: 'addNode',
      name: (
        <>
          <BsNodePlusFill /> Add Node{' '}
        </>
      ),
      status: ['Active'],
      function: setGroupAction
    },
    {
      id: 'deleteNode',
      name: (
        <>
          <BsNodeMinusFill /> Delete Node{' '}
        </>
      ),
      status: ['Active'],
      function: setGroupAction
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
      targetColumn: 'statusStorage'
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

  const emptyGridByFilter = {
    title: 'No nodes found',
    subTitle: 'The applied filter criteria did not match any items.',
    action: {
      type: 'function',
      href: () => setFilter('', true),
      label: 'Clear filters'
    }
  }

  const [nodes, setNodes] = useState([])
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(undefined)

  const loading = useSuperComputerStore((state) => state.loading)

  // Hooks

  useEffect(() => {
    getNodesGridInfo()
  }, [nodeInfo.nodes])

  // Functions

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

  function setAction(action, item) {
    switch (action.id) {
      case 'sshconnect': {
        // TODO
        break
      }
      default:
        break
    }
  }

  function getStatusInfo(node) {
    return getActionItemLabel(node.state, getNodeStatusMessage(node))
  }

  const getNodesGridInfo = () => {
    if (!nodeInfo.nodes || nodeInfo.nodes.length === 0) {
      setNodes([])
      return
    }
    const gridInfo = []

    for (const index in nodeInfo.nodes) {
      const instance = { ...nodeInfo.nodes[index] }
      const wekaStorage = { ...instance.wekaStorage }
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
        statusStorage: wekaStorage.status ? wekaStorage.status : 'No Status',
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

  function setFilter(event, clear) {
    if (clear) {
      setEmptyGridObject(undefined)
      setFilterText('')
    } else {
      setEmptyGridObject(emptyGridByFilter)
      setFilterText(event.target.value)
    }
  }

  return (
    <SuperComputerWorkerNodeInfo
      columns={columns}
      nodes={nodes ?? []}
      loading={loading || nodes === null}
      nodeInfo={nodeInfo}
      nodeGroupIndex={nodeGroupIndex}
      groupActionsOptions={groupActionsOptions}
      type={type}
      setGroupAction={setGroupAction}
      filterText={filterText}
      setFilter={setFilter}
      emptyGrid={emptyGridObject}
      isClusterInActiveState={isClusterInActiveState}
    />
  )
}

export default SuperComputerWorkerNodeInfoContainer
