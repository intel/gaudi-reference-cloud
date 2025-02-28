import { useEffect, useState } from 'react'

import useCouponsStore from '../../store/couponStore/CouponStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useToastStore from '../../store/toastStore/ToastStore'
import CouponsView from '../../components/coupons/CouponsView'
import { useNavigate } from 'react-router-dom'
import moment from 'moment/moment'
import CouponService from '../../services/CouponService'

const CouponsContainer = () => {
  const couponsColumn = [
    {
      columnName: 'Code',
      targetColumn: 'code'
    },
    {
      columnName: 'Creator',
      targetColumn: 'creator',
      className: 'text-break',
      width: '10rem'
    },
    {
      columnName: 'Created',
      targetColumn: 'created',
      width: '8rem'
    },
    {
      columnName: 'Start',
      targetColumn: 'start',
      width: '8rem'
    },
    {
      columnName: 'Expires',
      targetColumn: 'expires',
      width: '8rem'
    },
    {
      columnName: 'Disabled',
      targetColumn: 'disabled'
    },
    {
      columnName: 'Amount',
      targetColumn: 'amount'
    },
    {
      columnName: 'Uses',
      targetColumn: 'numberUses'
    },
    {
      columnName: 'Redeemed',
      targetColumn: 'numberRedeemed'
    },
    {
      columnName: 'Is Standard',
      targetColumn: 'isStandard'
    },
    {
      columnName: 'Redeemed',
      targetColumn: 'redeemedDetails',
      redeemedDetails: {
        behaviorType: 'hyperLink',
        behaviorFunction: 'setDetails'
      }
    },
    {
      columnName: 'Actions',
      targetColumn: 'disableCouponLink',
      redeemedDetails: {
        behaviorType: 'hyperLink',
        behaviorFunction: 'disableCouponLink'
      }
    }
  ]

  const redemptionsColumns = [
    {
      columnName: 'Code',
      targetColumn: 'code'
    },
    {
      columnName: 'CloudAccountId',
      targetColumn: 'cloudaccountid'
    },
    {
      columnName: 'Redeemed',
      targetColumn: 'redeemed'
    },
    {
      columnName: 'Applied',
      targetColumn: 'applied'
    }
  ]

  // Header to add to CSV file.
  const csvHeaders = [
    { label: 'Code', key: 'code' },
    { label: 'Creator', key: 'creator' },
    { label: 'Created', key: 'created' },
    { label: 'Start', key: 'start' },
    { label: 'Disabled', key: 'disabled' },
    { label: 'Amount', key: 'amount' },
    { label: 'No.of Uses', key: 'numUses' },
    { label: 'Redeemed Numbers', key: 'numRedeemed' },
    { label: 'Is Standard', key: 'isStandard' },
    { label: 'Cloud Account Ids', key: 'CloudAccountIds' }
  ]

  const emptyGrid = {
    title: 'Coupons details empty',
    subTitle: 'No details found',
    action: {
      type: 'redirect',
      href: '/cloudcredits/create',
      label: 'Create'
    }
  }

  const emptyGridByFilter = {
    title: 'No details found',
    subTitle: 'The applied filter criteria did not match any items',
    action: {
      type: 'function',
      href: () => setFilter('', true),
      label: 'Clear filters'
    }
  }

  const [couponsList, setCouponsList] = useState([])
  const [redemptionsDetails, setRedemptionsDetails] = useState(null)
  const [isDisableCouponModalOpen, setIsDisableCouponModalOpen] = useState(false)
  const [coupon, setCoupon] = useState(false)
  const [allCouponsWithRedeemptionsCSV, setAllCouponsWithRedeemptionsCSV] = useState([])
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)

  // Store.
  const loading = useCouponsStore((state) => state.loading)
  const coupons = useCouponsStore((state) => state.coupons)
  const setCoupons = useCouponsStore((state) => state.setCoupons)
  const setLoading = useCouponsStore((state) => state.setLoading)
  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)

  // Navigation
  const navigate = useNavigate()

  // Error Boundry
  const throwError = useErrorBoundary()

  // Hooks
  useEffect(() => {
    const fetch = async () => {
      try {
        await setCoupons()
      } catch (error) {
        const errorData = error?.response?.data
        if (errorData && errorData?.code) {
          setLoading(false)
          setEmptyGridObject(emptyGrid)
        } else {
          throwError(error)
        }
      }
    }
    fetch()
  }, [])

  useEffect(() => {
    setGridInfo()
  }, [coupons])

  const triggerFormatDate = (dateVal) => {
    return moment(dateVal).format('MM/DD/YYYY hh:mm:ss')
  }

  const setGridInfo = () => {
    const gridInfo = []
    const gridInfoCouponsWithRedeemptionsCSV = []

    for (const index in coupons) {
      const coupon = { ...coupons[index] }
      const redemptions = coupon?.redemptions

      const dataObj = {
        code: coupon?.code,
        creator: coupon?.creator,
        created: triggerFormatDate(coupon?.created),
        start: triggerFormatDate(coupon?.start),
        expires: triggerFormatDate(coupon?.expires),
        disabled: coupon?.disabled == null ? coupon?.disabled : 'Yes',
        amount: coupon?.amount,
        numUses: coupon?.numUses,
        numRedeemed: coupon?.numRedeemed,
        isStandard: coupon?.isStandard ? 'Yes' : 'No'
      }

      // UI Grid
      const uiDataObj = { ...dataObj }
      uiDataObj.redemptions =
        redemptions.length > 0
          ? {
              showField: true,
              type: 'HyperLink',
              value: 'Details',
              function: () => {
                setRedemptionsData(redemptions)
              }
            }
          : ''
      uiDataObj.disableCoupon =
        coupon?.disabled == null
          ? {
              showField: true,
              type: 'button',
              value: 'Disable',
              function: () => {
                setDisabledCouponData(coupon?.code)
              }
            }
          : ''
      gridInfo.push(uiDataObj)

      // CSV Grid
      const csvDataObj = { ...dataObj }
      csvDataObj.CloudAccountIds = ''

      if (redemptions.length > 0) {
        for (const rindex in redemptions) {
          const innerObj = { ...csvDataObj }
          innerObj.CloudAccountIds = redemptions[rindex]?.cloudAccountId
          gridInfoCouponsWithRedeemptionsCSV.push(innerObj)
        }
      } else {
        gridInfoCouponsWithRedeemptionsCSV.push(csvDataObj)
      }
    }
    setCouponsList(gridInfo)
    setAllCouponsWithRedeemptionsCSV(gridInfoCouponsWithRedeemptionsCSV)
  }

  // Performs action when we click on Detials link from Coupons list page.
  const setDisabledCouponData = (coupon) => {
    setIsDisableCouponModalOpen(true)
    // Sets the specific coupon code to disable.
    setCoupon(coupon)
  }

  // Hides Disable coupon modal.
  const triggerHideDisableCouponModal = () => {
    setIsDisableCouponModalOpen(false)
  }

  const handleDisableCoupon = async (coupon) => {
    const payload = { code: coupon }
    setLoading(true)
    setIsDisableCouponModalOpen(false)

    try {
      await CouponService.disableCoupon(payload)
      showSuccess(`Successfully disabled coupon '${coupon}'`)
      await setCoupons()
    } catch (error) {
      let errorMessage = ''
      if (error.response) {
        errorMessage = error.message
      }
      showError(errorMessage)
    }
    setLoading(false)
  }

  // Sets redemptions details for specific coupon.
  const setRedemptionsData = (redemptionsData) => {
    setRedemptionsDetails(redemptionsData)
    scrollToTheBottom()
  }

  // Scrolls the screen to the bottom of the page.
  const scrollToTheBottom = () => {
    window.scrollTo({
      top: 1000,
      behavior: 'smooth'
    })
  }

  // Method to redirect back to homepage.
  const onCancel = () => {
    navigate('/')
  }

  function setFilter(event, clear) {
    if (clear) {
      setEmptyGridObject(emptyGrid)
      setFilterText('')
    } else {
      setEmptyGridObject(emptyGridByFilter)
      setFilterText(event.target.value)
    }
  }

  return (
    <CouponsView
      emptyGrid={emptyGridObject}
      filterText={filterText}
      loading={loading}
      coupons={couponsList}
      redemptionsDetails={redemptionsDetails}
      allCouponsWithRedeemptionsCSV={allCouponsWithRedeemptionsCSV}
      coupon={coupon}
      couponsColumn={couponsColumn}
      redemptionsColumns={redemptionsColumns}
      csvHeaders={csvHeaders}
      isDisableCouponModalOpen={isDisableCouponModalOpen}
      setFilter={setFilter}
      onCancel={onCancel}
      triggerHideDisableCouponModal={triggerHideDisableCouponModal}
      handleDisableCoupon={handleDisableCoupon}
    />
  )
}

export default CouponsContainer
