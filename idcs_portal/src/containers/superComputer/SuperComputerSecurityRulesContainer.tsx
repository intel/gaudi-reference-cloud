// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import useSuperComputerStore from '../../store/superComputer/SuperComputerStore'
import { BsPencilFill, BsTrash3 } from 'react-icons/bs'
import { useNavigate, useParams } from 'react-router'
import SuperComputerSecurityRules from '../../components/superComputer/superComputerSecurityRules/SuperComputerSecurityRules'
import SuperComputerService from '../../services/SuperComputerService'
import useToastStore from '../../store/toastStore/ToastStore'

const SuperComputerSecurityRulesContainer = (): JSX.Element => {
  // *****
  // Global state
  // *****
  const clusterDetail = useSuperComputerStore((state) => state.clusterDetail)
  const setDebounceDetailRefresh = useSuperComputerStore((state) => state.setDebounceDetailRefresh)
  const loading = useSuperComputerStore((state) => state.loading)
  const setEditSecurityRule = useSuperComputerStore((state) => state.setEditSecurityRule)
  const showSuccess = useToastStore((state) => state.showSuccess)
  const showError = useToastStore((state) => state.showError)
  // *****
  // Local state
  // *****
  const columns = [
    {
      columnName: 'Name',
      targetColumn: 'vipname'
    },
    {
      columnName: 'Type',
      targetColumn: 'viptype'
    },
    {
      columnName: 'State',
      targetColumn: 'state'
    },
    {
      columnName: 'Source IPs',
      targetColumn: 'sourceip'
    },
    {
      columnName: 'Protocol',
      targetColumn: 'protocol'
    },
    {
      columnName: 'Target IP',
      targetColumn: 'destinationip'
    },
    {
      columnName: 'Target Port',
      targetColumn: 'port'
    },
    {
      columnName: 'Actions',
      targetColumn: 'actions',
      columnConfig: {
        behaviorType: 'Buttons',
        behaviorFunction: null
      }
    }
  ]

  const actionsOptions = [
    {
      id: 'edit',
      label: 'Edit',
      name: (
        <>
          <BsPencilFill /> Edit
        </>
      ),
      status: ['Active', 'Not Specified', 'Deleted']
    },
    {
      id: 'delete',
      label: 'Delete',
      name: (
        <>
          <BsTrash3 /> Delete
        </>
      ),
      status: ['Active', 'Deleting', 'Reconciling']
    }
  ]

  const deleteModalInitial = {
    show: false,
    label: 'Delete Security Rule',
    question: 'Do you want to delete security rule $<Name> ?',
    feedback: 'If the security rule is deleted. All your rule security information will be lost.',
    buttonLabel: 'Delete',
    name: '',
    vipId: ''
  }

  const emptyGrid = {
    title: 'No security rules found',
    subTitle: 'No security rules are created'
  }

  const [securityRules, setSecurityRules] = useState<any[] | null>(null)
  const [deleteModal, setDeleteModal] = useState(deleteModalInitial)
  const { param: name } = useParams()
  const navigate = useNavigate()

  // *****
  // Hooks
  // *****
  useEffect(() => {
    getSecurityGrid()
  }, [clusterDetail])

  // *****
  // Functions
  // *****
  const getSecurityGrid = (): void => {
    const gridInfo: any[] = []
    for (const index in clusterDetail?.securityRules) {
      const securityRule = { ...clusterDetail?.securityRules[Number(index)] }
      if (securityRule) {
        const state = securityRule.state ?? ''
        gridInfo.push({
          vipname: securityRule.vipname,
          viptype: securityRule.viptype,
          state: securityRule.state,
          sourceip: securityRule.sourceip?.join(' - ') ?? '',
          protocol: securityRule.protocol?.join(' - ') ?? '',
          destinationip: securityRule.destinationip,
          internalport: securityRule.port,
          actions: {
            showField: true,
            type: 'Buttons',
            value: securityRule,
            selectableValues: getActionsByStatus(state, actionsOptions),
            function: setAction
          }
        })
      }
    }
    setSecurityRules(gridInfo)
  }

  const setAction = (action: any, item: any): void => {
    switch (action.id) {
      case 'delete': {
        const question = deleteModal.question.replace('$<Name>', item.vipname)
        const vipId = item.vipid
        setDeleteModal({ ...deleteModal, show: true, question, vipId, name: item.vipname })
        break
      }
      default:
        setEditSecurityRule(item)
        navigate({
          pathname: `/supercomputer/d/${name}/editSecurityRule`
        })
        break
    }
  }

  const getActionsByStatus = (status: string, options: any): any[] => {
    const result = []

    for (const index in options) {
      const option = { ...options[index] }
      if (option.status.find((item: string) => item === status)) {
        result.push(option)
      }
    }

    return result
  }

  const onClickActionModal = async (result: boolean): Promise<void> => {
    if (!result) {
      setDeleteModal(deleteModalInitial)
    } else {
      if (clusterDetail) {
        await deleteRule(clusterDetail?.uuid, deleteModal.vipId)
      }
    }
  }

  const deleteRule = async (clusteruuid: string, vipId: string): Promise<void> => {
    try {
      await SuperComputerService.deleteSecurityRule(clusteruuid, vipId)
      showSuccess('Security Rule updated successfully', false)
      setDeleteModal(deleteModalInitial)
      setDebounceDetailRefresh(true)
    } catch (error: any) {
      setDeleteModal(deleteModalInitial)
      const message = String(error.message)
      if (error.response) {
        const errData = error.response.data
        const errMessage = errData.message
        showError(errMessage, false)
      } else {
        showError(message, false)
      }
    }
  }

  return (
    <SuperComputerSecurityRules
      securityRules={securityRules ?? []}
      columns={columns}
      emptyGrid={emptyGrid}
      loading={loading || securityRules === null}
      deleteModal={deleteModal}
      onClickActionModal={onClickActionModal}
    />
  )
}

export default SuperComputerSecurityRulesContainer
