import GridPagination from '../../utility/gridPagination/gridPagination'
import OnConfirmModal from '../../utility/modals/onConfirmModal/OnConfirmModal'
import OnSubmitModal from '../../utility/modals/onSubmitModal/OnSubmitModal'
import { Button } from 'react-bootstrap'
import { NavLink } from 'react-router-dom'
import SearchBox from '../../utility/searchBox/SearchBox'

const BannerDetailsList = (props) => {
  // props Variables
  const columns = props.columns
  const banners = props.banners
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
      for (const index in banners) {
        const quota = { ...banners[index] }
        if (getFilterCheck(filterText, quota)) {
          filteredData.push(quota)
        }
      }
    } else {
      filteredData = [...banners]
    }

    return filteredData
  }

  const gridItems = getFilteredData()

  function getFilterCheck(filterValue, data) {
    filterValue = filterValue.toLowerCase()

    return data?.id?.toString().toLowerCase().indexOf(filterValue) > -1 ||
    data?.type?.toString().toLowerCase().indexOf(filterValue) > -1 ||
    data?.message?.toString().toLowerCase().indexOf(filterValue) > -1 ||
    data?.userTypes?.toString().toLowerCase().indexOf(filterValue) > -1 ||
    data?.regions?.toString().toLowerCase().indexOf(filterValue) > -1 ||
    data?.title?.toString().toLowerCase().indexOf(filterValue) > -1
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
      <div className="filter">
        <NavLink
          to="/bannermanagement/create"
          className="btn btn-primary"
          intc-id={'btn-navigate-AlertCreate'}>
            Create New Banner
        </NavLink>
        <SearchBox
          intc-id="searchBanners"
          value={filterText}
          onChange={setFilter}
          placeholder="Search banners..."
          aria-label="Type to search banners.."
        />
      </div>
      <div className="section">
        <GridPagination data={gridItems} columns={columns} loading={loading} emptyGrid={emptyGrid} />
      </div>
    </>
  )
}

export default BannerDetailsList
