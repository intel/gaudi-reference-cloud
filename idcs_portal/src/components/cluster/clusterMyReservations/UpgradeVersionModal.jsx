// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Modal from 'react-bootstrap/Modal'
import CustomInput from '../../../utils/customInput/CustomInput'
import idcConfig from '../../../config/configurator'
import Button from 'react-bootstrap/Button'

const UpgradeVersionModal = (props) => {
  const show = props.show
  const onHide = props.onHide
  const backdrop = props.backdrop
  const size = props.size
  const centered = props.centered
  const closeButton = props.closeButton
  const formElementskeys = props.formElementskeys
  const onChangeUpgradeForm = props.onChangeUpgradeForm
  const isValidUpgradeForm = props.isValidUpgradeForm
  const submitUpgradeK8sVersion = props.submitUpgradeK8sVersion
  return (
    <Modal
      show={show}
      onHide={() => onHide(false)}
      backdrop={backdrop}
      size={size}
      centered={centered}
      aria-label="Upgrade version modal"
    >
      <Modal.Header closeButton={closeButton}>
        <Modal.Title>Upgrade Kubernetes version</Modal.Title>
      </Modal.Header>
      <Modal.Body>
        {formElementskeys.map((element, index) => (
          <CustomInput
            key={index}
            type={element.configInput.type}
            fieldSize={element.configInput.fieldSize}
            placeholder={element.configInput.placeholder}
            isRequired={element.configInput.validationRules.isRequired}
            label={
              element.configInput.validationRules.isRequired
                ? element.configInput.label + ' *'
                : element.configInput.label
            }
            value={element.configInput.value}
            onChanged={(event) => onChangeUpgradeForm(event, element.id)}
            isValid={element.configInput.isValid}
            isTouched={element.configInput.isTouched}
            helperMessage={element.configInput.helperMessage}
            isReadOnly={element.configInput.isReadOnly}
            options={element.configInput.options}
            validationMessage={element.configInput.validationMessage}
            readOnly={element.configInput.readOnly}
            refreshButton={element.configInput.refreshButton}
            extraButton={element.configInput.extraButton}
            labelButton={element.configInput.labelButton}
            emptyOptionsMessage={element.configInput.emptyOptionsMessage}
          />
        ))}
        <div className="section">
          <h6>Release Notes:</h6>
          <span>
            Please read the following release notes:
            <a
              href={idcConfig.REACT_APP_KUBERNETES_RELEASE_URL}
              target="_blank"
              rel="noreferrer"
              className="alert-link"
            >
              {' '}
              Here
            </a>
          </span>
        </div>
      </Modal.Body>
      <Modal.Footer>
        <Button
          intc-id="btn-iksUpgradeClusterModal-cancel"
          data-wap_ref="btn-iksUpgradeClusterModal-cancel"
          variant="outline-primary"
          onClick={() => onHide(false)}
        >
          Cancel
        </Button>
        <Button
          intc-id="btn-iksUpgradeClusterModal-upgrade"
          data-wap_ref="btn-iksUpgradeClusterModal-upgrade"
          disabled={!isValidUpgradeForm}
          onClick={() => submitUpgradeK8sVersion()}
          variant="primary"
        >
          Upgrade
        </Button>
      </Modal.Footer>
    </Modal>
  )
}

export default UpgradeVersionModal
