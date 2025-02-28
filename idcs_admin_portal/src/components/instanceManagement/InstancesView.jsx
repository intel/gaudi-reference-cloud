// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import GridPagination from '../../utility/gridPagination/gridPagination'
import { Button } from 'react-bootstrap'
import InstanceTerminateConfirm from '../../utility/modals/instanceManagement/InstanceTerminateConfirm'
import OnSubmitModal from '../../utility/modals/onSubmitModal/OnSubmitModal'
import SearchBox from '../../utility/searchBox/SearchBox'
import SelectedAccountCard from '../../utility/selectedAccountCard/SelectedAccountCard'

const InstancesView = (props) => {
  const deleteInstanceMessage = 'Are you sure you want to terminate this instance?'

  // props.
  const instances = props.instances
  const instanceGroups = props.instanceGroups
  const instancesColumns = props.instancesColumns
  const instanceGroupsColumns = props.instanceGroupsColumns
  const loadingInstances = props.loadingInstances
  const instancesEmptyGrid = props.instancesEmptyGrid
  const instanceGroupsEmptyGrid = props.instanceGroupsEmptyGrid
  const cloudAccount = props.cloudAccount
  const instancesModalData = props.instancesModalData
  const showLoader = props.showLoader
  const selectedCloudAccount = props.selectedCloudAccount
  const cloudAccountError = props.cloudAccountError

  // props functions.
  const backToHome = props.backToHome
  const handleSearchInputChange = props.handleSearchInputChange
  const handleSubmit = props.handleSubmit
  const onInstanceTerminateSubmit = props.onInstanceTerminateSubmit

  // local variables
  const customStyle = { minWidth: 'fit-content' }

  return (
    <>
        <div className="section">
          <Button variant="link" className='p-s0' onClick={backToHome}>
            ⟵ Back to Home
          </Button>
        </div>
        <OnSubmitModal showModal={showLoader.isShow} message={showLoader.message}></OnSubmitModal>
        <InstanceTerminateConfirm
          instancesModalData={instancesModalData}
          message={deleteInstanceMessage}
          onInstanceTerminateSubmit={onInstanceTerminateSubmit}
        />
        <div className="section">
          <div className='row'>
            <div className='col-lg-6'>
              <div className="row">
                <div className='col-12 col-lg-5 fs-6 mt-4' intc-id="filterText">
                  Cloud Account ID: *
                </div>
                <div className='col-12 col-lg-7'>
                  <div className="d-inline-flex">
                    <SearchBox
                      intc-id="searchCloudAccounts"
                      placeholder={'Enter Cloud Account Email or ID'}
                      aria-label="Type to search cloud account..."
                      value={ cloudAccount || ''}
                      onChange={handleSearchInputChange}
                      onClickSearchButton={handleSubmit}
                    />
                  </div>
                  <div className='invalid-feedback pt-s3' intc-id='terminateInstanceError'>{cloudAccountError}</div>
                </div>
              </div>
            </div>
            {selectedCloudAccount && (
              <div className='col-lg-3' style={customStyle}>
                <SelectedAccountCard selectedCloudAccount={selectedCloudAccount} className='col-lg-3' style={customStyle}/>
              </div>
            )}
          </div>
        </div>
        <div className='section' intc-id='instancesGrid'>
            {<h2 className='h4'>Instances</h2>}
            <GridPagination data={instances} columns={instancesColumns} loadingInstances={loadingInstances} emptyGrid={instancesEmptyGrid} />
        </div>
        <div className="section" intc-id='instanceGroupsGrid'>
            {<h2 className='h4'>Instance Groups</h2>}
            <GridPagination data={instanceGroups} columns={instanceGroupsColumns} loadingInstances={loadingInstances} emptyGrid={instanceGroupsEmptyGrid} />
        </div>
    </>
  )
}

export default InstancesView
