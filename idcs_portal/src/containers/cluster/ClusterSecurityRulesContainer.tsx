// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import useClusterStore from '../../store/clusterStore/ClusterStore'
import { BsPencilFill, BsTrash3 } from 'react-icons/bs'
import ClusterSecurityRules from '../../components/cluster/clusterMyReservations/ClusterSecurityRules'
import { useNavigate, useParams } from 'react-router'
import useToastStore from '../../store/toastStore/ToastStore'
import ClusterService from '../../services/ClusterService'

const ClusterSecurityRulesContainer = (): JSX.Element => {
  // *****
  // Global state
  // *****
  const clusterSecurityRules = useClusterStore((state) => state.clusterSecurityRules)
  const securityRuleUuid = useClusterStore((state) => state.securityRuleUuid)
  const setEditSecurityRule = useClusterStore((state) => state.setEditSecurityRule)
  const setShouldRefreshSecurityRules = useClusterStore((state) => state.setShouldRefreshSecurityRules)
  const loading = useClusterStore((state) => state.loading)
  const navigate = useNavigate()
  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)
  const { param: name } = useParams()

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
  // *****
  // use effect
  // *****

  useEffect(() => {
    getSecurityGrid()
    setShouldRefreshSecurityRules(true)
    return () => {
      setShouldRefreshSecurityRules(false)
    }
  }, [clusterSecurityRules])

  // *****
  // Functions
  // *****
  const getSecurityGrid = (): void => {
    const gridInfo: any[] = []
    for (const index in clusterSecurityRules) {
      const securityRule = { ...clusterSecurityRules[Number(index)] }
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
          selectableValues: getActionsByStatus(securityRule.state, actionsOptions),
          function: setAction
        }
      })
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
          pathname: `/cluster/d/${name}/editSecurityRule`
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
      if (securityRuleUuid) {
        await deleteRule(securityRuleUuid, deleteModal.vipId)
      }
    }
  }

  const deleteRule = async (clusteruuid: string, vipId: string): Promise<void> => {
    try {
      await ClusterService.deleteSecurityRule(clusteruuid, vipId)
      showSuccess('Security Rule updated successfully', false)
      setDeleteModal(deleteModalInitial)
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
    <ClusterSecurityRules
      securityRules={securityRules ?? []}
      columns={columns}
      loading={loading || securityRules === null}
      emptyGrid={emptyGrid}
      deleteModal={deleteModal}
      onClickActionModal={onClickActionModal}
    />
  )
}

export default ClusterSecurityRulesContainer
