// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Badge, DropdownButton, Dropdown, ButtonGroup } from 'react-bootstrap'
import { Link } from 'react-router-dom'
import { BsExclamationTriangle } from 'react-icons/bs'
import JupiterLaunchModal from '../trainingAndWorkshops/jupiterLaunchModal/JupiterLaunchModal'
import ErrorModal from '../../utils/modals/errorModal/ErrorModal'
import { type GetStartedCard, type GetStartedCardResource } from '../../containers/homePage/Home.types'
import { appFeatureFlags, isFeatureRegionBlocked } from '../../config/configurator'
import Spinner from '../../utils/spinner/Spinner'

interface GetStartedProps {
  tab: string
  errorModal: any
  launchModal: any
  getStartedDetails: GetStartedCard | undefined
  loading: boolean
  openJupiterLab: (route: string) => Promise<void>
  setErrorModal: (data: any) => void
  setLaunchModal: (data: any) => void
}

const GetStarted: React.FC<GetStartedProps> = (props): JSX.Element => {
  const errorModal = props.errorModal
  const launchModal = props.launchModal
  const getStartedDetails = props.getStartedDetails
  const loading = props.loading
  const openJupiterLab: any = props.openJupiterLab
  const setErrorModal = props.setErrorModal
  const setLaunchModal = props.setLaunchModal

  const spinner = <Spinner />

  // functions
  function getID(resource: GetStartedCardResource): string {
    const id = `btn-getStarted-${resource.label}-${resource?.actions?.label ?? ''}`
    return id.replace(' ', '')
  }

  function getResourcesAction(resource: GetStartedCardResource): JSX.Element {
    if (resource === undefined || resource.actions === undefined) {
      return <></>
    }
    let resourceActions = <></>
    if (resource.actions.type === 'dropdownButton') {
      resourceActions = (
        <>
          <DropdownButton
            as={ButtonGroup}
            variant={resource.actions.variant}
            title={resource.actions.label}
            intc-id={getID(resource)}
            data-wap_ref={getID(resource)}
            aria-label={`${resource.actions.label ?? ''}`}
            disabled={isFeatureRegionBlocked(appFeatureFlags.REACT_APP_FEATURE_TRAINING)}
          >
            {resource.actions.actions?.map((item, idx) => {
              return (
                <Dropdown.Item
                  key={idx}
                  intc-id={getID(resource) + `-${item.label.replace(' ', '')}`}
                  data-wap_ref={getID(resource) + `-${item.label.replace(' ', '')}`}
                  className="text-nowrap"
                  onClick={() => {
                    if (resource.actions?.redirectTo === 'jupyterLab') openJupiterLab(item.redirectTo)
                  }}
                >
                  {item.label}
                </Dropdown.Item>
              )
            })}
          </DropdownButton>
          {isFeatureRegionBlocked(appFeatureFlags.REACT_APP_FEATURE_TRAINING) && (
            <span className="d-flex flex-row align-items-center gap-s4">
              <BsExclamationTriangle />
              Unavailable in this region
            </span>
          )}
        </>
      )
    } else if (resource.actions.type === 'button') {
      const extraAttributes = resource.actions.openInNewTab
        ? {
            rel: 'noreferrer',
            target: '_blank'
          }
        : {}
      resourceActions = (
        <Link
          className={`btn btn-${resource.actions.variant} align-self-start mb-auto`}
          intc-id={getID(resource)}
          data-wap_ref={getID(resource)}
          target={resource.actions.openInNewTab ? 'blank' : undefined}
          aria-label={`${resource.actions.label ?? ''}`}
          to={resource?.actions.redirectTo}
          {...extraAttributes}
        >
          {resource.actions.leftIcon ? <resource.actions.leftIcon /> : <></>}
          {resource.actions.label} {resource.actions.rigthIcon ? <resource.actions.rigthIcon /> : <></>}
        </Link>
      )
    }
    if (resource.actions.badge && !isFeatureRegionBlocked(appFeatureFlags.REACT_APP_FEATURE_TRAINING)) {
      resourceActions = (
        <div className="d-flex flex-row gap-s4 align-items-center">
          {resourceActions}
          <Badge bg="primary" className="mb-0">
            {resource.actions.badge}
          </Badge>
        </div>
      )
    }
    return resourceActions
  }

  function getResourcesInfo(resources: GetStartedCardResource[] | undefined): JSX.Element {
    if (resources === undefined) {
      return <></>
    }
    return (
      <div className="row mx-sm-s6 mx-xs-s4 h-100">
        {resources.map((resource: GetStartedCardResource, index: number) => {
          const comlumnSize = Math.floor(12 / resources.length)
          return (
            <div
              className={`col-lg-${comlumnSize} col-sm-6 col-xs-12 d-flex flex-row ${!resource?.label ? 'd-xs-none' : ''}`}
              key={index}
            >
              {index > 0 && resource?.label && (
                <div className={`vr me-s8 d-xs-none ${index % 2 === 0 ? 'd-lg-block' : 'd-sm-block'}`}></div>
              )}
              <div className="d-flex flex-column gap-s5 h-100">
                {index > 0 && (
                  <div className="d-sm-none my-s4 d-xs-block">
                    <hr />
                  </div>
                )}
                {resource?.label ? <span className="h6">{resource.label}</span> : <></>}
                {resource?.text ? <span>{resource.text}</span> : <></>}
                {getResourcesAction(resource)}
                {getResourcesInfo(resource?.resources)}
              </div>
            </div>
          )
        })}
      </div>
    )
  }

  function getDetailsInfo(content: GetStartedCard | undefined): JSX.Element {
    if (content === undefined) {
      return <></>
    }
    return (
      <>
        <div className="d-flex flex-column gap-s6 h-100">
          <div className="d-flex flex-column gap-s3 align-items-baseline">
            <span className="text-muted h6">{content.subTitle}</span>
            {content.getStartedPageText && <span>{content.getStartedPageText}</span>}
          </div>
          {getResourcesInfo(content.resources)}
        </div>
      </>
    )
  }

  const getStartedContent = (
    <>
      <div className="section bd-highlight">
        <h2 className="d-flex flex-row gap-s4">
          {getStartedDetails?.title}
          {getStartedDetails?.badge && (
            <span className="d-flex align-items-center">
              <Badge bg="primary">{getStartedDetails?.badge}</Badge>
            </span>
          )}
        </h2>
      </div>
      <div className="section">{getDetailsInfo(getStartedDetails)}</div>
    </>
  )

  return (
    <>
      <JupiterLaunchModal
        jupyterRes={launchModal.response}
        show={launchModal.show}
        onClose={() => {
          setLaunchModal({ show: false, response: '' })
        }}
      />
      <ErrorModal
        showModal={errorModal.show}
        message={errorModal.message}
        titleMessage={'Could not launch notebook'}
        description={'There was an error while processing your request.'}
        onClickCloseErrorModal={() => {
          setErrorModal({ show: false, message: '' })
        }}
      />
      {loading ? spinner : getStartedContent}
    </>
  )
}
export default GetStarted
