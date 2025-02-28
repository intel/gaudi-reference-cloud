import React from 'react'
import { Button, Modal } from 'react-bootstrap'
import GridPagination from '../../utility/gridPagination/gridPagination'
import CustomInput from '../../utility/customInput/CustomInput'
import { MdDangerous } from 'react-icons/md'
import SearchBox from '../../utility/searchBox/SearchBox'
import SearchCloudAccountModal from '../../utility/modals/searchCloudAccountModal/SearchCloudAccountModal'
function CloudAccountApprovalListComponent(props) {
  const {
    gridData: { data, columns, emptyGrid, loading },
    modalsID: cloudAccApproveListFormID,
    modalData: { isOpen, openModal, addCloudAccFormElements, isFormValid, formErrorMessage, createCloudAccCTA, onChangeCreateCloudAccInput, onCreateCloudAccFormSubmit },
    updateModalData: {
      showUpdateModal,
      closeUpdateModal,
      approveListFormData,
      updateFormEvents,
      isUpdateFormValid,
      updateFormErrorMessage,
      updateFormFooterEvents
    },
    authModalData: {
      showAuthModal,
      closeAuthModal,
      authPassword,
      setAuthPassword,
      authFormFooterEvents,
      showUnAuthModal,
      setShowUnAuthModal
    },
    filterData: { filterText, setFilter },
    cloudAccount,
    selectedCloudAccount,
    cloudAccountError,
    showSearchCloudAccount,
    backToHome,
    handleSearchInputChange,
    handleSubmit,
    setShowSearchModal,
    showSearchAccountLoader
  } = props

  const buildCustomInput = (element, option) => {
    let onClickActionTag
    let onChangeTagValue
    let onChangeInputOption

    switch (option) {
      case 'addCloudAccount':
        onChangeInputOption = onChangeCreateCloudAccInput
        break

      case 'updateCloudAccount':
        onChangeInputOption = updateFormEvents.onChange
        break

      default:
        break
    }

    return (
      <CustomInput
        key={element.id}
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
        onChanged={(event) => onChangeInputOption(event, element.id)}
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
        onClickActionTag={onClickActionTag}
        onChangeTagValue={onChangeTagValue}
      />
    )
  }

  let gridItems = data
  if (filterText !== '' && data) {
    const input = filterText.toLowerCase()

    gridItems = data.filter((item) => item.account.value.toString().startsWith(input))
  }

  return (
    <>
      <SearchCloudAccountModal
        showModal={showSearchCloudAccount}
        selectedCloudAccount={selectedCloudAccount}
        cloudAccount={cloudAccount}
        cloudAccountError={cloudAccountError}
        showLoader={showSearchAccountLoader}
        setShowModal={setShowSearchModal}
        handleSearchInputChange={handleSearchInputChange}
        onClickSearchButton={handleSubmit}
      />
      <div className='section'>
        <Button
          variant='link'
          className='p-s0'
          onClick={() => backToHome()}
        >
          ⟵ Back to Home
        </Button>
      </div>
      <div className="section">
        <div className='filter flex-wrap p-0'>
          <Button variant='primary' className='me-md-2' intc-id="btn-AddCloudAccount" data-wap_ref="btn-AddCloudAccount" onClick={() => { openModal(true) }}>Add Cloud Account</Button>
          <div className='d-flex flex-sm-row flex-xs-column gap-s6'>
            <Button variant='primary' onClick={() => {
            setShowSearchModal(true)
            }}
          >
            Search Cloud Account
          </Button>
            <SearchBox
              intc-id="filterAccounts"
              value={filterText}
              onChange={setFilter}
              placeholder="Filter accounts..."
              aria-label="Type to filter cloud accounts..."
            />
          </div>
        </div>
      </div>

      <div className='section'>
        <GridPagination
          data={gridItems}
          columns={columns}
          emptyGrid={emptyGrid}
          loading={loading}
        />
      </div>

      <Modal
        id={cloudAccApproveListFormID}
        className='modal-lg'
        show={isOpen}
        onHide={() => { openModal(false) }}
      >
        <Modal.Header closeButton>
          <Modal.Title>Add Cloud Account</Modal.Title>
        </Modal.Header>

        <Modal.Body>
          <div className="section">
            {
              addCloudAccFormElements &&
                addCloudAccFormElements.map((element, index) => (
                  <div className='w-100' key={index}>
                    {buildCustomInput(element, 'addCloudAccount')}
                  </div>
                ))
            }

            <div className='d-flex justify-content-center w-100'>
              <div
                className='invalid-feedback d-block p-0'
                intc-id='terminateInstanceError'
              >
                {formErrorMessage}
              </div>
            </div>
          </div>
        </Modal.Body>

        <Modal.Footer>
        {
          createCloudAccCTA.map((item, index) => (
            <Button
              key={index}
              disabled={item.buttonLabel === 'Create' ? !isFormValid : false}
              variant={item.buttonVariant}
              className='btn'
              onClick={item.buttonLabel === 'Create' ? onCreateCloudAccFormSubmit : item.buttonFunction} >
              {item.buttonLabel}
            </Button>
          ))
        }
        </Modal.Footer>
      </Modal>

      <Modal
        id={cloudAccApproveListFormID}
        className="modal-lg"
        show={showUpdateModal}
        onHide={() => {
          closeUpdateModal()
        }}
      >
        <Modal.Header closeButton>
          <Modal.Title>Update Cloud Account</Modal.Title>
        </Modal.Header>

        <Modal.Body className="m-3">
          <section>
            <div className="row">
              {approveListFormData &&
                approveListFormData.map((element) => (
                  <div className="col-6 col-lg-6" key={element.id}>{buildCustomInput(element, 'updateCloudAccount')}</div>
                ))}
            </div>
          </section>

          <div className="py-3">
            <div className="invalid-feedback d-block pt-2" intc-id="terminateInstanceError">
              {updateFormErrorMessage}
            </div>
          </div>
        </Modal.Body>

        <Modal.Footer>
          <Button
            disabled={!isUpdateFormValid}
            variant={'primary'}
            className='btn'
            onClick={updateFormFooterEvents.onPrimaryBtnClick}
          >
            Update
          </Button>

          <Button
            variant={'link'}
            className='btn'
            onClick={updateFormFooterEvents.onSecondaryBtnClick}
          >
            Cancel
          </Button>
        </Modal.Footer>
      </Modal>

      <Modal
        className='modal-lg'
        backdrop="static"
        centered
        show={showAuthModal}
        onHide={() => {
          closeAuthModal()
        }}
      >
        <Modal.Header closeButton>
          <Modal.Title>{'Please Enter Password to Perform Action'}</Modal.Title>
        </Modal.Header>

        <Modal.Body className='m-3'>
          <div>
            <div className='row'>
              <div className='col-12 col-lg-6'>
                <CustomInput
                  key={'password'}
                  type={'password'}
                  fieldSize={'small'}
                  placeholder={'Enter Password'}
                  isRequired={true}
                  label={'Authorization Password'}
                  value={authPassword}
                  onChanged={(event) => { setAuthPassword(event.target.value) }}
                  isValid={true}
                />
              </div>
            </div>
          </div>
        </Modal.Body>

        <Modal.Footer>
          <Button
            variant={'primary'}
            className='btn'
            onClick={authFormFooterEvents.onPrimaryBtnClick}
          >
            Submit
          </Button>

          <Button variant={'link'} className='btn' onClick={authFormFooterEvents.onSecondaryBtnClick}>
            Cancel
          </Button>
        </Modal.Footer>
      </Modal>

      <Modal
        className='modal-lg'
        centered
        show={showUnAuthModal}
        onHide={() => {
          setShowUnAuthModal(false)
        }}
      >
        <Modal.Body className='m-3'>
          <br />
          <div className="text-center">
            <MdDangerous color="red" size="3em" />
            <h5>Unauthorized</h5>
            <p>
              You are not allowed to perform this action. Please contact Admin.
            </p>
          </div>
        </Modal.Body>

        <Modal.Footer>
          <Button
            variant={'primary'}
            className='btn'
            onClick={() => {
              setShowUnAuthModal(false)
            }}
          >
            Go Back
          </Button>
        </Modal.Footer>
      </Modal>
    </>
  )
}

export default CloudAccountApprovalListComponent
