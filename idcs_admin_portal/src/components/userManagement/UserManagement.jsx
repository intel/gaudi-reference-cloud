import GridPagination from '../../utility/gridPagination/gridPagination'
import SearchBox from '../../utility/searchBox/SearchBox'
import OnConfirmModal from '../../utility/modals/onConfirmModal/OnConfirmModal'
import OnSubmitModal from '../../utility/modals/onSubmitModal/OnSubmitModal'
import { Button } from 'react-bootstrap'

const UserManagement = (props) => {
  // props Variables
  const columns = props.columns
  const users = props.users
  const emptyGrid = props.emptyGrid
  const filterText = props.filterText
  const loading = props.loading
  const showLoader = props.showLoader
  const confirmModalData = props.confirmModalData

  // Props Functions
  const setFilter = props.setFilter
  const onSubmit = props.onSubmit
  const backToHome = props.backToHome

  function getFilteredData() {
    let filteredData = []

    if (filterText !== '') {
      for (const index in users) {
        const quota = { ...users[index] }
        if (getFilterCheck(filterText, quota)) {
          filteredData.push(quota)
        }
      }
    } else {
      filteredData = [...users]
    }

    return filteredData
  }

  const gridItems = getFilteredData()

  function getFilterCheck(filterValue, data) {
    filterValue = filterValue.toLowerCase()

    return data.id.toString().toLowerCase().indexOf(filterValue) > -1 ||
    data.owner.toString().toLowerCase().indexOf(filterValue) > -1 ||
    data.type.toString().toLowerCase().indexOf(filterValue) > -1 ||
    data.restricted.toString().toLowerCase().indexOf(filterValue) > -1 ||
    data.adminName.toString().toLowerCase().indexOf(filterValue) > -1
  }

  return (
    <>
      <div className="section">
        <Button variant="link" className='p-s0' onClick={() => backToHome()}>
          ⟵ Back to Home
        </Button>
      </div>
      <OnSubmitModal showModal={showLoader.isShow} message={showLoader.message}></OnSubmitModal>
      <OnConfirmModal confirmModalData={confirmModalData} onSubmit={onSubmit}></OnConfirmModal>
        <div className="filter flex-wrap align-items-center">
          <h2 className='h4'>Cloud Account Details</h2>
          <SearchBox
            intc-id="searchCloudAccounts"
            value={filterText}
            onChange={setFilter}
            placeholder="Search accounts..."
            aria-label="Type to search cloud accounts.."
          />
        </div>
      <div className="section">
        <GridPagination data={gridItems} columns={columns} loading={loading} emptyGrid={emptyGrid} />
      </div>
    </>
  )
}

export default UserManagement
