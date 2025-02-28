// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import Modal from 'react-bootstrap/Modal'
import ManagePaymentContainers from '../../../containers/billing/paymentMethods/ManagePaymentContainers'
import CustomAlerts from '../../customAlerts/CustomAlerts'

const ModalUpgradeAccount = (props) => {
  // props Variables
  const showPremiumModal = props.showPremiumModal
  const cancelButtonOptions = props.cancelButtonOptions
  const formActions = props.formActions
  const error = props.error

  return (
    <Modal
      show={showPremiumModal}
      onHide={cancelButtonOptions.onClick}
      backdrop="static"
      keyboard={false}
      size="xl"
      aria-label="Upgrade modal"
    >
      <Modal.Header closeButton>
        <Modal.Title>Upgrade your account to continue</Modal.Title>
      </Modal.Header>
      <Modal.Body>
        <div className="section-component">
          <div className="p-3">
            <CustomAlerts
              showAlert={error.isShow}
              alertType="error"
              strongText="Error"
              message={error.errorMessage}
              onCloseAlert={() => formActions.afterError('', false)}
              showIcon={true}
            />
            <p>Provide a payment method for billing purposes </p>
            <ManagePaymentContainers
              cancelButtonOptions={cancelButtonOptions}
              formActions={formActions}
            ></ManagePaymentContainers>
          </div>
        </div>
      </Modal.Body>
    </Modal>
  )
}

export default ModalUpgradeAccount
