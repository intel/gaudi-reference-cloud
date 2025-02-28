// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import Modal from 'react-bootstrap/Modal'
import Button from 'react-bootstrap/Button'
import { Link } from 'react-router-dom'
import { BsExclamationOctagon } from 'react-icons/bs'
import useUserStore from '../../../store/userStore/UserStore'
import { appFeatureFlags, isFeatureFlagEnable } from '../../../config/configurator'

const UpgradeNeededModal = (props) => {
  const showModal = props.showModal
  const onClose = props.onClose
  const { isStandardUser, isPremiumUser, isEnterprisePendingUser, isIntelUser, isOwnCloudAccount } =
    useUserStore.getState()
  let messageToDisplay = null
  let messageOptions = null

  if (isStandardUser() && isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_UPGRADE_TO_PREMIUM)) {
    messageToDisplay = (
      <div className="section align-items-center text-center">
        <BsExclamationOctagon color="red" size="2.5em" />
        <h5>Account upgrade needed</h5>
        <div className="section align-items-center">
          <span>To launch this service you need to add a payment method.</span>
        </div>
      </div>
    )
    messageOptions = (
      <div className="mx-auto">
        <Button onClick={onClose} variant="link" className="text-decoration-none">
          Cancel
        </Button>
        <div className="space"></div>
        <Link to="/upgradeaccount">
          <Button variant="primary" className="text-decoration-none">
            Upgrade account
          </Button>
        </Link>
      </div>
    )
  }

  if (isPremiumUser()) {
    messageToDisplay = (
      <div className="section align-items-center text-center">
        <BsExclamationOctagon color="red" size="2.5em" />
        <h5>Invalid payment methods</h5>
        <div className="section align-items-center">
          <span>
            We&rsquo;re sorry, but we couldn&rsquo;t find a valid payment method on your account. Please add a valid
            payment method to continue with your transaction.
          </span>
          <span>If you believe this is an error, please contact our support team for further assistance.</span>
        </div>
      </div>
    )

    messageOptions = (
      <div className="mx-auto">
        <Button onClick={onClose} variant="link" className="text-decoration-none">
          Cancel
        </Button>
        <div className="space"></div>
        <Link to="/billing/managePaymentMethods">
          <Button variant="primary" className="text-decoration-none">
            View payment methods
          </Button>
        </Link>
      </div>
    )
  }

  if (isEnterprisePendingUser()) {
    messageToDisplay = (
      <div className="section align-items-center text-center">
        <BsExclamationOctagon color="red" size="2.5em" />
        <h5>Account confirmation required</h5>
        <div className="section align-items-center">
          <span>
            Your enterprise account needs to be confirmed before launching this service. While the confirmation is in
            progress you can only use free services.
          </span>
        </div>
      </div>
    )
    messageOptions = (
      <div className="mx-auto">
        <div className="space"></div>
        <Button onClick={onClose} variant="primary" className="text-decoration-none">
          Go back
        </Button>
      </div>
    )
  }

  if (
    isIntelUser() ||
    (isStandardUser() && !isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_UPGRADE_TO_PREMIUM))
  ) {
    messageToDisplay = (
      <div className="section align-items-center text-center">
        <BsExclamationOctagon color="red" size="2.5em" />
        <h5>Insufficient cloud credits</h5>
        <div className="section align-items-center">
          <span>
            We&rsquo;re sorry, but you don&rsquo;t have an outstanding credit balance to complete this transaction.
          </span>
          <span>Please ensure that you have sufficient credits or redeem a new coupon.</span>
          <span>
            If you believe this is an error, please contact our support team for further assistance. <br />
          </span>
        </div>
      </div>
    )

    messageOptions = (
      <div className="mx-auto">
        <Button onClick={onClose} variant="link" className="text-decoration-none">
          Cancel
        </Button>
        <div className="space"></div>
        <Link to="/billing/credits">
          <Button variant="primary" className="text-decoration-none">
            View cloud credits
          </Button>
        </Link>
      </div>
    )
  }

  if (!isOwnCloudAccount) {
    messageToDisplay = (
      <div className="section align-items-center text-center">
        <BsExclamationOctagon color="red" size="2.5em" />
        <h5>Invalid payment methods</h5>
        <div className="section align-items-center">
          <span>We&rsquo;re sorry. We couldn&rsquo;t find a valid payment method for your member account.</span>
          <span>If you believe this is an error, please contact the account owner for further assistance.</span>
        </div>
      </div>
    )

    messageOptions = (
      <div className="mx-auto">
        <Button onClick={onClose} variant="link" className="text-decoration-none">
          Cancel
        </Button>
      </div>
    )
  }

  return (
    <Modal
      show={showModal}
      backdrop="static"
      size="lg"
      aria-labelledby="contained-modal-title-vcenter"
      centered
      aria-label="Upgrade needed modal"
    >
      <Modal.Body>{messageToDisplay}</Modal.Body>
      <Modal.Footer>{messageOptions}</Modal.Footer>
    </Modal>
  )
}

export default UpgradeNeededModal
