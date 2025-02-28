// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import CreditCardContainers from '../../../containers/billing/paymentMethods/CreditCardContainers'

const ManageCreditCard = (props) => {
  // props Variables
  const titles = props.titles
  const cancelButtonOptions = props.cancelButtonOptions
  const formActions = props.formActions

  return (
    <>
      <div className="section">
        <h2 intc-id="manageCreditCardTitle">{titles.pageTitle}</h2>
      </div>
      <div className="section">
        <CreditCardContainers
          formActions={formActions}
          cancelButtonOptions={cancelButtonOptions}
        ></CreditCardContainers>
      </div>
    </>
  )
}

export default ManageCreditCard
