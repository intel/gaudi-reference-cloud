// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import GridPagination from '../../../utils/gridPagination/gridPagination'
import TapContent from '../../../utils/TapContent/TapContent'
import { Button, Dropdown, DropdownButton, ButtonGroup } from 'react-bootstrap'
import ActionConfirmation from '../../../utils/modals/actionConfirmation/ActionConfirmation'
import { BsNodePlus, BsFillPlugFill, BsQuestionCircle } from 'react-icons/bs'
import UpgradeVersionModal from './UpgradeVersionModal'
import ClusterNodeGroupContainer from '../../../containers/cluster/ClusterNodeGroupContainer'
import idcConfig from '../../../config/configurator'
import ClusterStorage from './ClusterStorage'
import TabsNavigation from '../../../utils/tabsNavigation/TabsNavagation'
import LabelValuePair from '../../../utils/labelValuePair/LabelValuePair'
import Spinner from '../../../utils/spinner/Spinner'
import { ReactComponent as ExternalLink } from '../../../assets/images/ExternalLink.svg'
import ClusterGraphsContainer from '../../../containers/metrics/ClusterGraphsContainer'

interface ClusterMyReservationsDetailProps {
  lbcolumns: any[]
  myloadbalancers: any[]
  reserveDetails: any
  actionsReserveDetails: any[]
  tabs: any
  tabDetails: any
  activeTab: number
  clusterResourceLimit: any
  nodegroupsInfo: any[]
  showActionModal: boolean
  actionModalContent: any
  loading: boolean
  upgradeModal: any
  upgradeForm: any
  myStorages: any[]
  emptyStorage: any
  setAction: (action: any, item: any) => void
  setActiveTab: (value: number) => void
  actionOnModal: (value: boolean) => Promise<void>
  getActionItemLabel: (text: string, statusStep?: string | null, option?: any) => JSX.Element
  goToAddNodeGroup: () => void
  goToAddLoadBalancer: () => void
  onChangeUpgradeForm: (event: any, inputName: string) => void
  submitUpgradeK8sVersion: () => Promise<void>
  debounceClusterRefresh: () => void
}

