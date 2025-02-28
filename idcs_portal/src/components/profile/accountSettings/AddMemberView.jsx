// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import Modal from 'react-bootstrap/Modal'
import Button from 'react-bootstrap/Button'
import CustomInput from '../../../utils/customInput/CustomInput'

const AddMemberView = ({ cloudAccountId, addMemberForm, onChangeInput, onChangeDropdownMultiple, roles }) => {
  const modalTitle = `Grant access to cloud account ID: ${cloudAccountId}`
  const form = addMemberForm.fields
  const isValidForm = addMemberForm.isValidForm
  const formElements = []

  const formatRoleOptions = () => {
    const options = []
    for (const role of roles) {
      const option = { name: role.alias, value: role.id }
      options.push(option)
    }
    return options
  }

  for (const key in form) {
    const formElement = { ...form[key] }

    if (key === 'roles') {
      formElement.options = formatRoleOptions()
    }

    formElements.push({
      id: key,
      config: formElement
    })
  }

  if (!addMemberForm.modalOpened) {
    return null
  }

  return (
    <Modal
      show={addMemberForm.modalOpened}
      onHide={() => addMemberForm.cancelAddNewMember()}
      backdrop="static"
      keyboard={false}
      centered
      size="lg"
      aria-label="Add member modal"
    >
      <Modal.Header closeButton>
        <Modal.Title>{modalTitle}</Modal.Title>
      </Modal.Header>
      <Modal.Body>
        <div className="section">
          {formElements.map((element, index) => (
            <CustomInput
              key={index}
              type={element.config.type}
              fieldSize={element.config.fieldSize}
              placeholder={element.config.placeholder}
              isRequired={element.config.validationRules.isRequired}
              label={element.config.validationRules.isRequired ? element.config.label + ' *' : element.config.label}
              value={element.config.value}
              onChanged={(event) => onChangeInput(event, element.id)}
              isValid={element.config.isValid}
              isTouched={element.config.isTouched}
              helperMessage={element.config.helperMessage}
              isReadOnly={element.config.isReadOnly}
              options={element.config.options}
              validationMessage={element.config.validationMessage}
              readOnly={element.config.readOnly}
              refreshButton={element.config.refreshButton}
              extraButton={element.config.extraButton}
              labelButton={element.config.labelButton}
              emptyOptionsMessage={element.config.emptyOptionsMessage}
              maxLength={element.config.maxLength}
              min={element.config.min}
              onChangeDropdownMultiple={(values) => {
                onChangeDropdownMultiple(values, element.id)
              }}
            />
          ))}
        </div>
      </Modal.Body>
      <Modal.Footer>
        <Button
          intc-id={'btn-accessmanagement-addMember-cancel'}
          data-wap_ref={'btn-confirm-addNewMember-cancel'}
          variant="outline-primary"
          aria-label="Cancel"
          onClick={() => addMemberForm.cancelAddNewMember()}
        >
          Cancel
        </Button>
        <Button
          intc-id={'btn-accessmanagement-addMember-grant'}
          data-wap_ref={'btn-accessmanagement-addMember-grant'}
          aria-label="Grant"
          variant="primary"
          disabled={!isValidForm}
          onClick={() => addMemberForm.sendInvite()}
        >
          Grant
        </Button>
      </Modal.Footer>
    </Modal>
  )
}

export default AddMemberView
