// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useRef } from 'react'
import { Accordion, Popover, OverlayTrigger, Button, ButtonGroup } from 'react-bootstrap'
import { BsNodeMinus, BsNodeMinusFill, BsNodePlusFill, BsInfoCircle } from 'react-icons/bs'
import GridPagination from '../../../utils/gridPagination/gridPagination'
import HowToConnect from '../../../utils/modals/howToConnect/HowToConnect'
import ActionConfirmation from '../../../utils/modals/actionConfirmation/ActionConfirmation'
import { computeCategoriesEnum, iksNodeGroupActionsEnum } from '../../../utils/Enums'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'
import idcConfig from '../../../config/configurator'
import { ReactComponent as ExternalLink } from '../../../assets/images/ExternalLink.svg'
import SearchBox from '../../../utils/searchBox/SearchBox'

const ClusterNodeGroup = (props) => {
  const columns = props.columns
  const emptyGrid = props.emptyGrid
  const nodeGroup = props.nodeGroup
  const nodes = props.nodes
  const showHowToConnectModal = props.showHowToConnectModal
  const setShowHowToConnectModal = props.setShowHowToConnectModal
  const selectedNodeDetails = props.selectedNodeDetails
  const showConfirmationModal = props.showConfirmationModal
  const openDeleteNodeGroup = props.openDeleteNodeGroup
  const openUpgradeNodeImage = props.openUpgradeNodeImage
  const openAddNode = props.openAddNode
  const openRemoveNode = props.openRemoveNode
  const confirmationModalAction = props.confirmationModalAction
  const getAction = props.getAction
  const errorModal = props.errorModal
  const onClickCloseErrorModal = props.onClickCloseErrorModal
  const canUpdateNodeGroup = props.canUpdateNodeGroup
  const getNodegroupStatusMessage = props.getNodegroupStatusMessage
  const filterText = props.filterText
  const setFilter = props.setFilter

  const nodesQuantity = nodeGroup.count

  const nodesRemaining = 10 - nodeGroup.count
  const canAddNode = nodesQuantity < 10
  const canRemoveNode = nodesQuantity > 0
  const isNodeGroupDeleting = nodeGroup.nodegroupstate === 'Deleting'
  const shouldShowAddAndDeleteNodes = nodeGroup.instanceTypeDetails.category === computeCategoriesEnum.singleNode

  const deleteNodeGroupRef = useRef(null)
  const upgradeNodeGroupRef = useRef(null)

  let gridItems = []

  if (filterText !== '') {
    const input = filterText.toLowerCase()
    gridItems = nodes.filter((item) =>
      item.name.value ? item.name.value.toLowerCase().includes(input) : item.name.toLowerCase().includes(input)
    )
    if (gridItems.length === 0) {
      gridItems = nodes.filter((item) => item.ipNbr.toLowerCase().includes(input))
    }
    if (gridItems.length === 0) {
      gridItems = nodes.filter((item) => item.imi.toLowerCase().includes(input))
    }
    if (gridItems.length === 0) {
      gridItems = nodes.filter((item) => item.status.value.state.toLowerCase().includes(input))
    }
  } else {
    gridItems = nodes
  }

  const getConfirmationDialogTitle = () => {
    switch (confirmationModalAction) {
      case iksNodeGroupActionsEnum.deleteNodeGroup:
        return 'Delete node group'
      case iksNodeGroupActionsEnum.upgradeImage:
        return 'Upgrade IMI for Nodegroup'
      case iksNodeGroupActionsEnum.addNode:
      case iksNodeGroupActionsEnum.removeNode:
        return 'Change Node Group Count'
      default:
        return ''
    }
  }

  const popover = (comp) => {
    return <Popover className="p-3">{comp}</Popover>
  }

  const getConfirmationDialogQuestion = () => {
    if (!canUpdateNodeGroup()) {
      return 'Can not modify worker node group while it is updating.'
    }

    switch (confirmationModalAction) {
      case iksNodeGroupActionsEnum.deleteNodeGroup:
        return `Are you sure you want to delete node group ${nodeGroup.name}?`
      case iksNodeGroupActionsEnum.upgradeImage:
        return `Are you sure you want to update the image of node group ${nodeGroup.name}?`
      case iksNodeGroupActionsEnum.addNode:
        return `The node count of node group ${nodeGroup.name} will change from ${nodeGroup.count} to ${
          nodeGroup.count + 1
        }. Continue?`
      case iksNodeGroupActionsEnum.removeNode:
        return `The node count of node group ${nodeGroup.name} will change from ${nodeGroup.count} to ${
          nodeGroup.count - 1
        }. Continue?`
      default:
        return ''
    }
  }

  const getConfirmationDialogButtonLabel = () => {
    if (!canUpdateNodeGroup()) {
      return 'Continue'
    }

    switch (confirmationModalAction) {
      case iksNodeGroupActionsEnum.deleteNodeGroup:
        return 'Delete'
      case iksNodeGroupActionsEnum.upgradeImage:
        return 'Update'
      case iksNodeGroupActionsEnum.addNode:
      case iksNodeGroupActionsEnum.removeNode:
        return 'Continue'
      default:
        return ''
    }
  }

  const getConfirmationDialogButtonVariant = () => {
    if (!canUpdateNodeGroup()) {
      return 'primary'
    }

    switch (confirmationModalAction) {
      case iksNodeGroupActionsEnum.deleteNodeGroup:
      case iksNodeGroupActionsEnum.removeNode:
        return 'danger'
      case iksNodeGroupActionsEnum.upgradeImage:
      case iksNodeGroupActionsEnum.addNode:
        return 'primary'
      default:
        return ''
    }
  }

  const getConfirmationDialogFeedback = () => {
    if (!canUpdateNodeGroup()) {
      return ''
    }
    switch (confirmationModalAction) {
      case iksNodeGroupActionsEnum.deleteNodeGroup:
        return 'If the node group is running it will be stopped. All your information will be lost.'
      default:
        return ''
    }
  }

  useEffect(() => {
    const mouseEnter = () => {
      const accordionButton = upgradeNodeGroupRef.current.parentElement
      accordionButton.setAttribute('data-bs-toggle', '')
    }
    const mouseLeave = () => {
      const accordionButton = upgradeNodeGroupRef.current.parentElement
      accordionButton.setAttribute('data-bs-toggle', 'collapse')
    }

    if (upgradeNodeGroupRef && upgradeNodeGroupRef.current) {
      upgradeNodeGroupRef.current.addEventListener('mouseenter', mouseEnter)
      upgradeNodeGroupRef.current.addEventListener('mouseleave', mouseLeave)
    }

    return () => {
      if (upgradeNodeGroupRef && upgradeNodeGroupRef.current) {
        upgradeNodeGroupRef.current.removeEventListener('mouseenter', mouseEnter)
        upgradeNodeGroupRef.current.removeEventListener('mouseleave', mouseLeave)
      }
    }
  }, [upgradeNodeGroupRef])

  useEffect(() => {
    const mouseEnter = () => {
      const accordionButton = deleteNodeGroupRef.current.parentElement
      accordionButton.setAttribute('data-bs-toggle', '')
    }
    const mouseLeave = () => {
      const accordionButton = deleteNodeGroupRef.current.parentElement
      accordionButton.setAttribute('data-bs-toggle', 'collapse')
    }

    if (deleteNodeGroupRef && deleteNodeGroupRef.current) {
      deleteNodeGroupRef.current.addEventListener('mouseenter', mouseEnter)
      deleteNodeGroupRef.current.addEventListener('mouseleave', mouseLeave)
    }

    return () => {
      if (deleteNodeGroupRef && deleteNodeGroupRef.current) {
        deleteNodeGroupRef.current.removeEventListener('mouseenter', mouseEnter)
        deleteNodeGroupRef.current.removeEventListener('mouseleave', mouseLeave)
      }
    }
  }, [deleteNodeGroupRef])

  return (
    <>
      <HowToConnect
        flowType="compute"
        data={selectedNodeDetails}
        showHowToConnectModal={showHowToConnectModal}
        onClickHowToConnect={setShowHowToConnectModal}
      />
      <ActionConfirmation
        actionModalContent={{
          label: getConfirmationDialogTitle(),
          question: getConfirmationDialogQuestion(),
          buttonLabel: getConfirmationDialogButtonLabel(),
          buttonVariant: getConfirmationDialogButtonVariant(),
          feedback: getConfirmationDialogFeedback(),
          name: nodeGroup.name
        }}
        showModalActionConfirmation={showConfirmationModal}
        onClickModalConfirmation={getAction}
      />
      <ErrorModal
        showModal={errorModal.showErrorModal}
        titleMessage={errorModal?.titleMessage}
        description={errorModal.errorDescription}
        message={errorModal.errorMessage}
        hideRetryMessage={errorModal.errorHideRetryMessage}
        onClickCloseErrorModal={onClickCloseErrorModal}
      />
      <Accordion intc-id={`accordion-${nodeGroup.name}-nodegroup`} className="mt-s4 w-100" alwaysOpen>
        <Accordion.Item className="border-0">
          <Accordion.Header>
            <div className="d-flex flex-row gap-s6">
              <h5 className="align-self-center">Node Group: {nodeGroup.name}</h5>
              {nodeGroup.upgradeavailable ? (
                <a ref={upgradeNodeGroupRef} className="btn btn-outline-primary" onClick={openUpgradeNodeImage}>
                  Upgrade Image
                </a>
              ) : null}
              <a
                ref={deleteNodeGroupRef}
                className={isNodeGroupDeleting ? 'btn btn-outline-primary disabled' : 'btn btn-outline-primary'}
                onClick={openDeleteNodeGroup}
                style={isNodeGroupDeleting ? { pointerEvents: 'none' } : {}}
              >
                <BsNodeMinus />
                Delete
              </a>
            </div>
          </Accordion.Header>
          <Accordion.Body className="section">
            <div className="d-flex flex-row align-items-center gap-s6">
              <div className="d-md-flex justify-content-md-start gap-s4">
                <label className="fw-semibold" htmlFor={`nodegroup-${nodeGroup.name}-instanceType`}>
                  Node Instance Type:
                </label>
                <span className="fw-normal" id={`nodegroup-${nodeGroup.name}-instanceType`}>
                  {nodeGroup?.instanceTypeDetails.displayName}
                  <a
                    href={idcConfig.REACT_APP_INSTANCE_SPEC}
                    target="_blank"
                    rel="noreferrer"
                    className="ps-s4"
                    style={{
                      minWidth: 'fit-content'
                    }}
                  >
                    <span>View spec</span>
                    <ExternalLink />
                  </a>
                </span>
              </div>
              <div className="d-md-flex justify-content-md-start gap-s4">
                <label className="fw-semibold" htmlFor={`nodegroup-${nodeGroup.name}-state`}>
                  State:
                </label>
                <span className="fw-normal" id={`nodegroup-${nodeGroup.name}-state`}>
                  {`${nodeGroup.nodegroupstate} - ${getNodegroupStatusMessage(nodeGroup.nodegroupstatus)}`}
                </span>
              </div>
              {nodeGroup.userdataurl ? (
                <div className="d-md-flex justify-content-md-start gap-s4">
                  <label className="fw-semibold" htmlFor={`nodegroup-${nodeGroup.name}-state`}>
                    User data URL:
                  </label>
                  <OverlayTrigger trigger="focus" placement="right" overlay={popover(nodeGroup.userdataurl)}>
                    <Button aria-label="User data url information" variant="icon-link">
                      <BsInfoCircle />
                    </Button>
                  </OverlayTrigger>
                </div>
              ) : null}
            </div>
            {nodes.length > 0 && (
              <div className="filter">
                <div className="d-flex justify-content-end">
                  <SearchBox
                    intc-id="searchNodes"
                    value={filterText}
                    onChange={setFilter}
                    placeholder="Search nodes..."
                    aria-label="Type to search node.."
                  />
                </div>
              </div>
            )}
            <GridPagination data={gridItems} columns={columns} emptyGrid={emptyGrid} loading={false} />
            {shouldShowAddAndDeleteNodes ? (
              <ButtonGroup>
                <Button
                  intc-id="btn-iksMyClusters-addNode"
                  data-wap_ref="btn-iksMyClusters-addNode"
                  variant="outline-primary"
                  disabled={!canAddNode || isNodeGroupDeleting}
                  onClick={openAddNode}
                >
                  <BsNodePlusFill />
                  Add Node
                </Button>
                <Button
                  intc-id="btn-iksMyClusters-removeNode"
                  data-wap_ref="btn-iksMyClusters-removeNode"
                  variant="outline-primary"
                  disabled={!canRemoveNode || isNodeGroupDeleting}
                  onClick={openRemoveNode}
                >
                  <BsNodeMinusFill />
                  Delete Node
                </Button>
                <span>{`Up to 10 nodes max. (${nodesRemaining} remaining)`}</span>
              </ButtonGroup>
            ) : null}
          </Accordion.Body>
        </Accordion.Item>
      </Accordion>
    </>
  )
}

export default ClusterNodeGroup
