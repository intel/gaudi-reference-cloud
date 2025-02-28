// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { BsNodePlus } from 'react-icons/bs'
import GridPagination from '../../../utils/gridPagination/gridPagination'
import ActionConfirmation from '../../../utils/modals/actionConfirmation/ActionConfirmation'
import { Button } from 'react-bootstrap'

const SuperComputerLoadBalancer = ({
  loadBalancerItems,
  isClusterInActiveState,
  goToAddLoadBalancer,
  columns,
  vipsLimit,
  actionModal,
  onActionModal
}) => {
  let content = null
  const gridFeedBack = (
    <>
      <Button
        intc-id="btn-iksMyClusters-addVip"
        data-wap_ref="btn-iksMyClusters-addVip"
        variant="outline-primary"
        onClick={goToAddLoadBalancer}
      >
        <BsNodePlus /> Add Load Balancer
      </Button>
      <span className="mx-2">{`Up to ${vipsLimit} load balancer max.`}</span>
    </>
  )
  if (loadBalancerItems.length === 0) {
    content = (
      <div className="section align-self-center align-items-center" intc-id="data-view-empty">
        <h5 className="h4">No Load Balancers found</h5>
        {isClusterInActiveState ? (
          <span className="add-break-line lead">
            The Cluster is being updated, please wait for worker node groups to be ready
          </span>
        ) : null}
        <Button
          intc-id="btn-iksMyClusters-addVip"
          data-wap_ref="btn-iksMyClusters-addVip"
          variant="outline-primary"
          onClick={goToAddLoadBalancer}
          disabled={isClusterInActiveState}
        >
          <BsNodePlus /> Add Load Balancer
        </Button>
      </div>
    )
  } else {
    content = (
      <>
        <GridPagination
          feedBack={gridFeedBack}
          data={loadBalancerItems}
          columns={columns}
          hidePaginationControl={true}
        />
        <div className="d-flex flex-xs-column flex-sm-row gap-s6 align-items-sm-center">
          <Button
            intc-id="btn-iksMyClusters-addVip"
            data-wap_ref="btn-iksMyClusters-addVip"
            variant="outline-primary"
            disabled={isClusterInActiveState || loadBalancerItems.length >= Number(vipsLimit)}
            onClick={goToAddLoadBalancer}
          >
            <BsNodePlus /> Add Load Balancer
          </Button>
          <span>{`Up to ${vipsLimit} load balancers max. (${vipsLimit - loadBalancerItems.length} remaining)`}</span>
        </div>
      </>
    )
  }

  return (
    <>
      <ActionConfirmation
        actionModalContent={actionModal}
        onClickModalConfirmation={onActionModal}
        showModalActionConfirmation={actionModal.show}
      />
      {content}
    </>
  )
}

export default SuperComputerLoadBalancer
