// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import EmptyView from '../../../utils/emptyView/EmptyView'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import JupiterLaunchModal from '../../trainingAndWorkshops/jupiterLaunchModal/JupiterLaunchModal'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'
import ComingSoonBanner from '../../../utils/comingSoonBanner/ComingSoonBanner'
import './SoftwareDetail.scss'
import TabsNavigation from '../../../utils/tabsNavigation/TabsNavagation'
import TapContent from '../../../utils/TapContent/TapContent'
import MediaCarousel from '../../../utils/mediaCarousel/MediaCarousel'
import Spinner from '../../../utils/spinner/Spinner'
import { Link } from 'react-router-dom'
import { Button } from 'react-bootstrap'

const SoftwareDetailThirdParty = (props) => {
  const loading = props.loading
  const activeTab = props.activeTab
  const setActiveTab = props.setActiveTab
  const noFoundSoftware = props.noFoundSoftware
  const software = props.softwareDetail
  const launchModal = props.launchModal
  const setLaunchModal = props.setLaunchModal
  const jupyterRes = props.jupyterRes
  const errorModal = props.errorModal
  const setErrorModal = props.setErrorModal
  const comingMessage = props.comingMessage
  const isAvailable = props.isAvailable
  const softwareLogo = props.softwareLogo
  const softwareLogoAsTitle = props.softwareLogoAsTitle
  const onClickLaunch = props.onClickLaunch

  const tabs = [
    {
      label: 'Overview',
      id: 'overview'
    }
  ]
  let content = null

  function getTabOverview() {
    return (
      <>
        <div className="section gap-s5 px-0">
          {software.overview && (
            <ReactMarkdown linkTarget="_blank" remarkPlugins={[remarkGfm]} components={{ p: 'span' }}>
              {software.overview}
            </ReactMarkdown>
          )}
          <MediaCarousel mediaList={software.mediaArray} />
          {software.features ? (
            <div className="d-flex flex-column text gap-s5">
              <ReactMarkdown linkTarget="_blank" remarkPlugins={[remarkGfm]} components={{ p: 'span' }}>
                {software.features}
              </ReactMarkdown>
            </div>
          ) : (
            ''
          )}
        </div>
      </>
    )
  }

  function getTabPricing() {
    return <></>
  }

  function getTapInfo(activeTap) {
    const tabContent = {
      tapConfig: { type: 'custom' },
      customContent: null
    }
    switch (activeTap) {
      case 0:
        tabContent.hiddenTitle = 'Overview'
        tabContent.customContent = getTabOverview()
        break
      case 1:
        tabContent.hiddenTitle = 'Pricing'
        tabContent.customContent = getTabPricing()
        break
    }
    return tabContent
  }

  if (!isAvailable) {
    return <ComingSoonBanner message={comingMessage} />
  } else {
    if (loading) {
      content = <Spinner />
    } else {
      if (software) {
        content = (
          <>
            <div className="section">
              <div className="d-flex flex-row align-items-center gap-s4">
                {softwareLogoAsTitle ? (
                  <h2 intc-id={`h1-title-${software.displayName}`} aria-label={software.displayName}>
                    {softwareLogo}
                  </h2>
                ) : (
                  <>
                    {softwareLogo}
                    <h2 intc-id={`h1-title-${software.displayName}`}>{software.displayName}</h2>
                  </>
                )}
                <span>{software.displayCatalogDesc}</span>
              </div>
            </div>
            <div className="section flex-row flex-wrap">
              {software.launchLink && software.launch !== 'internalLink' && (
                <a
                  intc-id="btn-third-party-get-it-now"
                  className="btn btn-primary"
                  href={software.launchLink}
                  rel="noreferrer"
                  target="_blank"
                >
                  Get it now
                </a>
              )}

              {software.demoURL && (
                <a
                  intc-id="btn-third-party-book-a-demo"
                  className="btn btn-outline-primary"
                  href={software.demoURL}
                  rel="noreferrer"
                  target="_blank"
                >
                  Book a demo
                </a>
              )}
              {software.helpURL && (
                <a
                  intc-id="btn-third-party-support"
                  className="link"
                  href={software.helpURL}
                  rel="noreferrer"
                  target="_blank"
                >
                  Support
                </a>
              )}
              {software.licenseURL && (
                <a
                  intc-id="btn-third-party-license-agreement"
                  className="link"
                  href={software.licenseURL}
                  rel="noreferrer"
                  target="_blank"
                >
                  License agreement
                </a>
              )}
              {software.launch === 'internalLink' && software.launchLink && (
                <Link
                  intc-id="btn-third-party-try-it-now"
                  data-wap_ref="btn-third-party-try-it-now"
                  className="btn btn-primary"
                  to={`/software/d/${software.id}/llm`}
                  target={software.launchLink.startsWith('/') ? undefined : '_blank'}
                  rel={software.launchLink.startsWith('/') ? undefined : 'noopener noreferrer'}
                  aria-label="Try it now"
                >
                  Try it now
                </Link>
              )}
              {software.launch === 'jupyterlab' && (
                <Button
                  intc-id="btn-third-party-launch-software-jupiter"
                  variant="primary"
                  aria-label="Launch Jupyter Notebook"
                  data-wap_ref="btn-details-launch-software-jupiter"
                  onClick={(e) => onClickLaunch('jupyter')}
                >
                  Launch Jupyter Notebook
                </Button>
              )}
              {software.launch === 'image' && (
                <Button
                  intc-id="btn-third-party-launch-software"
                  variant="primary"
                  aria-label="Launch"
                  data-wap_ref="btn-details-launch-software"
                  onClick={() => onClickLaunch('image')}
                >
                  Launch
                </Button>
              )}
            </div>
            <div className="section">
              <TabsNavigation tabs={tabs} activeTab={activeTab} setTabActive={setActiveTab} />
              <TapContent infoToDisplay={getTapInfo(activeTab)} />
            </div>
          </>
        )
      } else {
        content = (
          <EmptyView
            title={noFoundSoftware.title}
            subTitle={noFoundSoftware.subTitle}
            action={noFoundSoftware.action}
          />
        )
      }
    }
  }

  return (
    <>
      <JupiterLaunchModal show={launchModal} type={'software'} jupyterRes={jupyterRes} onClose={setLaunchModal} />
      <ErrorModal
        showModal={errorModal.show}
        titleMessage={'Could not launch software'}
        description={'We are sorry for the inconvenience, but we are having an unexpected technical issue right now.'}
        onClickCloseErrorModal={() => setErrorModal({ show: false, message: '' })}
      />
      {content}
    </>
  )
}

export default SoftwareDetailThirdParty
