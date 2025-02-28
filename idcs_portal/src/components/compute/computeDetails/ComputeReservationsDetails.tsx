// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Button, Dropdown, DropdownButton, ButtonGroup } from 'react-bootstrap'
import { BsTerminal } from 'react-icons/bs'
import TapContent from '../../../utils/TapContent/TapContent'
import TabsNavigation from '../../../utils/tabsNavigation/TabsNavagation'
import GridPagination from '../../../utils/gridPagination/gridPagination'
import CustomAlerts from '../../../utils/customAlerts/CustomAlerts'
import LabelValuePair from '../../../utils/labelValuePair/LabelValuePair'
import idcConfig, { appFeatureFlags, isFeatureFlagEnable } from '../../../config/configurator'
import HowToConnect from '../../../utils/modals/howToConnect/HowToConnect'
import ActionConfirmation from '../../../utils/modals/actionConfirmation/ActionConfirmation'
import MetricsGraphsContainer from '../../../containers/metrics/MetricsGraphsContainer'
import Spinner from '../../../utils/spinner/Spinner'
import ImageComponents from '../../../utils/imageComponents/ImageComponents'

const HIDE_HOW_TO_CONNECT_STATUSES = ['Waitlisted', 'Rejected', 'Maintenance']

interface ComputeReservationsDetailsProps {
  reserveDetails: any
  activeTap: number | string
  loading: boolean
  taps: any[]
  tapDetails: any
  actionsReserveDetails: any[]
  actionModalContent: any
  showHowToConnectModal: boolean
  showActionModal: boolean
  setActiveTap: (value: string | number) => void
  setShowHowToConnectModal: (value: boolean) => void
  setAction: (action: any, item: any) => void
  setShowActionModal: (result: boolean, action: string | null) => Promise<void>
  openLinkWithNewTab?: (link: string) => void
  openCloudConnectlink?: (link: string) => void
}

