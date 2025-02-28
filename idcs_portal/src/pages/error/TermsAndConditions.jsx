// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import idcConfig from '../../config/configurator'
import Button from 'react-bootstrap/Button'
import useUserStore from '../../store/userStore/UserStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import { useState } from 'react'
import Spinner from '../../utils/spinner/Spinner'

const TermsAndConditions = ({ acceptCallback }) => {
  const enroll = useUserStore((state) => state.enroll)
  const setcloudAccounts = useUserStore((state) => state.setcloudAccounts)
  const [loading, setLoading] = useState(false)
  const throwError = useErrorBoundary()

  const acceptTermsAndConditions = async () => {
    try {
      setLoading(true)
      await enroll(false, true)
      await setcloudAccounts()
      setTimeout(() => {
        setLoading(false)
        if (acceptCallback) {
          acceptCallback()
        }
      }, 250)
    } catch (error) {
      setLoading(false)
      throwError(error)
    }
  }

  return (
    <>
      <div className="section">
        <h1>Terms and conditions</h1>
      </div>
      <div className="section">
        <h2 className="align-self-center text-center">
          {`${idcConfig.REACT_APP_CONSOLE_LONG_NAME}`} Terms and Conditions
        </h2>
        <p>
          Thank you for your interest in {idcConfig.REACT_APP_CONSOLE_LONG_NAME} articulated in the Order Form (the
          &apos;Services&apos;).&nbsp;{`${idcConfig.REACT_APP_COMPANY_SHORT_NAME}`}’s offer and delivery of the Services
          to You are subject to the terms and conditions articulated in this Order Form,&nbsp;
          <a target="_blank" href={idcConfig.REACT_APP_SERVICE_AGREEMENT_URL} rel="noreferrer">
            {`${idcConfig.REACT_APP_COMPANY_SHORT_NAME}’s ${idcConfig.REACT_APP_CONSOLE_SHORT_NAME} Service Agreement`}
          </a>
          &nbsp;and&nbsp;
          <a target="_blank" href={idcConfig.REACT_APP_SOFTWARE_AGREEMENT_URL} rel="noreferrer">
            {`${idcConfig.REACT_APP_COMPANY_SHORT_NAME}’s Standard Commercial Software and Services Terms and Conditions`}
          </a>
          &nbsp;(collectively, the &apos;Legal Terms and Conditions&apos;).
          <br />
          <br />
          By clicking &apos;I Accept&apos;, below, You acknowledge that You have reviewed, understand, and accept the
          Legal Terms and Conditions as a prerequisite to and condition of Your access and use of the Services. Please
          do not access or use the Services unless and until you agree to the Legal Terms and Conditions.
        </p>
      </div>
      <div className="section align-items-center justify-content-center">
        {loading ? (
          <Spinner />
        ) : (
          <Button
            variant="primary"
            intc-id={'btn-terms-and-conditions-accept'}
            data-wap_ref={'btn-terms-and-conditions-accept'}
            aria-label="Accept term and conditions"
            onClick={() => acceptTermsAndConditions()}
          >
            I accept
          </Button>
        )}
      </div>
    </>
  )
}

export default TermsAndConditions
