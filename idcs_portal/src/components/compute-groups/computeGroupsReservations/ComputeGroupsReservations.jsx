// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import GridPagination from '../../../utils/gridPagination/gridPagination'
import HowToConnect from '../../../utils/modals/howToConnect/HowToConnect'
import ActionConfirmation from '../../../utils/modals/actionConfirmation/ActionConfirmation'
import { NavLink } from 'react-router-dom'
import SearchBox from '../../../utils/searchBox/SearchBox'

const ComputeGroupsReservations = (props) => {
  // props
  const columns = props.columns
  const myreservations = props.myreservations
  const selectedNodeDetails = props.selectedNodeDetails
  const showHowToConnectModal = props.showHowToConnectModal
  const setShowHowToConnectModal = props.setShowHowToConnectModal
  const showActionModal = props.showActionModal
  const setShowActionModal = props.setShowActionModal
  const actionModalContent = props.actionModalContent
  const emptyGrid = props.emptyGrid
  const loading = props.loading
  const filterText = props.filterText
  const setFilter = props.setFilter

  let gridItems = []

  if (filterText !== '') {
    const input = filterText.toLowerCase()
    gridItems = myreservations.filter((item) => item.name.value.toLowerCase().includes(input))
    if (gridItems.length === 0) {
      gridItems = myreservations.filter((item) => item.instanceCount.toLowerCase().includes(input))
    }
    if (gridItems.length === 0) {
      gridItems = myreservations.filter((item) =>
        item.displayName.value.instanceTypeDetails.name.toLowerCase().includes(input)
      )
    }
  } else {
    gridItems = myreservations
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
      <div className={`filter ${!myreservations || myreservations.length === 0 ? 'd-none' : ''}`}>
        <NavLink to="/compute-groups/reserve" className="btn btn-primary" intc-id={'btn-navigate Launch Instance'}>
          Launch instance group
        </NavLink>
        <div className="d-flex justify-content-end">
          <SearchBox
            intc-id="searchInstances"
            value={filterText}
            onChange={setFilter}
            placeholder="Search instances..."
            aria-label="Type to search instance.."
          />
        </div>
      </div>
      <div className="section">
        <GridPagination data={gridItems} columns={columns} emptyGrid={emptyGrid} loading={loading} fixedFirstColumn />
      </div>
    </>
  )
}

export default ComputeGroupsReservations
