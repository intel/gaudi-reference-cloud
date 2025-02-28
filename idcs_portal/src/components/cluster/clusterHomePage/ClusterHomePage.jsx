// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { NavLink } from 'react-router-dom'
import idcConfig from '../../../config/configurator'
import './ClusterHomePage.scss'
import { Badge } from 'react-bootstrap'

const ClusterHomeStep = ({ index, step }) => {
  return (
    <div className={`ClusterStepItem p-s6 gap-md-s8 gap-xs-s6 ${index % 2 === 1 ? 'defaultBackground' : ''}`}>
      <>
        <div className="d-xs-none d-md-block text-center">
          <h3 className="d-flex flex-column h4 gap-s4">
            <div>
              <Badge bg="primary" className="rounded-circle">
                {index + 1}
              </Badge>
            </div>
            <span className="strong">{step.title}</span>
          </h3>
        </div>
        <div className="d-none d-xs-block d-md-none">
          <div className="d-flex flex-row gap-s6">
            <div className="d-block">
              <h3>
                <Badge bg="primary" className="rounded-circle">
                  {index + 1}
                </Badge>
              </h3>
            </div>
            <div className="d-block text-start">
              <span className="strong mx-1">{step.title}</span>
              <span className="small">{step.description}</span>
            </div>
          </div>
        </div>
      </>
      <div className="text-center d-xs-none d-md-block">
        <p>{step.description}</p>
      </div>
    </div>
  )
}

const ClusterHomePage = (props) => {
  const state = props.state
  const steps = state.steps

  return (
    <>
      <div className="filter">
        <NavLink to="/cluster/reserve" className="btn btn-primary">
          Launch Cluster
        </NavLink>
      </div>
      <div className="section">
        <h2>
          {`Run your Kubernetes workloads at scale using ${idcConfig.REACT_APP_COMPANY_SHORT_NAME}’s GPU and AI accelerators`}
        </h2>
        <div className="row">
          <div className="col-xs-12 col-xl-6">
            <div className="d-flex flex-xs-column flex-md-row">
              {steps.slice(0, 2).map((step, index) => (
                <ClusterHomeStep key={index} index={index} step={step}></ClusterHomeStep>
              ))}
            </div>
          </div>
          <div className="col-xs-12 col-xl-6">
            <div className="d-flex flex-xs-column flex-md-row">
              {steps.slice(-2).map((step, index) => (
                <ClusterHomeStep key={index} index={index + 2} step={step}></ClusterHomeStep>
              ))}
            </div>
          </div>
        </div>
      </div>
    </>
  )
}

export default ClusterHomePage
