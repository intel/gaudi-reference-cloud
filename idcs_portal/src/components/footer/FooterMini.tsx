// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { memo, useRef, useState, useEffect, useCallback } from 'react'
import idcConfig, { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'
import { useMediaQuery } from 'react-responsive'
import useAppStore from '../../store/appStore/AppStore'
import './FooterMini.scss'

declare const window: any

const FooterMini: React.FC = (): JSX.Element => {
  const showSideNavBar = useAppStore((state) => state.showSideNavBar)
  const showLearningBar = useAppStore((state) => state.showLearningBar)
  const learningArticlesAvailable = useAppStore((state) => state.learningArticlesAvailable)
  const refFooter = useRef<HTMLDivElement>(null)
  const [height, setHeight] = useState('0px')
  const [hasScrollBar, setHasScrollBar] = useState(false)

  const isSmScreen = useMediaQuery({
    query: '(max-width: 767px)'
  })

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
    updateHeighOfHiddenDiv()
  }, [showSideNavBar, showLearningBar, learningArticlesAvailable])

  useEffect(() => {
    try {
      const rootElement = document.getElementById('root')
      const resizeObserver = new ResizeObserver(() => {
        const scrollHeight = rootElement?.scrollHeight ?? 0
        const clientHeight = rootElement?.clientHeight ?? 0
        const hasScrollBar = scrollHeight > clientHeight
        setHasScrollBar(hasScrollBar)
      })
      resizeObserver.observe(rootElement as Element)
      return () => {
        resizeObserver?.disconnect()
      }
    } catch (error) {
      // No Support for resizeObserver
    }
  }, [])

  useEffect(() => {
    if (window?.wap_tms && window?.wap_tms.consent && window?.wap_tms.consent.enablePrivacyLinks) {
      window?.wap_tms.consent.enablePrivacyLinks()
    }
  }, [])

  return (
    <>
      <div style={{ clear: 'both', height: `${height}` }}></div>
      <div
        className={`footer ${isSmScreen || hasScrollBar ? `bottom  ${showLearningBar && learningArticlesAvailable ? 'learningBarMargingEnd' : ''} ${showSideNavBar ? 'sideNavBarMargingStart' : ''}` : 'fixed-bottom'} w-100`}
        style={{ left: '0', right: 'auto' }}
        ref={refFooter}
        aria-label="Footer"
      >
        <a className="d-flex align-items-flex-start mr-auto" href="https://www.intel.com/">
          {idcConfig.REACT_APP_COMPANY_LONG_NAME}
        </a>

        <div className="additional-links">
          {isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_FEEDBACK) && (
            <a
              target="_blank"
              className="footer-link btn"
              href={idcConfig.REACT_APP_PUBLIC_FEEDBACK_URL}
              rel="noreferrer"
            >
              Send Feedback
            </a>
          )}
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
            onClick={(e) => {
              e.preventDefault()
              if (window?.wap_tms && window?.wap_tms.consent && window?.wap_tms.consent.openConsentMenu) {
                window?.wap_tms.consent.openConsentMenu()
              }
            }}
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
          <a
            className="footer-link btn"
            href="/#"
            data-wap_ref="dns"
            id="wap_dns"
            onClick={(e) => {
              e.preventDefault()
              if (window?.wap_tms && window?.wap_tms.consent && window?.wap_tms.consent.openConsentMenu) {
                window?.wap_tms.consent.openConsentMenu()
              }
            }}
          >
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
}

export default memo(FooterMini)
