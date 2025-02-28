// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState, useEffect } from 'react'
import useSuperComputerStore from '../../store/superComputer/SuperComputerStore'
import SuperComputerStorage from '../../components/superComputer/superComputerStorage/SuperComputerStorage'
import { useNavigate } from 'react-router'
import { BsDashCircle, BsCheckCircle, BsStopCircle, BsPlayCircle, BsTrash3 } from 'react-icons/bs'
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

const SuperComputerStorageContainer = () => {
  // *****
  // Global state
  // *****
  const clusterDetail = useSuperComputerStore((state) => state.clusterDetail)

  // *****
  // Hooks
  // *****
  useEffect(() => {
    getStorageInfo()
  }, [clusterDetail])

  // *****
  // local state
  // *****
  const displayInfoInitial = [
    { label: 'Size:', field: 'size', value: '' },
    { label: 'State:', field: 'state', value: '' },
    { label: 'Provider:', field: 'storageprovider', value: '' },
    { label: 'Message:', field: 'message', value: '' }
  ]

  const [storageItems, setStorageItems] = useState([])
  const [isClusterInActiveState, setIsClusterInActiveState] = useState(true)
  const navigate = useNavigate()
  // *****
  // functions
  // *****
  const getStorageInfo = () => {
    setIsClusterInActiveState(clusterDetail.clusterstate !== 'Active')
    const gridInfo = []
    for (const storageIndex in clusterDetail.storages) {
      const storage = { ...clusterDetail.storages[storageIndex] }
      for (const index in displayInfoInitial) {
        const item = { ...displayInfoInitial[index] }
        if (item.formula === 'length') {
          item.value = storage[item.field].length
        } else if (item.field === 'state') {
          item.value = getActionItemLabel(storage[item.field])
        } else {
          item.value = storage[item.field]
        }
        gridInfo.push(item)
      }
      setStorageItems(gridInfo)
    }
  }

  const goToAddStorage = () => {
    navigate(`/supercomputer/d/${clusterDetail.name}/addstorage`)
  }

  return (
    <SuperComputerStorage
      storageItems={storageItems}
      isClusterInActiveState={isClusterInActiveState}
      goToAddStorage={goToAddStorage}
    />
  )
}

export default SuperComputerStorageContainer
