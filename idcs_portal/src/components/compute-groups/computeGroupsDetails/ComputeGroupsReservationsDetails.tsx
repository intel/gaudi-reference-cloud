// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import GridPagination from '../../../utils/gridPagination/gridPagination'
import TapContent from '../../../utils/TapContent/TapContent'
import { Dropdown, DropdownButton, ButtonGroup } from 'react-bootstrap'
import HowToConnect from '../../../utils/modals/howToConnect/HowToConnect'
import ActionConfirmation from '../../../utils/modals/actionConfirmation/ActionConfirmation'
import idcConfig from '../../../config/configurator'
import CustomAlerts from '../../../utils/customAlerts/CustomAlerts'
import TabsNavigation from '../../../utils/tabsNavigation/TabsNavagation'
import LabelValuePair from '../../../utils/labelValuePair/LabelValuePair'
import { type TapDetail } from '../../../containers/compute-groups/ComputeGroupsReservationsDetailsContainers'
import Spinner from '../../../utils/spinner/Spinner'
import ImageComponents from '../../../utils/imageComponents/ImageComponents'
import MetricsGraphsContainer from '../../../containers/metrics/MetricsGraphsContainer'

interface ComputeGroupsReservationsDetailsProps {
  detailsColumns: any[]
  reserveDetails: any
  selectedNodeDetails: any
  setActiveTap: (value: string | number) => void
  activeTap: string | number
  tapDetails: any[]
  actionsReserveDetails: any[]
  setAction: (action: any, item: any) => void
  showHowToConnectModal: boolean
  setShowHowToConnectModal: (value: boolean) => void
  showActionModal: boolean
  setShowActionModal: (value: boolean) => any
  actionModalContent: any
  taps: any[]
  loading: boolean
  instanceGroupInstances: any[]
  allowedInstanceCategoriesMetrics: any[]
  allowedInstanceStatus: any[]
}

