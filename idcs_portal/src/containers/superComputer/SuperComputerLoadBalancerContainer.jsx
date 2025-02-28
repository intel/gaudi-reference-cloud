// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import SuperComputerLoadBalancer from '../../components/superComputer/superComputerLoadBalancer/SuperComputerLoadBalancer'
import useSuperComputerStore from '../../store/superComputer/SuperComputerStore'
import { useNavigate } from 'react-router'
import { BsDashCircle, BsCheckCircle, BsStopCircle, BsPlayCircle, BsTrash3 } from 'react-icons/bs'
import SuperComputerService from '../../services/SuperComputerService'
import StateTooltipCell from '../../utils/gridPagination/cellRender/StateTooltipCell'

const getActionItemLabel = (text, statusStep = null, option = null) => {
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
    case 'Starting':
      message = (
        <>
          {' '}
          <BsPlayCircle /> {text}{' '}
        </>
      )
      break
    case 'Stopping':
      message = (
        <>
          {' '}
          <BsStopCircle /> {text}{' '}
        </>
      )
      break
    case 'Stopped':
      message = (
        <>
          {' '}
          <BsStopCircle /> {text}{' '}
        </>
      )
      break
    case 'Updating':
    case 'Pending':
    case 'Provisioning':
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
    case 'Terminating':
      message = (
        <>
          {' '}
          <BsDashCircle /> {text}{' '}
        </>
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
    case 'Delete':
      message = (
        <>
          {' '}
          <BsTrash3 /> {text}{' '}
        </>
      )
      break
    default:
      message = <> {text} </>
      break
  }

  return message
}

const SuperComputerLoadBalancerContainer = () => {
  // *****
  // Global state
  // *****
  const clusterDetail = useSuperComputerStore((state) => state.clusterDetail)
  const clusterResourceLimit = useSuperComputerStore((state) => state.clusterResourceLimit)
  const setDebounceDetailRefresh = useSuperComputerStore((state) => state.setDebounceDetailRefresh)
  // *****
  // local state
  // *****
  const columns = [
    {
      columnName: 'Load Balancer Name',
      targetColumn: 'name'
    },
    {
      columnName: 'State',
      targetColumn: 'status'
    },
    {
      columnName: 'IP',
      targetColumn: 'ip',
      className: 'text-end'
    },
    {
      columnName: 'Type',
      targetColumn: 'type'
    },
    {
      columnName: 'Port',
      targetColumn: 'port'
    },
    {
      columnName: 'Actions',
      targetColumn: 'actions'
    }
  ]

  const ilbactionsOptions = [
    {
      id: 'deleteilb',
      name: getActionItemLabel('Delete'),
      status: ['Active', 'Error', 'Ready'],
      label: 'Delete load balancer',
      question: 'Do you want to delete load balancer $<Name> ?',
      feedback: 'All your information will be lost.',
      buttonLabel: 'Delete'
    }
  ]

  const modalContentInitial = {
    show: false,
    label: '',
    buttonLabel: '',
    buttonVariant: '',
    name: '',
    vipId: '',
    question: '',
    feedback: ''
  }

  const [myVips, setMyVips] = useState([])
  const navigate = useNavigate()
  const [actionModal, setActionModal] = useState(modalContentInitial)
  const [isClusterInActiveState, setIsClusterInActiveState] = useState(true)
  // *****
  // Hooks
  // *****
  useEffect(() => {
    updateClusterInfo()
  }, [clusterDetail])
  // *****
  // functions
  // *****
  const updateClusterInfo = () => {
    setIsClusterInActiveState(clusterDetail.clusterstate !== 'Active')
    const gridInfo = []
    for (const index in clusterDetail.vips) {
      const vip = { ...clusterDetail.vips[index] }
      gridInfo.push({
        name: vip.name,
        status: {
          showField: true,
          type: 'function',
          value: vip,
          sortValue: vip.vipstate,
          function: getStatusInfo
        },
        ip: vip.vipIp,
        type: vip.viptype,
        port: vip.poolport,
        actions: {
          showField: true,
          type: 'Buttons',
          value: vip,
          selectableValues: getActionsByStatus(vip.vipstate, ilbactionsOptions),
          function: setAction
        }
      })
    }
    setMyVips(gridInfo)
  }
  function getStatusInfo(cluster) {
    return getActionItemLabel(cluster.vipstate)
  }
  function goToAddLoadBalancer() {
    navigate(`/supercomputer/d/${clusterDetail.name}/addloadbalancer`)
  }
  function setAction(action, item) {
    const question = action.question.replace('$<Name>', item.name)
    setActionModal({
      ...actionModal,
      show: true,
      name: item.name,
      vipId: item.vipid,
      question,
      feedback: action.feedback,
      buttonLabel: action.buttonLabel,
      label: action.label
    })
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
  function onActionModal(result) {
    if (!result) {
      setActionModal({ ...actionModal, show: false })
      return
    }
    deleteLb()
  }
  async function deleteLb() {
    try {
      await SuperComputerService.deleteLoadBalancer(clusterDetail.uuid, actionModal.vipId)
      setDebounceDetailRefresh(true)
      setDebounceDetailRefresh(true)
      setActionModal({ ...actionModal, show: false })
    } catch (error) {
      setActionModal({ ...actionModal, show: false })
    }
  }
  return (
    <SuperComputerLoadBalancer
      vipsLimit={clusterResourceLimit?.maxvipspercluster || 2}
      loadBalancerItems={myVips}
      columns={columns}
      isClusterInActiveState={isClusterInActiveState}
      goToAddLoadBalancer={goToAddLoadBalancer}
      actionModal={actionModal}
      onActionModal={onActionModal}
    />
  )
}

export default SuperComputerLoadBalancerContainer
