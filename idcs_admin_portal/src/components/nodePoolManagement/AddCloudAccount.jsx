// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import CustomInput from '../../utility/customInput/CustomInput'
import OnSubmitModal from '../../utility/modals/onSubmitModal/OnSubmitModal'
import { Button, Card, ButtonGroup } from 'react-bootstrap'
import SearchBox from '../../utility/searchBox/SearchBox'
import SelectedAccountCard from '../../utility/selectedAccountCard/SelectedAccountCard'
const AddCloudAccount = (props) => {
  // props variables
  const state = props.state
  const mainSubtitle = state.mainSubtitle
  const form = state.form
  const isValidForm = state.isValidForm
  const showLoader = props.showLoader
  const backButtonLabel = props.backButtonLabel
  const selectedPool = props.selectedPool
  const selectedCloudAccount = props.selectedCloudAccount

  // props functions
  const onChangeInput = props.onChangeInput
  const onSearchCloudAccount = props.onSearchCloudAccount
  const onSubmit = props.onSubmit
  const onCancel = props.onCancel

  // functions
  function buildCustomInput(element) {
    return (
      <CustomInput
        key={element.id}
        type={element.configInput.type}
        fieldSize={element.configInput.fieldSize}
        placeholder={element.configInput.placeholder}
        isRequired={element.configInput.validationRules.isRequired}
        label={
          element.configInput.validationRules.isRequired ? element.configInput.label + ' *' : element.configInput.label
        }
        value={element.configInput.value}
        onChanged={(event) => onChangeInput(event, element.id)}
        isValid={element.configInput.isValid}
        isTouched={element.configInput.isTouched}
        helperMessage={element.configInput.helperMessage}
        isReadOnly={element.configInput.isReadOnly}
        options={element.configInput.options}
        validationMessage={element.configInput.validationMessage}
        readOnly={element.configInput.readOnly}
        refreshButton={element.configInput.refreshButton}
        isMultiple={element.configInput.isMultiple}
        onChangeSelectValue={element.configInput.onChangeSelectValue}
        extraButton={element.configInput.extraButton}
        emptyOptionsMessage={element.configInput.emptyOptionsMessage}
      />
    )
  }

  // variables
  const formElementsPool = []
  const formElementsCloudAccount = []

  for (const key in form) {
    const formItem = {
      ...form[key]
    }

    if (formItem.sectionGroup === 'pool') {
      formElementsPool.push({
        id: key,
        configInput: formItem
      })
    }

    if (formItem.sectionGroup === 'cloudAccount') {
      formElementsCloudAccount.push({
        id: key,
        configInput: formItem
      })
    }
  }

  return (
    <>
      <OnSubmitModal showModal={showLoader.isShow} message={showLoader.message}></OnSubmitModal>
      <div className="section">
        <Button intc-id="navigationTop-BackButton" variant="link" className="p-s0" onClick={onCancel}>
          {backButtonLabel}
        </Button>
      </div>

      <div className="section">
        <h2 className="h4">{mainSubtitle}</h2>
      </div>

      <div className="section">
        <div className="row">
          <div className="col-md-6 d-flex flex-column gap-s6">
            {formElementsPool.map((element, index) => (
              <div className="row" key={index}>
                <div className="col-12 col-lg-12">{buildCustomInput(element)}</div>
              </div>
            ))}

            {formElementsCloudAccount.map((element, index) => (
              <div className="d-flex-customInput" key={index}>
                <div className='customInputLabel gap-s6' intc-id="filterText">
                  Cloud Account: *
                </div>
                <div className='col-12'>
                  <SearchBox
                    intc-id="searchCloudAccounts"
                    placeholder={`${element.configInput.placeholder}`}
                    aria-label="Type to search cloud account..."
                    value={ element.configInput.value || ''}
                    onChange={(e) => onChangeInput(e, element.id)}
                    onClickSearchButton={onSearchCloudAccount}
                  />
                  {element.configInput.validationMessage && <div className='invalid-feedback pt-s3' intc-id='cloudAccountSearchError'>{element.configInput.validationMessage}</div>}
                </div>
              </div>
            ))}

            <ButtonGroup>
              <Button
                intc-id="navigationBottom-nodepool-whitelistAccount"
                data-wap_ref="navigationBottom-nodepool-whitelistAccount"
                disabled={!isValidForm}
                variant="primary"
                onClick={onSubmit}
              >
                Save
              </Button>
              <Button intc-id="navigationBottom-Cancel" variant="link" onClick={onCancel}>
                Cancel
              </Button>
            </ButtonGroup>
          </div>
          <div className="col-md-6 d-flex flex-column gap-s6">
            <Card>
              <Card.Header>Selected Node Pool</Card.Header>
              {selectedPool && (
                <Card.Body>
                  <Card.Title>{`Pool ID: ${selectedPool.poolId}`}</Card.Title>
                  <Card.Text>
                    <span className="m-0 d-block">
                      <span className="h6">Pool Name:</span> {selectedPool.poolName}
                    </span>
                    <span className="m-0 d-block">
                      <span className="h6">Pool Account Manager AGS Role: </span>
                      {selectedPool.poolAccountManagerAgsRole}
                    </span>
                    <span className="m-0 d-block">
                      <span className="h6">Number od Nodes: </span>
                      {selectedPool.numberOfNodes}
                    </span>
                  </Card.Text>
                </Card.Body>
              )}
            </Card>
            <SelectedAccountCard selectedCloudAccount={selectedCloudAccount}/>
          </div>
        </div>
      </div>
    </>
  )
}

export default AddCloudAccount