const ComputeGroupsReservationsDetails: React.FC<ComputeGroupsReservationsDetailsProps> = ({
  detailsColumns,
  reserveDetails,
  selectedNodeDetails,
  setActiveTap,
  activeTap,
  tapDetails,
  actionsReserveDetails,
  setAction,
  showHowToConnectModal,
  setShowHowToConnectModal,
  showActionModal,
  setShowActionModal,
  actionModalContent,
  taps,
  loading,
  instanceGroupInstances,
  allowedInstanceCategoriesMetrics,
  allowedInstanceStatus
}): JSX.Element => {
  // variables
  let instanceDetails = null

  function getTapInfo(activeTap: any): TapDetail {
    let content = null
    switch (activeTap) {
      case 0:
        tapDetails[activeTap].customContent = getDetailsTapInfo(tapDetails[activeTap])
        content = tapDetails[activeTap]
        break
      case 1:
        tapDetails[activeTap].customContent = getNetworkTapInfo(tapDetails[activeTap])
        content = tapDetails[activeTap]
        break
      case 2:
        tapDetails[activeTap].customContent = getSecurityTapInfo(tapDetails[activeTap])
        content = tapDetails[activeTap]
        break
      case 3:
        tapDetails[activeTap].customContent = getMetrics()
        content = tapDetails[activeTap]
        break
      default:
        content = tapDetails[activeTap]
        break
    }

    return content
  }

  const getMetrics = (): JSX.Element => {
    const instancesValues = instanceGroupInstances
      .filter(
        (x) =>
          allowedInstanceCategoriesMetrics.includes(x.instanceTypeDetails?.instanceCategory as string) &&
          allowedInstanceStatus.includes(x.status)
      )
      .map((instance) => {
        return {
          name: String(instance.name) + ` (${instance.instanceType})`,
          value: instance.resourceId,
          instanceName: instance.name,
          instanceCategory: instance.instanceTypeDetails?.instanceCategory
        }
      })

    return <MetricsGraphsContainer instances={instancesValues} showInstances={true} />
  }

  function getDetailsTapInfo(tapContent: TapDetail): JSX.Element {
    const columnSize = 3
    const { machineImageInfo, nodesInformation } = tapContent

    let machineDisplayName = null
    let description = null
    let components = []
    let labels = null
    let componentsHtml = <></>

    if (machineImageInfo) {
      machineDisplayName = machineImageInfo.displayName ? machineImageInfo.displayName : null
      description = machineImageInfo.description ? machineImageInfo.description : null
      components = machineImageInfo.components ? machineImageInfo.components : []
      labels = machineImageInfo.labels ? machineImageInfo.labels : null
    }

    componentsHtml = <ImageComponents components={components} />

    return (
      <>
        <h3>Nodes information</h3>
        <GridPagination data={nodesInformation} columns={detailsColumns} emptyGrid={null} loading={loading} />
        <div className="row">
          <h4>Node Information</h4>
          {tapContent.fields.map((item: any, index: number) => (
            <LabelValuePair className={`col-md-${columnSize}`} label={item.label} key={index}>
              {item.value}
            </LabelValuePair>
          ))}
        </div>
        {machineImageInfo ? (
          <>
            <h4 className="mt-s4">Machine image on group ({machineDisplayName})</h4>
            <div className="row">
              <LabelValuePair className="col-md-3" label="Description">
                {description}
              </LabelValuePair>
              <LabelValuePair className="col-md-3" label="Components">
                <div className="d-flex flex-column">{componentsHtml}</div>
              </LabelValuePair>
              <LabelValuePair className="col-md-3" label="Architecture">
                {labels ? labels.architecture : null}
              </LabelValuePair>
              <LabelValuePair className="col-md-3" label="Family">
                {labels ? labels.family : null}
              </LabelValuePair>
            </div>
          </>
        ) : null}
      </>
    )
  }

  function getNetworkTapInfo(tapContent: TapDetail): JSX.Element {
    const { fields, nodeIpsInformation } = tapContent

    const columnsIps = [
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

    const columnsInterfaces = [
      {
        columnName: 'Interface name',
        targetColumn: 'interface'
      },
      {
        columnName: 'vNet name',
        targetColumn: 'vnet'
      }
    ]
    const dataInterfaces = [{ interface: fields[0].value, vnet: fields[1].value }]

    return (
      <>
        <CustomAlerts
          showAlert={true}
          alertType="secondary"
          message={`The network is configured automatically by the ${idcConfig.REACT_APP_CONSOLE_LONG_NAME}`}
          onCloseAlert={() => null}
          showIcon={true}
        />
        <h4>IPs:</h4>
        <div className="row">
          <div className="col-xs-12 col-md-5 col-xl-4">
            <GridPagination
              data={nodeIpsInformation}
              columns={columnsIps}
              emptyGrid={null}
              hidePaginationControl
              hideSortControls
            />
          </div>
        </div>
        <h4>Interfaces for all nodes:</h4>
        <div className="row">
          <div className="col-xs-12 col-md-5 col-xl-4">
            <GridPagination
              data={dataInterfaces}
              columns={columnsInterfaces}
              emptyGrid={null}
              hidePaginationControl
              hideSortControls
            />
          </div>
        </div>
      </>
    )
  }

  function getSecurityTapInfo(tapContent: TapDetail): JSX.Element {
    const fields = tapContent.fields
    const columns = [
      {
        columnName: 'Key name',
        targetColumn: 'key'
      }
    ]
    const data = fields.length > 0 ? fields[0].value.map((x: any) => ({ key: x })) : []
    return (
      <div className="row">
        <div className="col-xs-12 col-md-5 col-xl-4">
          <GridPagination data={data} columns={columns} emptyGrid={null} hidePaginationControl />
        </div>
      </div>
    )
  }

  const spinner = <Spinner />

  if (reserveDetails != null) {
    instanceDetails = (
      <>
        <div className="section flex-row bd-highlight">
          <h2>Instance group: {reserveDetails?.name}</h2>
          <DropdownButton
            as={ButtonGroup}
            variant="outline-primary"
            title="Actions"
            intc-id="myReservationActionsDropdownButton"
          >
            {actionsReserveDetails.map((action, index) => (
              <Dropdown.Item
                key={index}
                onClick={() => {
                  setAction(action, reserveDetails)
                }}
                intc-id={`myReservationActionsDropdownItemButton${index}`}
              >
                {action.name}
              </Dropdown.Item>
            ))}
          </DropdownButton>
        </div>
        <div className="section">
          <TabsNavigation tabs={taps} activeTab={activeTap} setTabActive={setActiveTap} />
          <TapContent infoToDisplay={getTapInfo(activeTap)} />
        </div>
      </>
    )
  }

  return (
    <>
      <ActionConfirmation
        actionModalContent={actionModalContent}
        onClickModalConfirmation={setShowActionModal}
        showModalActionConfirmation={showActionModal}
      />
      <HowToConnect
        flowType="compute"
        data={selectedNodeDetails}
        showHowToConnectModal={showHowToConnectModal}
        onClickHowToConnect={setShowHowToConnectModal}
      />
      {loading || !reserveDetails ? spinner : instanceDetails}
    </>
  )
}

export default ComputeGroupsReservationsDetails
