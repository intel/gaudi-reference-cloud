// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import Modal from 'react-bootstrap/Modal'
import Button from 'react-bootstrap/Button'
import CustomAlerts from '../../../utils/customAlerts/CustomAlerts'
import idcConfig from '../../../config/configurator'

const RemoveMemberView = ({ removeMember }) => {
  const modalTitle = removeMember.modalTitle || 'Remove access'
  const buttonTitle = removeMember.buttonTitle || 'Remove access'

  if (!removeMember.modalOpened) {
    return null
  }

  const functionArgument = {
    action: removeMember.action,
    email: removeMember.emailAddress,
    invitationState: removeMember.invitationState
  }

  return (
    <Modal
      show={removeMember.modalOpened}
      onHide={() => removeMember.cancelRemoveMember()}
      backdrop="static"
      keyboard={false}
      centered
      size="lg"
      aria-label="Remove member modal"
    >
      <Modal.Header closeButton>
        <Modal.Title>{modalTitle}</Modal.Title>
      </Modal.Header>
      <Modal.Body>
        <div className="section">
          {removeMember.action === 'removeAcceptedInvite' ? (
            <CustomAlerts
              showAlert
              showIcon
              alertType="warning"
              message="Manually remove all member's SSH keys from shared instances. This ensures they no longer have access."
              link={{
                openInNewTab: true,
                href: `${idcConfig.REACT_APP_MULTIUSER_GUIDE}#revoke-access`,
                label: 'Learn how to manually remove keys'
              }}
            />
          ) : null}
          Do you want to {modalTitle.toString().toLowerCase()} to {removeMember.emailAddress}?<br />
          {removeMember.action === 'removeAcceptedInvite'
            ? 'They will no longer be able to view your resources. Manually remove their SSH keys from all running resources.'
            : null}
        </div>
      </Modal.Body>
      <Modal.Footer>
        <Button
          intc-id={'btn-accessmanagement-cancel'}
          data-wap_ref={'btn-accessmanagement-cancel'}
          variant="outline-primary"
          aria-label="Cancel"
          onClick={() => removeMember.cancelRemoveMember()}
        >
          Cancel
        </Button>
        <Button
          intc-id={'btn-accessmanagement-confirm'}
          data-wap_ref={'btn-accessmanagement-confirm'}
          variant="danger"
          aria-label={buttonTitle}
          onClick={() => removeMember.okRemoveMember(functionArgument)}
        >
          {buttonTitle}
        </Button>
      </Modal.Footer>
    </Modal>
  )
}

export default RemoveMemberView
