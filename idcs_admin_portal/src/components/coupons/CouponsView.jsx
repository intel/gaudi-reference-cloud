import moment from 'moment'
import ExportDataToCSV from '../../utility/ExportDataToCSV'
import GridPagination from '../../utility/gridPagination/gridPagination'
import RedemptionsView from './RedemptionsView'
import SearchBox from '../../utility/searchBox/SearchBox'
import { Button } from 'react-bootstrap'
import DisableCouponModal from '../../utility/modals/coupons/DisableCouponModal'

const CouponsView = (props) => {
  const dateFormat = 'M/D/YYYY HH:mm:ss'
  let redemptionsSection = null

  // Props Variable
  const coupons = props.coupons
  const filterText = props.filterText
  const emptyGrid = props.emptyGrid
  const loading = props.loading
  const isDisableCouponModalOpen = props.isDisableCouponModalOpen
  const coupon = props.coupon
  const redemptionsDetails = props.redemptionsDetails
  const redemptionsColumns = props.redemptionsColumns
  const couponsColumn = props.couponsColumn
  const allCouponsWithRedeemptionsCSV = props.allCouponsWithRedeemptionsCSV
  const csvHeaders = props.csvHeaders

  // Props Function
  const setFilter = props.setFilter
  const onCancel = props.onCancel
  const triggerHideDisableCouponModal = props.triggerHideDisableCouponModal
  const handleDisableCoupon = props.handleDisableCoupon

  function getFilteredData() {
    let filteredData = []

    if (filterText !== '') {
      for (const index in coupons) {
        const coupon = { ...coupons[index] }
        if (getFilterCheck(filterText, coupon)) {
          filteredData.push(coupon)
        }
      }
    } else {
      filteredData = [...coupons]
    }

    return filteredData
  }

  const gridItems = getFilteredData()

  function getFilterCheck(filterValue, data) {
    filterValue = filterValue.toLowerCase()

    return (
      data.code.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data.creator.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data.amount.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data.numUses.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data.numRedeemed.toString().toLowerCase().indexOf(filterValue) > -1 ||
      moment(data.created).format(dateFormat).toLowerCase().indexOf(filterValue) > -1 ||
      moment(data.start).format(dateFormat).toLowerCase().indexOf(filterValue) > -1 ||
      moment(data.expires).format(dateFormat).toLowerCase().indexOf(filterValue) > -1
    )
  }

  if (redemptionsDetails) {
    redemptionsSection = (
      <div className="section" intc-id="Redemptions">
        <h3>Redemptions</h3>
        <RedemptionsView redemptionsDetails={redemptionsDetails} redemptionsColumns={redemptionsColumns} />
      </div>
    )
  }

  return (
    <>
      <div className="section">
        <Button variant="link" className='p-s0' onClick={() => onCancel()}>
          ⟵ Back to Home
        </Button>
      </div>
      {isDisableCouponModalOpen && (
        <DisableCouponModal
          showModal={isDisableCouponModalOpen}
          triggerHideDisableCouponModal={triggerHideDisableCouponModal}
          coupon={coupon}
          handleDisableCoupon={handleDisableCoupon}
        />
      )}
      <div className="filter">
        <ExportDataToCSV csvHeaders={csvHeaders} data={allCouponsWithRedeemptionsCSV} />
        <SearchBox
          intc-id="searchCoupons"
          value={filterText}
          onChange={setFilter}
          placeholder="Search coupons..."
          aria-label="Type to search coupons..."
        />
      </div>
      <div className="section">
          <GridPagination data={gridItems} columns={couponsColumn} loading={loading} emptyGrid={emptyGrid} fixedFirstColumn/>
      </div>
      {redemptionsSection}

    </>
  )
}

export default CouponsView