const ClusterMyReservationsDetail: React.FC<ClusterMyReservationsDetailProps> = (props) => {
  // *****
  // props
  // *****

  const lbcolumns = props.lbcolumns
  const myloadbalancers = props.myloadbalancers
  const reserveDetails = props.reserveDetails
  const actionsReserveDetails = props.actionsReserveDetails
  const tabs = props.tabs ? props.tabs.filter((x: any) => x.visible) : []
  const tabDetails = props.tabDetails
  const activeTab = props.activeTab
  const nodegroupsInfo = props.nodegroupsInfo
  const showActionModal = props.showActionModal
  const actionModalContent = props.actionModalContent
  const loading = props.loading
  const upgradeModal = props.upgradeModal
  const upgradeForm = props.upgradeForm
  const isValidUpgradeForm = upgradeForm.isValidForm
  const myStorages = props.myStorages
  const emptyStorage = props.emptyStorage
  const clusterResourceLimit = props.clusterResourceLimit
  const setAction = props.setAction
  const setActiveTab: any = props.setActiveTab
  const actionOnModal: any = props.actionOnModal
  const getActionItemLabel = props.getActionItemLabel
  const goToAddNodeGroup = props.goToAddNodeGroup
  const goToAddLoadBalancer = props.goToAddLoadBalancer
  const onChangeUpgradeForm = props.onChangeUpgradeForm
  const submitUpgradeK8sVersion = props.submitUpgradeK8sVersion
  const debounceClusterRefresh = props.debounceClusterRefresh
  // *****
  // variables
  // *****
  const spinner = <Spinner />

  const clusterDetails = (
    <>
      <div className="section flex-row bd-highlight">
        <h2>Cluster name: {reserveDetails?.name}</h2>
        {actionsReserveDetails && actionsReserveDetails?.length > 0 ? (
          <DropdownButton
            as={ButtonGroup}
            variant="outline-primary"
            title="Actions"
            intc-id="myReservationActionsDropdownButton"
          >
            {actionsReserveDetails.map((action, index) => {
              return action?.isTextItem ? (
                <div key={index}>
                  <Dropdown.Divider />
                  <Dropdown.ItemText className="fw-semibold" intc-id={`myReservationActionsDropdownItemText${index}`}>
                    {action.name}
                  </Dropdown.ItemText>
                </div>
              ) : (
                <Dropdown.Item
                  key={index}
                  onClick={() => {
                    setAction(action, reserveDetails)
                  }}
                  intc-id={`myReservationActionsDropdownItemButton${index}`}
                >
                  {action.name}
                </Dropdown.Item>
              )
            })}
          </DropdownButton>
        ) : null}
        <a href={idcConfig.REACT_APP_CLUSTER_HOW_TO_CONNECT} target="_blank" rel="noreferrer" className="link">
          <BsFillPlugFill />
          Learn how to connect
          <ExternalLink />
        </a>
      </div>
      <div className="section">
        <TabsNavigation tabs={tabs} activeTab={activeTab} setTabActive={setActiveTab} />
        <TapContent infoToDisplay={getTapContent(activeTab)} />
      </div>
    </>
  )

  // *****
  // functions
  // *****

  function getTapContent(tapNumber: number): any {
    let content = null

    switch (tapNumber) {
      case 0:
        tabDetails[activeTab].customContent = getDetailsInfoTab(tabDetails[activeTab])
        content = tabDetails[activeTab]
        break
      case 1:
        tabDetails[activeTab].customContent = getManagedNodes()
        content = tabDetails[activeTab]
        break
      case 2:
        tabDetails[activeTab].customContent = getNetwork()
        content = tabDetails[activeTab]
        break
      case 3:
        tabDetails[activeTab].customContent = (
          <ClusterStorage
            storageItems={myStorages as never[]}
            emptyStorage={emptyStorage}
            reserveDetails={reserveDetails}
            setAction={setAction}
            nodegroupsInfo={nodegroupsInfo}
            goToAddNodeGroup={goToAddNodeGroup}
          />
        )
        content = tabDetails[activeTab]
        break
      case 4:
        content = tabDetails[activeTab]
        break
      case 5:
        tabDetails[activeTab].customContent = getMetrics()
        content = tabDetails[activeTab]
        break
      default:
        tabDetails[activeTab].customContent = getDefaultContent()
        content = tabDetails[activeTab]
        break
    }

    return content
  }

  function getDefaultContent(): JSX.Element {
    return <div></div>
  }

  // *****
  // taps
  // *****

  // intel load balancer tap

  function getNetwork(): JSX.Element {
    const addNodeGroupLabel = 'Add Node Group'
    const addLoadBalancerButton = 'Add Load Balancer'
    const isClusterInActiveState = reserveDetails?.clusterstate === 'Active'
    const waitForClusterMessage = 'The Cluster is being updated, please wait before adding a load balancer'
    const waitForNodeLengthMessage =
      'To add a load balancer, please add a worker node group with at least one node first.'
    const maxLoadbalancers = clusterResourceLimit?.maxvipspercluster || 2
    let emptyGridMessage = 'Your cluster has no load balancers'
    if (!isClusterInActiveState) {
      emptyGridMessage = waitForClusterMessage
    }
    if (nodegroupsInfo.length === 0) {
      emptyGridMessage = waitForNodeLengthMessage
    }
    return (
      <>
        {myloadbalancers.length > 0 ? (
          <>
            <GridPagination data={myloadbalancers} columns={lbcolumns} loading={loading} />
            <div className="d-flex flex-xs-column flex-sm-row gap-s6 align-items-sm-center">
              <Button
                intc-id="btn-iksMyClusters-addVip"
                data-wap_ref="btn-iksMyClusters-addVip"
                variant="outline-primary"
                disabled={!isClusterInActiveState || myloadbalancers.length >= maxLoadbalancers}
                onClick={goToAddLoadBalancer}
              >
                <BsNodePlus /> {addLoadBalancerButton}
              </Button>
              <span>{`Up to ${maxLoadbalancers} load balancers max. (${
                maxLoadbalancers - myloadbalancers.length
              } remaining)`}</span>
            </div>
          </>
        ) : (
          <div className="section align-self-center align-items-center" intc-id="data-view-empty">
            <h5 className="h4">No Load Balancers found</h5>
            <p className="add-break-line lead">{emptyGridMessage}</p>
            {isClusterInActiveState && nodegroupsInfo.length === 0 ? (
              <Button
                intc-id="btn-iksMyClusters-addWorkerNodeGroup"
                data-wap_ref="btn-iksMyClusters-addWorkerNodeGroup"
                variant="outline-primary"
                onClick={() => {
                  goToAddNodeGroup()
                }}
              >
                <BsNodePlus /> {addNodeGroupLabel}
              </Button>
            ) : (
              <Button
                intc-id="btn-iksMyClusters-addVip"
                data-wap_ref="btn-iksMyClusters-addVip"
                disabled={!isClusterInActiveState}
                variant="outline-primary"
                onClick={goToAddLoadBalancer}
              >
                <BsNodePlus /> {addLoadBalancerButton}
              </Button>
            )}
          </div>
        )}
      </>
    )
  }

  // compute tap

  function getManagedNodes(): JSX.Element {
    const addNodeGroupLabel = 'Add Node Group'
    const tabTitle = `Node groups (${nodegroupsInfo.length})`
    const isClusterInActiveState = reserveDetails?.clusterstate === 'Active'
    const waitForClusterMessage = 'The Cluster is being updated, please wait before adding a worker node group'

    if (nodegroupsInfo.length === 0) {
      return (
        <div className="section align-self-center align-items-center" intc-id="data-view-empty">
          <h5 className="h4">No worker node groups found</h5>
          <p className="lead">{!isClusterInActiveState ? waitForClusterMessage : 'Your cluster has no node groups'}</p>
          <Button
            intc-id="btn-iksMyClusters-addWorkerNodeGroup"
            data-wap_ref="btn-iksMyClusters-addWorkerNodeGroup"
            variant="outline-primary"
            disabled={!isClusterInActiveState}
            onClick={() => {
              goToAddNodeGroup()
            }}
          >
            <BsNodePlus />
            {addNodeGroupLabel}
          </Button>
        </div>
      )
    }

    return (
      <>
        <div className="d-flex flex-row gap-s6">
          <h4 className="align-self-center">{tabTitle}</h4>
          <Button
            intc-id="btn-iksMyClusters-addWorkerNodeGroup"
            data-wap_ref="btn-iksMyClusters-addWorkerNodeGroup"
            variant="outline-primary"
            disabled={!isClusterInActiveState}
            onClick={() => {
              goToAddNodeGroup()
            }}
          >
            <BsNodePlus />
            {addNodeGroupLabel}
          </Button>
        </div>
        {nodegroupsInfo.map((nodeGroup, index) => (
          <ClusterNodeGroupContainer
            key={`${nodeGroup.name}-${index}`}
            nodeGroup={nodeGroup}
            debounceClusterRefresh={debounceClusterRefresh}
            getActionItemLabel={getActionItemLabel}
            isClusterInActiveState={isClusterInActiveState}
          />
        ))}
      </>
    )
  }

  // Details tap
  function getDetailsInfoTab(tapContent: any): JSX.Element {
    const fields = tapContent.fields

    const generateActionsButtons = (field: string, actions: any[]): JSX.Element => {
      let content: any = <></>
      if (reserveDetails?.clusterstate !== 'Active') {
        return content
      }
      switch (field) {
        case 'k8sversion': {
          content = actions.map((action: any, actionIndx: number) => {
            return reserveDetails?.upgradeavailable ? (
              <Button
                variant="link"
                size="sm"
                key={actionIndx}
                onClick={action.func}
                intc-id={`btn-details-tab-${field}-${action.label}`}
                data-wap_ref={`btn-details-tab-${field}-${action.label}`}
              >
                <BsQuestionCircle />
                {action.label}
              </Button>
            ) : (
              <span className="small" key={actionIndx}>
                You are already running the latest version of Kubernetes
              </span>
            )
          })
          break
        }
        case 'kubeconfig': {
          const groupSize = 2
          const groups = []
          for (let i = 0; i < actions.length; i += groupSize) {
            groups.push(actions.slice(i, i + groupSize))
          }
          content = (
            <>
              {groups.map((group, index) => (
                <div className={`d-flex ${groups.length === 1 ? 'flex-row' : 'flex-column'}`} key={index}>
                  {group.map((action, indexAction) => (
                    <Button
                      variant={action.type}
                      size="sm"
                      key={indexAction}
                      onClick={() => action.func(reserveDetails)}
                      intc-id={`btn-details-tab-${field}-${action.label}`}
                      data-wap_ref={`btn-details-tab-${field}-${action.label}`}
                    >
                      {action.name}
                    </Button>
                  ))}
                </div>
              ))}
            </>
          )
          break
        }
        default:
          content = <></>
          break
      }
      return content
    }

    return (
      <div className="row">
        {fields.map((item: any, index: number) => (
          <LabelValuePair className="col-md-4" label={item.label} key={index}>
            <div
              className={`d-flex ${item.field === 'k8sversion' && !reserveDetails?.upgradeavailable ? 'flex-column' : 'flex-column flex-sm-row gap-s3'} text-wrap`}
            >
              {item.value} {item.actions ? generateActionsButtons(item.field, item.actions) : null}
            </div>
          </LabelValuePair>
        ))}
      </div>
    )
  }

  // Metrics Tab
  function getMetrics(): JSX.Element {
    const cluster = {
      name: reserveDetails.name,
      value: reserveDetails.uuid
    }
    return <ClusterGraphsContainer clusters={[cluster]} showClusters={false} />
  }

  // variables
  const formElementskeys = []

  for (const key in upgradeForm.form) {
    const formItem = {
      ...upgradeForm.form[key]
    }

    formElementskeys.push({
      id: key,
      configInput: formItem
    })
  }

  return (
    <>
      <UpgradeVersionModal
        show={upgradeModal.show}
        centered={upgradeModal.centered}
        closeButton={upgradeModal.closeButton}
        onHide={upgradeModal.onHide}
        onChangeUpgradeForm={onChangeUpgradeForm}
        formElementskeys={formElementskeys}
        isValidUpgradeForm={isValidUpgradeForm}
        submitUpgradeK8sVersion={submitUpgradeK8sVersion}
      />
      <ActionConfirmation
        actionModalContent={actionModalContent}
        onClickModalConfirmation={actionOnModal}
        showModalActionConfirmation={showActionModal}
      />
      {loading || !reserveDetails ? spinner : clusterDetails}
    </>
  )
}

export default ClusterMyReservationsDetail
