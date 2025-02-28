// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { BsNodePlusFill } from 'react-icons/bs'
import ActionConfirmation from '../../../utils/modals/actionConfirmation/ActionConfirmation'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'
import CustomInput from '../../../utils/customInput/CustomInput'
import { Button } from 'react-bootstrap'
import SuperComputerWorkerNodeInfoContainer from '../../../containers/superComputer/SuperComputerWorkerNodeInfoContainer'

const SuperComputerWorkerNodes = ({
  tabs,
  activeTab,
  onChangeFilter,
  addNodeGroup,
  setGroupAction,
  actionModal,
  onActionModal,
  aiWorkerGroupItems,
  computeWorkerGroupItems,
  groupActionsOptions,
  isClusterInActiveState,
  errorModal,
  onClickCloseErrorModal,
  nodeGroupSelection,
  isNodeGroupDeleting
}) => {
  let content = null
  const currentTab = tabs[activeTab]

  switch (currentTab.id) {
    case 'ai': {
      content = aiWorkerGroupItems.map((item, index) => (
        <SuperComputerWorkerNodeInfoContainer
          key={index}
          nodeInfo={item}
          nodeGroupIndex={index}
          setGroupAction={setGroupAction}
          type={currentTab.id}
          isClusterInActiveState={isClusterInActiveState}
        />
      ))
      break
    }
    default:
      content = (
        <>
          <div className="section flex-row">
            <h3 className="h5" intc-id="GcTittle">
              General compute nodes
            </h3>
            <Button
              intc-id="btnAddNodeGroup"
              data-wap_ref="btnAddNodeGroup"
              onClick={() => addNodeGroup()}
              variant="outline-primary"
              disabled={!isClusterInActiveState}
            >
              <BsNodePlusFill /> Add node group
            </Button>
          </div>
          <div className="d-flex flex-column gap-s6 w-100">
            {computeWorkerGroupItems.map((item, index) => (
              <SuperComputerWorkerNodeInfoContainer
                key={index}
                nodeInfo={item}
                nodeGroupIndex={index}
                setGroupAction={setGroupAction}
                type={currentTab.id}
                isClusterInActiveState={isClusterInActiveState}
              />
            ))}
          </div>
        </>
      )
      break
  }

  let mainContent = null
  if (computeWorkerGroupItems.length > 0 || aiWorkerGroupItems.length > 0) {
    mainContent = (
      <div className="row">
        <div className="section col-md-2 col-xs-12">
          <CustomInput {...nodeGroupSelection} value={activeTab} />
        </div>
        <div className="col-md-10 col-xs-12">{content}</div>
      </div>
    )
  } else {
    mainContent = (
      <div className="section align-self-center align-items-center">
        <h5 className="h4" intc-id="NodeGroupTitleEmpty">
          No worker node groups found
        </h5>
        {!isClusterInActiveState ? (
          <span className="add-break-line lead">
            The Cluster is being updated, please wait for worker node groups to be ready
          </span>
        ) : null}
        <Button
          intc-id="btnAddNodeGroupEmpty"
          data-wap_ref="btnAddNodeGroupEmpty"
          disabled={!isClusterInActiveState}
          onClick={() => addNodeGroup()}
          variant="outline-primary"
        >
          <BsNodePlusFill /> Add node group
        </Button>
      </div>
    )
  }
  return (
    <>
      <ActionConfirmation
        actionModalContent={actionModal}
        onClickModalConfirmation={onActionModal}
        showModalActionConfirmation={actionModal.show}
      />
      <ErrorModal
        showModal={errorModal.show}
        titleMessage={errorModal?.titleMessage}
        description={errorModal.errorDescription}
        message={errorModal.errorMessage}
        hideRetryMessage={errorModal.errorHideRetryMessage}
        onClickCloseErrorModal={onClickCloseErrorModal}
      />
      {mainContent}
    </>
  )
}

export default SuperComputerWorkerNodes
