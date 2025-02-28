// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { memo, useRef, useState, useEffect, useCallback } from 'react'
import idcConfig from '../../config/configurator'
import { useLocation } from 'react-router-dom'

interface FooterMiniProps {
  validateRoute?: boolean
}

const FooterMini: React.FC<FooterMiniProps> = ({ validateRoute = false }): JSX.Element => {
  const refFooter = useRef<HTMLDivElement>(null)
  const [height, setHeight] = useState('0px')
  const location = validateRoute ? useLocation() : window.location
  const [showFooter, setShowFooter] = useState(true)
  const allowedPaths = ['/']

  const shouldShowFooter = (): void => {
    allowedPaths.includes(location.pathname) ? setShowFooter(true) : setShowFooter(false)
  }

  const updateHeighOfHiddenDiv = useCallback(() => {
    if (refFooter.current) {
      setHeight(String(refFooter.current.clientHeight) + 'px')
    }
  }, [])

  useEffect(() => {
    setTimeout(updateHeighOfHiddenDiv, 50)
    window.addEventListener('resize', updateHeighOfHiddenDiv)
    return () => {
      window.removeEventListener('resize', updateHeighOfHiddenDiv)
    }
  }, [])

  useEffect(() => {
    shouldShowFooter()
  }, [location])

  return showFooter
    ? (
    <>
      <div style={{ clear: 'both', height: `${height}` }}></div>
      <div
        className="footer fixed-bottom w-100"
        style={{ left: '0', right: 'auto' }}
        ref={refFooter}
        aria-label="Footer"
      >
        <a className="d-flex align-items-flex-start mr-auto" href="https://www.intel.com/">
          {idcConfig.REACT_APP_COMPANY_LONG_NAME}
        </a>

        <div className="additional-links">
          <a className="footer-link btn" href="https://www.intel.com/content/www/us/en/legal/terms-of-use.html">
            Terms of Use
          </a>
          <a className="footer-link btn" href="https://www.intel.com/content/www/us/en/legal/trademarks.html">
            *Trademarks
          </a>
          <a
            className="footer-link btn"
            href="https://www.intel.com/content/www/us/en/privacy/intel-cookie-notice.html"
            data-cookie-notice="true"
          >
            Cookies
          </a>
          <a
            className="footer-link btn"
            href="https://www.intel.com/content/www/us/en/privacy/intel-privacy-notice.html"
            data-cookie-notice="true"
          >
            Privacy
          </a>
          <a
            className="footer-link btn"
            href="https://www.intel.com/content/www/us/en/corporate-responsibility/statement-combating-modern-slavery.html"
          >
            Supply Chain Transparency
          </a>
          <a className="footer-link btn" href="https://www.intel.com/content/www/us/en/siteindex.html">
            Site Map
          </a>
          <a className="footer-link btn" href="/#" data-wap_ref="dns" id="wap_dns">
            Your Privacy Choices
          </a>
          <a
            className="footer-link btn"
            href="https://www.intel.com/content/www/us/en/privacy/privacy-residents-certain-states.html"
            data-wap_ref="nac"
            id="wap_nac"
          >
            Notice at Collection
          </a>
        </div>
      </div>
    </>
      )
    : (
    <></>
      )
}

export default memo(FooterMini)
