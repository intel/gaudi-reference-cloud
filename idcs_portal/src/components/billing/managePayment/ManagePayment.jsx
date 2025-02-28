// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { ButtonGroup, Button } from 'react-bootstrap'
import CouponCodeContainers from '../../../containers/billing/paymentMethods/CouponCodeContainers'
import CreditCardContainers from '../../../containers/billing/paymentMethods/CreditCardContainers'
import { CreditCard, TicketFill } from 'react-bootstrap-icons'
import { appFeatureFlags, isFeatureFlagEnable } from '../../../config/configurator'

const ManagePayment = (props) => {
  // props Variables
  const showFormOnButton = props.showFormOnButton
  const cancelButtonOptions = props.cancelButtonOptions
  const formActions = props.formActions

  // props functions
  const actionOnButtonSelection = props.actionOnButtonSelection

  return (
    <>
      <ButtonGroup size="lg">
        {isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_DIRECTPOST) ? (
          <Button
            intc-id="btnSelectCreditPayment"
            variant={showFormOnButton.creditCardVariant}
            onClick={() => actionOnButtonSelection('credit')}
          >
            <CreditCard></CreditCard> Credit Card
          </Button>
        ) : null}
        {isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_UPGRADE_PREMIUM_COUPON_CODE) ? (
          <Button
            intc-id="btnSelectCouponPayment"
            variant={showFormOnButton.couponVariant}
            onClick={() => actionOnButtonSelection('coupon')}
          >
            <TicketFill></TicketFill> Coupon code
          </Button>
        ) : null}
      </ButtonGroup>

      {showFormOnButton.isShow === 'couponCode' && (
        <CouponCodeContainers
          cancelButtonOptions={cancelButtonOptions}
          formActions={formActions}
        ></CouponCodeContainers>
      )}

      {showFormOnButton.isShow === 'creditCard' && (
        <CreditCardContainers
          cancelButtonOptions={cancelButtonOptions}
          formActions={formActions}
        ></CreditCardContainers>
      )}
    </>
  )
}

export default ManagePayment
