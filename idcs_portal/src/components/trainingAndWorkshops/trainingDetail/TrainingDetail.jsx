// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Button from 'react-bootstrap/Button'
import EmptyView from '../../../utils/emptyView/EmptyView'
import LaunchTrainingModal from '../launchTrainingModal/LaunchTrainingModal'
import JupiterLaunchModal from '../jupiterLaunchModal/JupiterLaunchModal'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import './TrainingDetail.scss'
import LineDivider from '../../../utils/lineDivider/LineDivider'
import Spinner from '../../../utils/spinner/Spinner'

const TrainingDetail = (props) => {
  const training = props.training
  const loading = props.loading
  const launchModal = props.launchModal
  const onClickLaunch = props.onClickLaunch
  const noFoundTraining = props.noFoundTraining
  const showErrorModal = props.showErrorModal
  const setShowErrorModal = props.setShowErrorModal
  const enrolling = props.enrolling
  const serviceResponse = props.serviceResponse
  const onClickLaunchJupiter = props.onClickLaunchJupiter
  const jupiterLaunchModal = props.jupiterLaunchModal
  const jupiterExpiry = props.jupiterExpiry
  const jupyterRes = props.jupyterRes
  const closeJupyterModal = props.closeJupyterModal
  let content = null

  if (loading) {
    content = <Spinner />
  } else {
    if (training) {
      content = (
        <>
          <div className="section">
            <h2 intc-id={`h1-title-${training.displayName}`}>{training.displayName}</h2>
            <div className="d-flex d-md-none">
              <Button
                intc-id="btn-details-launch-training-jupiter"
                data-wap_ref="btn-details-launch-training-jupiter"
                variant="primary"
                onClick={(e) => onClickLaunchJupiter(e, true)}
              >
                Launch Jupyter notebook
              </Button>
            </div>
          </div>
          <div className="section flex-xs-column flex-md-row gap-s8">
            <div className="TrainingDetails text gap-s8">
              {training.overview ? (
                <div className="d-flex flex-column gap-s6">
                  <h3>Overview</h3>
                  <ReactMarkdown linkTarget="_blank" remarkPlugins={[remarkGfm]} components={{ p: 'span' }}>
                    {training.overview}
                  </ReactMarkdown>
                </div>
              ) : (
                ''
              )}
              {training.audience ? (
                <div className="d-flex flex-column gap-s6">
                  <h3>Audience</h3>
                  <ReactMarkdown linkTarget="_blank" remarkPlugins={[remarkGfm]} components={{ p: 'span' }}>
                    {training.audience}
                  </ReactMarkdown>
                </div>
              ) : (
                ''
              )}
              {training.expectations ? (
                <div className="d-flex flex-column gap-s6">
                  <h3>Learning objectives</h3>
                  <ReactMarkdown linkTarget="_blank" remarkPlugins={[remarkGfm]} components={{ p: 'span' }}>
                    {training.expectations}
                  </ReactMarkdown>
                </div>
              ) : (
                ''
              )}
              {training.gettingStarted ? (
                <div className="d-flex flex-column gap-s6">
                  <h3>Getting started</h3>
                  <ReactMarkdown linkTarget="_blank" remarkPlugins={[remarkGfm]} components={{ p: 'span' }}>
                    {training.gettingStarted}
                  </ReactMarkdown>
                </div>
              ) : (
                ''
              )}
              {training.prerrequisites ? (
                <div className="d-flex flex-column gap-s6">
                  <h3>Prerequisites</h3>
                  <ReactMarkdown linkTarget="_blank" remarkPlugins={[remarkGfm]} components={{ p: 'span' }}>
                    {training.prerrequisites}
                  </ReactMarkdown>
                </div>
              ) : (
                ''
              )}
            </div>
            <LineDivider className="d-xs-none d-md-flex" vertical />
            <div className="TrainingDetailsActions gap-s6">
              <div className="d-md-flex d-xs-none">
                <Button
                  intc-id="btn-details-launch-training-jupiter"
                  data-wap_ref="btn-details-launch-training-jupiter"
                  variant="primary"
                  onClick={(e) => onClickLaunchJupiter(e, true)}
                >
                  Launch Jupyter notebook
                </Button>
              </div>
              {jupiterExpiry !== 'Invalid date' && (
                <div className="d-flex flex-column gap-s4">
                  <h4 className="fw-semibold">Considerations</h4>
                  <span>Your Jupyter notebook access will be live until {jupiterExpiry}.</span>
                </div>
              )}
              {training.featuredSoftware ? (
                <div className="d-flex flex-column gap-s6 text">
                  <h4 className="fw-semibold">Featured software</h4>
                  <ReactMarkdown linkTarget="_blank" remarkPlugins={[remarkGfm]} components={{ p: 'span' }}>
                    {training.featuredSoftware}
                  </ReactMarkdown>
                </div>
              ) : (
                ''
              )}
            </div>
          </div>
        </>
      )
    } else {
      content = (
        <EmptyView title={noFoundTraining.title} subTitle={noFoundTraining.subTitle} action={noFoundTraining.action} />
      )
    }
  }

  return (
    <>
      <LaunchTrainingModal
        show={launchModal}
        onClose={onClickLaunch}
        serviceResponse={serviceResponse}
        loading={enrolling}
      />
      <JupiterLaunchModal show={jupiterLaunchModal} jupyterRes={jupyterRes} onClose={closeJupyterModal} />
      <ErrorModal
        showModal={showErrorModal}
        titleMessage={'Could not launch notebook'}
        description={'We are sorry for the inconvenience, but we are having an unexpected technical issue right now.'}
        onClickCloseErrorModal={() => setShowErrorModal(false)}
      />
      {content}
    </>
  )
}

export default TrainingDetail