const ComputeReservationsDetails: React.FC<ComputeReservationsDetailsProps> = (props): JSX.Element => {
  const reserveDetails = props.reserveDetails
  const activeTap = props.activeTap
  const tapDetails = props.tapDetails
  const actionsReserveDetails = props.actionsReserveDetails
  const taps = props.taps
  const loading = props.loading
  const showHowToConnectModal = props.showHowToConnectModal
  const showActionModal = props.showActionModal
  const actionModalContent = props.actionModalContent
  const setActiveTap = props.setActiveTap
  const setShowHowToConnectModal = props.setShowHowToConnectModal
  const setAction = props.setAction
  const setShowActionModal: any = props.setShowActionModal
  const openLinkWithNewTab: any = props.openLinkWithNewTab
  const openCloudConnectlink: any = props.openCloudConnectlink

  function getTapInfo(activeTap: number | string): any {
    const tabLabel = taps[Number(activeTap)]?.label
    let content = null
    switch (tabLabel) {
      case 'Details':
        tapDetails[activeTap].customContent = getDetailsTapInfo(tapDetails[activeTap])
        content = tapDetails[activeTap]
        break
      case 'Networking':
        tapDetails[activeTap].customContent = getNetworkTapInfo(tapDetails[activeTap])
        content = tapDetails[activeTap]
        break
      case 'Security':
        tapDetails[activeTap].customContent = getSecurityTapInfo(tapDetails[activeTap])
        content = tapDetails[activeTap]
        break
      case 'Tags':
        tapDetails[activeTap].customContent = getLabelsTapInfo(tapDetails[activeTap])
        content = tapDetails[activeTap]
        break
      case 'Cloud Monitor':
        tapDetails[activeTap].customContent = getMetrics()
        content = tapDetails[activeTap]
        break
      default:
        content = tapDetails[activeTap]
        break
    }

    return content
  }

  function getDetailsTapInfo(tapContent: any): JSX.Element {
    const columnSize = 3
    const machineImageInfo = { ...tapContent.machineImageInfo }
    let machineDisplayName = null
    let description = null
    let components = null
    let labels = null
    let componentsHtml = <></>
    if (machineImageInfo.displayName) {
      machineDisplayName = machineImageInfo.displayName ? machineImageInfo.displayName : null
      description = machineImageInfo.description ? machineImageInfo.description : null
      components = machineImageInfo.components ? machineImageInfo.components : []
      labels = machineImageInfo.labels ? machineImageInfo.labels : null
      componentsHtml = <ImageComponents components={components} />
    }
    return (
      <>
        <div className="row">
          {tapContent.fields.map((item: any, index: number) => (
            <LabelValuePair className={`col-md-${columnSize}`} label={item.label} key={index}>
              {item.value}
            </LabelValuePair>
          ))}
        </div>
        {machineImageInfo.displayName ? (
          <div className="row mt-s4">
            <h3>Machine image information ({machineDisplayName})</h3>
            <LabelValuePair className="col-md-3" label="Description">
              {description}
            </LabelValuePair>
            <LabelValuePair className="col-md-3" label="Components">
              {componentsHtml}
            </LabelValuePair>
            <LabelValuePair className="col-md-3" label="Architecture">
              {labels ? labels.architecture : null}
            </LabelValuePair>
            <LabelValuePair className="col-md-3" label="Family">
              {labels ? labels.family : null}
            </LabelValuePair>
          </div>
        ) : null}
      </>
    )
  }

  function getNetworkTapInfo(tapContent: any): JSX.Element {
    const fields = tapContent.fields
    const columns = [
      {
        columnName: 'Interface name',
        targetColumn: 'interface'
      },
      {
        columnName: 'vNet name',
        targetColumn: 'vnet'
      }
    ]
    const data = [{ interface: fields[0].value, vnet: fields[1].value }]

    return (
      <>
        <CustomAlerts
          showAlert={true}
          alertType="secondary"
          message={`The network is configured automatically by the ${idcConfig.REACT_APP_CONSOLE_LONG_NAME}`}
          onCloseAlert={() => {}}
          showIcon={true}
        />
        <div className="d-flex flex-column">
          <span className="fw-semibold">IP</span>
          <span>{reserveDetails?.ipNbr}</span>
        </div>
        <div className="row">
          <div className="col-xs-12 col-md-5 col-xl-4">
            <GridPagination data={data} columns={columns} emptyGrid={null} hidePaginationControl hideSortControls />
          </div>
        </div>
      </>
    )
  }

  function getSecurityTapInfo(tapContent: any): JSX.Element {
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
          <GridPagination data={data} columns={columns} emptyGrid={null} hidePaginationControl hideSortControls />
        </div>
      </div>
    )
  }

  const getLabelsTapInfo = (tapContent: any): JSX.Element => {
    const labels = tapContent.fields[0].value
    const columns = [
      {
        columnName: 'Key',
        targetColumn: 'key'
      },
      {
        columnName: 'Value',
        targetColumn: 'value'
      }
    ]

    const tags: any[] = []

    for (const i in labels) {
      tags.push({
        tagKey: i,
        tagValue: labels[i]
      })
    }

    return (
      <div className="row">
        <div className="col-xs-12 col-md-5 col-xl-4">
          <GridPagination data={tags} columns={columns} emptyGrid={null} hidePaginationControl hideSortControls />
        </div>
      </div>
    )
  }

  const getMetrics = (): JSX.Element => {
    const instance = {
      name: (reserveDetails.name as string) + ` (${reserveDetails.instanceType})`,
      value: reserveDetails.resourceId,
      instanceName: reserveDetails.name,
      instanceCategory: reserveDetails.instanceCategory
    }
    return <MetricsGraphsContainer instances={[instance]} showInstances={false} />
  }

  const spinner: JSX.Element = <Spinner />

  const instanceDetails: JSX.Element = (
    <>
      <div className="section flex-row bd-highlight">
        <h2>Instance: {reserveDetails?.name}</h2>
        {reserveDetails?.status === 'Ready' && reserveDetails?.sshUrl && (
          <Button
            variant="primary"
            onClick={() => {
              openLinkWithNewTab(reserveDetails?.sshUrl)
            }}
            intc-id="btn-open-cloud-terminal"
            data-wap_ref="btn-computereservation-open-cloud-terminal"
          >
            <BsTerminal />
            Connect
          </Button>
        )}
        {reserveDetails?.status === 'Ready' &&
        isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_QUICK_CONNECT) &&
        reserveDetails?.quickConnectEnabled ? (
          <Button
            variant="primary"
            onClick={() => {
              void openCloudConnectlink()
            }}
            intc-id="btn-open-cloud-terminal"
            data-wap_ref="btn-computereservation-open-cloud-terminal"
          >
            <BsTerminal />
            Connect
          </Button>
        ) : null}
        {!HIDE_HOW_TO_CONNECT_STATUSES.includes(reserveDetails?.status) && reserveDetails?.sshPublicKey[0] !== '' && (
          <Button
            intc-id="btn-computereservation-how-to-connect"
            data-wap_ref="btn-computereservation-how-to-connect"
            variant="outline-primary"
            onClick={() => {
              setShowHowToConnectModal(true)
            }}
          >
            How to Connect via SSH
          </Button>
        )}
        {actionsReserveDetails.length > 0 ? (
          <DropdownButton
            as={ButtonGroup}
            variant="outline-primary"
            title="Actions"
            intc-id="myReservationActionsDropdownButton"
          >
            {actionsReserveDetails.map((action: any, index: number) => (
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
        ) : null}
      </div>
      <div className="section">
        <TabsNavigation tabs={taps} activeTab={activeTap} setTabActive={setActiveTap} />
        <TapContent infoToDisplay={getTapInfo(activeTap)} />
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
      <HowToConnect
        flowType="compute"
        data={reserveDetails}
        showHowToConnectModal={showHowToConnectModal}
        onClickHowToConnect={setShowHowToConnectModal}
      />
      {loading || !reserveDetails ? spinner : instanceDetails}
    </>
  )
}
export default ComputeReservationsDetails
