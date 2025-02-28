// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import CouponCodeContainers from '../../../containers/billing/paymentMethods/CouponCodeContainers'

const ManageCouponCode = (props) => {
  // props Variables
  const titles = props.titles
  const cancelButtonOptions = props.cancelButtonOptions
  const formActions = props.formActions

  return (
    <>
      <div className="section">
        <h2>{titles.pageTitle}</h2>
      </div>
      <div className="section">
        <CouponCodeContainers
          cancelButtonOptions={cancelButtonOptions}
          formActions={formActions}
        ></CouponCodeContainers>
      </div>
    </>
  )
}

export default ManageCouponCode
