// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { BsQuestionCircle } from 'react-icons/bs'
import UpgradeVersionModal from '../../cluster/clusterMyReservations/UpgradeVersionModal'
import { Button } from 'react-bootstrap'
import LabelValuePair from '../../../utils/labelValuePair/LabelValuePair'

const SuperComputerInfo = ({
  displayInfo,
  clusterDetail,
  upgradeModal,
  upgradeForm,
  onChangeInput,
  submitUpgradeK8sVersion
}) => {
  const isValidUpgradeForm = upgradeForm.isValidForm
  const columnSize = 12 / 3
  // variables
  const formElementskeys = []

  for (const key in upgradeForm.form) {
    const formItem = {
      ...upgradeForm.form[key]
    }

    formElementskeys.push({
      id: key,
      configInput: formItem
    })
  }

  function generateActionsButtons(field, actions, item) {
    let content = <></>
    if (clusterDetail?.clusterstate === 'Active') {
      content = actions.map((action, actionIndx) => {
        let returnValue = null
        if (field === 'kubeconfig') {
          returnValue = (
            <Button
              variant="link"
              size="sm"
              key={actionIndx}
              onClick={() => action.func(clusterDetail)}
              intc-id={`btn-details-tab-${field}-${action.label}`}
              data-wap_ref={`btn-details-tab-${field}-${action.label}`}
            >
              {action.name}
            </Button>
          )
        } else if (field === 'k8sversion') {
          returnValue = clusterDetail?.upgradeavailable ? (
            <Button
              variant="link"
              size="sm"
              key={actionIndx}
              onClick={action.func}
              intc-id={`btn-details-tab-${field}-${action.label}`}
              data-wap_ref={`btn-details-tab-${field}-${action.label}`}
            >
              <BsQuestionCircle />
              {action.label}
            </Button>
          ) : (
            <span className="small" key={actionIndx}>
              You are already running the latest version of Kubernetes
            </span>
          )
        }
        return returnValue
      })
    }

    return content
  }

  return (
    <>
      <UpgradeVersionModal
        show={upgradeModal.show}
        centered={upgradeModal.centered}
        closeButton={upgradeModal.closeButton}
        onHide={upgradeModal.onHide}
        onChangeUpgradeForm={onChangeInput}
        formElementskeys={formElementskeys}
        isValidUpgradeForm={isValidUpgradeForm}
        submitUpgradeK8sVersion={submitUpgradeK8sVersion}
      />
      <div className="section">
        <div className="row">
          {displayInfo.map((item, index) => (
            <LabelValuePair className={`col-md-${columnSize}`} key={index} label={item.label}>
              <div
                className={`d-flex ${item.field === 'k8sversion' && !clusterDetail?.upgradeavailable ? 'flex-column' : 'align-items-center gap-s4'} text-wrap`}
              >
                {item.value} {item.actions ? generateActionsButtons(item.field, item.actions, item) : null}
              </div>
            </LabelValuePair>
          ))}
        </div>
      </div>
    </>
  )
}

export default SuperComputerInfo
