// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useRef } from 'react'
import { Accordion, Button, ButtonGroup } from 'react-bootstrap'
import LabelValuePair from '../../../utils/labelValuePair/LabelValuePair'
import GridPagination from '../../../utils/gridPagination/gridPagination'
import { BsNodeMinus } from 'react-icons/bs'
import SearchBox from '../../../utils/searchBox/SearchBox'

const SuperComputerWorkerNodeInfo = ({
  groupActionsOptions,
  nodeInfo,
  nodeGroupIndex,
  type,
  columns,
  nodes,
  emptyGrid,
  setGroupAction,
  filterText,
  setFilter,
  loading
}) => {
  const nodeGroupState = nodeInfo.nodeGroupState
  const { instanceType, userDataUrl, nodeStatus, groupName, upgradeAvailable } = nodeInfo
  const isNodeGroupDeleting = nodeGroupState === 'Deleting'

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

  const workerNodeContent = (
    <>
      <div className="row">
        <div className="col-md-4 col-sm-12 d-flex flex-column">
          <LabelValuePair label="Node Instance Type:">{instanceType}</LabelValuePair>
        </div>
        <div className="col-md-4 col-sm-12 d-flex flex-column">
          <LabelValuePair label="User data URL:">{userDataUrl || 'N/A'}</LabelValuePair>
        </div>
        <div className="col-md-4 col-sm-12 d-flex flex-column">
          <LabelValuePair label="Status:">{nodeStatus}</LabelValuePair>
        </div>
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
      <GridPagination data={gridItems} columns={columns} emptyGrid={emptyGrid} loading={loading} />
      <>
        {type === 'compute' ? (
          <div className="d-flex flex-row">
            {groupActionsOptions.map((action, indexBtn) => (
              <ButtonGroup key={indexBtn} className="btn-group me-3" role="group" aria-label="First group">
                <Button
                  intc-id={`btn-action-${indexBtn}`}
                  data-wap_ref={`btn-action-${indexBtn}`}
                  disabled={!nodeGroupState || !action.status.some((x) => x === nodeGroupState)}
                  onClick={() => action.function(action, nodeGroupIndex)}
                  variant="outline-primary"
                >
                  {action.name}
                </Button>
              </ButtonGroup>
            ))}
          </div>
        ) : null}
      </>
    </>
  )

  if (type === 'ai') {
    return (
      <div className="section">
        <h3 className="h5">AI nodes </h3>
        {workerNodeContent}
      </div>
    )
  }

  return (
    <Accordion intc-id={`accordion-${groupName}-nodegroup`} className="mt-s4 w-100" alwaysOpen>
      <Accordion.Item className="border-0">
        <Accordion.Header>
          <div className="d-flex flex-row gap-s6">
            <h4 className="align-self-center">Node Group: {groupName}</h4>
            {upgradeAvailable ? (
              <a
                intc-id="btn-upgrade"
                data-wap_ref="btn-upgrade"
                ref={upgradeNodeGroupRef}
                className="btn btn-outline-primary"
                onClick={() => setGroupAction({ id: 'upgradeImage' }, nodeGroupIndex)}
              >
                Upgrade Image
              </a>
            ) : null}
            <a
              intc-id="btn-delete-node"
              data-wap_ref="btn-delete-node"
              ref={deleteNodeGroupRef}
              className={isNodeGroupDeleting ? 'btn btn-outline-primary disabled' : 'btn btn-outline-primary'}
              style={isNodeGroupDeleting ? { pointerEvents: 'none' } : {}}
              disabled={isNodeGroupDeleting}
              onClick={() => setGroupAction({ id: 'deleteGroup' }, nodeGroupIndex)}
            >
              <BsNodeMinus />
              Delete
            </a>
          </div>
        </Accordion.Header>
        <Accordion.Body className="section">{workerNodeContent}</Accordion.Body>
      </Accordion.Item>
    </Accordion>
  )
}

export default SuperComputerWorkerNodeInfo
