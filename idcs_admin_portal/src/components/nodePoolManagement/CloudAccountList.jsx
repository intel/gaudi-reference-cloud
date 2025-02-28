import GridPagination from '../../utility/gridPagination/gridPagination'
import SearchBox from '../../utility/searchBox/SearchBox'
import OnConfirmModal from '../../utility/modals/onConfirmModal/OnConfirmModal'
import OnSubmitModal from '../../utility/modals/onSubmitModal/OnSubmitModal'
import { NavLink } from 'react-router-dom'

const CloudAccountList = (props) => {
  // props Variables
  const columns = props.columns
  const cloudAccounts = props.cloudAccounts
  const emptyGrid = props.emptyGrid
  const filterText = props.filterText
  const loading = props.loading
  const showLoader = props.showLoader
  const confirmModalData = props.confirmModalData
  const poolId = props.poolId

  // Props Functions
  const setFilter = props.setFilter
  const onSubmit = props.onSubmit

  let gridItems = []

  if (filterText !== '') {
    const input = filterText.toLowerCase()

    if (gridItems.length === 0) {
      gridItems = cloudAccounts.filter((item) => item.poolId.toLowerCase().includes(input))
    }
    if (gridItems.length === 0) {
      gridItems = cloudAccounts.filter((item) => item.cloudAccountId.toLowerCase().includes(input))
    }
    if (gridItems.length === 0) {
      gridItems = cloudAccounts.filter((item) => item.createAdmin.toLowerCase().includes(input))
    }
  } else {
    gridItems = cloudAccounts
  }

  return (
    <>
      <div className="section">
        <NavLink to={`/npm/pools/edit/${poolId}`} className="btn btn-link ps-0">
          ⟵ Back to Pool {poolId}
        </NavLink>
      </div>
      <OnSubmitModal showModal={showLoader.isShow} message={showLoader.message}></OnSubmitModal>
      <OnConfirmModal confirmModalData={confirmModalData} onSubmit={onSubmit}></OnConfirmModal>
      <div className={`filter ${!cloudAccounts || cloudAccounts.length === 0 ? 'd-none' : ''}`}>
        <NavLink
          to={`/npm/pools/accounts/add/${poolId}`}
          className="btn btn-primary"
          aria-label="btn-navigate-add-cloud-account"
          intc-id="btn-navigate-add-cloud-account"
        >
          Add Cloud Account to {poolId}
        </NavLink>
        <SearchBox
          intc-id="searchCloudAccounts"
          value={filterText}
          onChange={setFilter}
          placeholder="Search cloud accounts..."
          aria-label="Type to search cloud accounts..."
        />
      </div>
      <div className="section">
        <GridPagination data={gridItems} columns={columns} loading={loading} emptyGrid={emptyGrid} />
      </div>
    </>
  )
}

export default CloudAccountList
