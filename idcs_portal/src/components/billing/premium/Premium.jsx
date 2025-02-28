// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import ManagePaymentContainers from '../../../containers/billing/paymentMethods/ManagePaymentContainers'
import PaymentSkipModal from '../../../utils/modals/paymentSkipModal/PaymentSkipModal'
import TopToolbarContainer from '../../../containers/navigation/TopToolbarContainer'

const Premium = (props) => {
  // props Variables
  const titles = props.titles
  const cancelButtonOptions = props.cancelButtonOptions
  const formActions = props.formActions
  const isShowSkip = props.isShowSkip

  return (
    <>
      <TopToolbarContainer />
      <PaymentSkipModal showSkipModal={isShowSkip} skipAction={cancelButtonOptions.onClick}></PaymentSkipModal>
      <div className="section">
        <h1>{titles.pageTitle}</h1>
        <h2 intc-id="premiumTitle">{titles.mainTitle}</h2>
        <span>{titles.mainDesc}</span>
        <ManagePaymentContainers
          cancelButtonOptions={cancelButtonOptions}
          formActions={formActions}
        ></ManagePaymentContainers>
      </div>
    </>
  )
}

export default Premium
