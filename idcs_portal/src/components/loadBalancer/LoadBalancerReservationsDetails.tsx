// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import GridPagination from '../../utils/gridPagination/gridPagination'
import TapContent from '../../utils/TapContent/TapContent'
import { Dropdown, DropdownButton, ButtonGroup } from 'react-bootstrap'
import ActionConfirmation from '../../utils/modals/actionConfirmation/ActionConfirmation'
import TabsNavigation from '../../utils/tabsNavigation/TabsNavagation'
import LabelValuePair from '../../utils/labelValuePair/LabelValuePair'
import Spinner from '../../utils/spinner/Spinner'

const LoadBalancerReservationsDetails = (props: any): JSX.Element => {
  // props

  const reserveDetails = props.reserveDetails
  const loadBalancerActiveTab = props.loadBalancerActiveTab
  const setLoadBalancerActiveTab = props.setLoadBalancerActiveTab
  const tabDetails = props.tabDetails
  const actionsReserveDetails = props.actionsReserveDetails
  const showActionModal = props.showActionModal
  const actionModalContent = props.actionModalContent
  const tabs = props.tabs
  const setShowActionModal = props.setShowActionModal
  const setAction = props.setAction
  const listenerColumns = props.listenerColumns
  const reserveListener = props.reserveListener
  const computeInstances = props.computeInstances
  const loading = props.loading

  const getTabInfo = (loadBalancerActiveTab: number): any => {
    let content = null
    switch (loadBalancerActiveTab) {
      case 2:
        tabDetails[loadBalancerActiveTab].customContent = getListenersTabInfo()
        content = tabDetails[loadBalancerActiveTab]
        break
      case 1:
        tabDetails[loadBalancerActiveTab].customContent = getSecurityIpsTabInfo()
        content = tabDetails[loadBalancerActiveTab]
        break
      default:
        content = tabDetails[loadBalancerActiveTab]
        break
    }

    return content
  }

  const getSecurityIpsTabInfo = (): JSX.Element => {
    const content = (
      <>
        <div className="row">
          <div className="col-12">
            <div className="section px-0">
              <ol className="ps-3">{getSecurityIps()}</ol>
            </div>
          </div>
        </div>
      </>
    )

    return content
  }

  const getSecurityIps = (): any[] => {
    const returnSecurityIps: any[] = []
    const sourceIps = reserveDetails?.sourceips ?? []
    for (const ip of sourceIps) {
      returnSecurityIps.push(<li key={ip}>{ip}</li>)
    }
    return returnSecurityIps
  }

  const getListenersTabInfo = (): JSX.Element => {
    const content = (
      <>
        <div className="row">
          <div className="col-12">{getListeners()}</div>
        </div>
      </>
    )

    return content
  }

  const getListeners = (): any[] => {
    const getField = (listener: any, column: string): JSX.Element => {
      return (
        <LabelValuePair label={listenerColumns[column]}>
          <span>{listener[column]}</span>
        </LabelValuePair>
      )
    }

    const getInstanceInfo = (instanceId: string): any => {
      return computeInstances.filter((x: any) => x.resourceId === instanceId)
    }

    const getInstances = (listener: any): any[] => {
      const returnInstances = []
      for (const inst of listener.instanceSelectors) {
        const instance = getInstanceInfo(inst)
        returnInstances.push(<span key={inst}>{instance.length > 0 ? instance[0].name : inst}</span>)
      }
      return returnInstances
    }

    const getTags = (listener: any): JSX.Element => {
      const tagColumns = [
        {
          columnName: 'Tag Name',
          targetColumn: 'tagKey'
        },
        {
          columnName: 'Tag Value',
          targetColumn: 'tagValue'
        }
      ]

      const tags: any[] = []
      for (const i in listener.instanceSelectors) {
        tags.push({
          tagKey: i,
          tagValue: listener.instanceSelectors[i]
        })
      }
      return (
        <GridPagination data={tags} columns={tagColumns} tableClassName="table-bordered" hidePaginationControl={true} />
      )
    }

    const getPoolMembers = (listener: any): JSX.Element => {
      const poolColumns = [
        {
          columnName: 'Instance',
          targetColumn: 'instance'
        },
        {
          columnName: 'IP',
          targetColumn: 'ip',
          className: 'text-end'
        }
      ]

      const data: any[] = []

      for (const member of listener.poolMembers) {
        const instance = getInstanceInfo(member.instanceRef)
        data.push({
          instance: instance.length > 0 ? instance[0].name : member.instanceRef,
          ip: member.ip
        })
      }

      return (
        <GridPagination
          data={data}
          columns={poolColumns}
          tableClassName="table-bordered"
          hidePaginationControl={true}
        />
      )
    }

    const returnListener: any[] = []
    for (const listener of reserveListener) {
      const div = (
        <div className="section px-0" key={listener.externalPort}>
          <h4>Listener</h4>
          <div className="row">
            <div className="col-md-3"> {getField(listener, 'externalPort')} </div>
            <div className="col-md-3">{getField(listener, 'internalPort')}</div>
            <div className="col-md-3">{getField(listener, 'monitor')}</div>
            <div className="col-md-3">{getField(listener, 'loadBalancingMode')}</div>
          </div>
          <div className="row">
            <div className="col-md-6"> {getField(listener, 'message')} </div>
          </div>
          <h5>
            Selector:&nbsp;
            {listener.instanceSelector === 'labels' ? listenerColumns.instanceLabels : listenerColumns.instances}
          </h5>

          <div className="d-flex flex-column gap-s4">
            {listener.instanceSelector === 'labels' ? getTags(listener) : getInstances(listener)}
          </div>

          {listener.poolMembers.length > 0 && (
            <>
              <h5>Pool Members</h5>
              <div className="d-flex flex-column gap-s4">{getPoolMembers(listener)}</div>
            </>
          )}

          <hr></hr>
        </div>
      )
      returnListener.push(div)
    }
    return returnListener
  }

  const spinner: JSX.Element = <Spinner />

  const instanceDetails: JSX.Element = (
    <>
      <div className="section flex-row bd-highlight">
        <h2>Load Balancer: {reserveDetails?.name}</h2>

        {actionsReserveDetails.length > 0 && (
          <DropdownButton
            as={ButtonGroup}
            variant="outline-primary"
            title="Actions"
            intc-id="loadBalancerReservationActionsDropdownButton"
          >
            {actionsReserveDetails.map((action: any, index: number) => (
              <Dropdown.Item
                key={index}
                onClick={() => setAction(action, reserveDetails)}
                intc-id={`loadBalancerReservationActionsDropdownButton${index}`}
              >
                {action.name}
              </Dropdown.Item>
            ))}
          </DropdownButton>
        )}
      </div>
      <div className="section">
        <TabsNavigation tabs={tabs} activeTab={loadBalancerActiveTab} setTabActive={setLoadBalancerActiveTab} />
        <TapContent infoToDisplay={getTabInfo(loadBalancerActiveTab)} />
      </div>
    </>
  )

  return (
    <>
      <ActionConfirmation
        actionModalContent={actionModalContent}
        onClickModalConfirmation={setShowActionModal}
        showModalActionConfirmation={showActionModal}
      />
      {loading || !reserveDetails ? spinner : instanceDetails}
    </>
  )
}

export default LoadBalancerReservationsDetails
