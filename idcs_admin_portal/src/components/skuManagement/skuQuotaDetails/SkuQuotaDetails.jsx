import { NavLink } from 'react-router-dom'
import GridPagination from '../../../utility/gridPagination/gridPagination'
import SearchBox from '../../../utility/searchBox/SearchBox'
import SkuQuotaCreateConfirm from '../../../utility/modals/skuManagement/SkuQuotaCreateConfirm'
import OnSubmitModal from '../../../utility/modals/onSubmitModal/OnSubmitModal'
import { Button } from 'react-bootstrap'
import CustomInput from '../../../utility/customInput/CustomInput'

const SkuQuotaDetails = (props) => {
  // props Variables
  const viewMode = props.viewMode
  const columns = props.columns
  const adminSkuQuotas = props.adminSkuQuotas
  const emptyGrid = props.emptyGrid
  const filterText = props.filterText
  const loading = props.loading
  const showLoader = props.showLoader
  const confirmModalData = props.confirmModalData
  const selectedServiceType = props.selectedServiceType
  const serviceTypes = props.serviceTypes
  // Props Functions
  const setFilter = props.setFilter
  const onSubmit = props.onSubmit
  const backToHome = props.backToHome
  const setServiceTypeFilter = props.setServiceTypeFilter

  // local variables
  const customStyle = { minWidth: '18rem' }

  // functions

  function getFilteredData() {
    let filteredData = []

    if (selectedServiceType && filterText) {
      for (const index in adminSkuQuotas) {
        const quota = { ...adminSkuQuotas[index] }
        if (getFilterCheck(filterText, quota) && getFilterCheck(selectedServiceType, quota)) {
          filteredData.push(quota)
        }
      }
    } else if (filterText !== '' || selectedServiceType) {
      for (const index in adminSkuQuotas) {
        const quota = { ...adminSkuQuotas[index] }
        if (getFilterCheck(filterText, quota) || getFilterCheck(selectedServiceType, quota)) {
          filteredData.push(quota)
        }
      }
    } else {
      filteredData = [...adminSkuQuotas]
    }

    return filteredData
  }

  const gridItems = getFilteredData()

  function getFilterCheck(filterValue, data) {
    if (!filterValue) return false
    filterValue = filterValue.toLowerCase()

    return data.family.toString().toLowerCase().indexOf(filterValue) > -1 ||
    data.instanceName.toString().toLowerCase().indexOf(filterValue) > -1 ||
    data.instanceType.toString().toLowerCase().indexOf(filterValue) > -1 ||
    data.name.toString().toLowerCase().indexOf(filterValue) > -1 ||
    data.serviceType.toString().toLowerCase().indexOf(filterValue) > -1 ||
    data.cloudAccountId.toString().toLowerCase().indexOf(filterValue) > -1 ||
    data.creator.toString().toLowerCase().indexOf(filterValue) > -1
  }

  return (
    <>
      {!viewMode && <div className="section">
          <Button variant="link" className='p-s0' onClick={() => backToHome()}>
            ⟵ Back to Home
          </Button>
        </div>
      }
      <OnSubmitModal showModal={showLoader.isShow} message={showLoader.message}></OnSubmitModal>
      <SkuQuotaCreateConfirm confirmModalData={confirmModalData} onSubmit={onSubmit}></SkuQuotaCreateConfirm>
      <div className="filter flex-wrap">
        <NavLink
          to="/camanagement/create"
          className="btn btn-primary"
          intc-id={'btn-navigate-SkuQuotaAssignment'}>
            Cloud Account Assignment
        </NavLink>
        <div className="d-flex flex-sm-row flex-column gap-s6">
          <div className='d-flex flex-sm-row flex-column gap-s4 align-items-sm-center'>
            <span>Service Type</span>
            <div style={customStyle}>
              <CustomInput
                type='dropdown'
                hiddenLabel={true}
                label='Service Type'
                placeholder='Service Type'
                fieldSize='medium'
                value={selectedServiceType}
                options={serviceTypes}
                onChanged={setServiceTypeFilter}
              />
            </div>
          </div>
          <SearchBox
            intc-id="searchCloudAccounts"
            value={filterText}
            onChange={setFilter}
            placeholder="Search accounts..."
            aria-label="Type to search cloud accounts.."
          />
        </div>
      </div>
      <div className="section">
        <GridPagination data={gridItems} columns={columns} loading={loading} emptyGrid={emptyGrid} />
      </div>
    </>
  )
}

export default SkuQuotaDetails
