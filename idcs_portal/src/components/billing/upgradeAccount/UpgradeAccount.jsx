// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import ManagePaymentContainers from '../../../containers/billing/paymentMethods/ManagePaymentContainers'

const UpgradeAccount = (props) => {
  // props Variables
  const titles = props.titles
  const cancelButtonOptions = props.cancelButtonOptions
  const formActions = props.formActions

  return (
    <>
      <div className="section">
        <h2 intc-id="UpgradeAccountTitle">{titles.pageTitle}</h2>
        <p className="lead">{titles.pageDesc}</p>
      </div>
      <div className="section">
        <h3>{titles.mainTitle}</h3>
        <span>{titles.mainDesc}</span>
        <ManagePaymentContainers
          cancelButtonOptions={cancelButtonOptions}
          formActions={formActions}
        ></ManagePaymentContainers>
      </div>
    </>
  )
}

export default UpgradeAccount
