import React from 'react'
import GridPagination from '../../utility/gridPagination/gridPagination'
import { Button, Modal } from 'react-bootstrap'
import CustomInput from '../../utility/customInput/CustomInput'
import { MdDangerous } from 'react-icons/md'
import SearchBox from '../../utility/searchBox/SearchBox'

function ConfirmationModal(props) {
  const {
    show,
    onHide,
    headerData: {
      title
    },
    bodyData,
    footerData: {
      primary,
      secondary
    }
  } = props

  return (
    <Modal
      className='modal-lg'
      show={show}
      onHide={onHide}
    >
      <Modal.Header closeButton>
        <Modal.Title>{title}</Modal.Title>
      </Modal.Header>

      <Modal.Body className='m-3'>
        <div className="w-100">
          {bodyData}
        </div>
      </Modal.Body>

      <Modal.Footer>
        <Button
          variant={'danger'}
          className='btn'
          onClick={primary.onClick}
        >
          {primary.label}
        </Button>

        <Button variant={'link'} className='btn' onClick={secondary.onClick}>
          Cancel
        </Button>
      </Modal.Footer>
    </Modal>
  )
}

function IMI(props) {
  const {
    gridData: { data, columns, emptyGridObject, loading },
    modalData: {
      id: IMIFormID,
      showModal,
      closeModal,
      primaryLabel,
      formErrorMsg,
      isFormValid,
      imiFormData,
      imiComponentsData,
      onAddComponentClick,
      onRemoveComponentClick,
      formEvents,
      componentFormEvents,
      formFooterEvents,
      showConfirmationModal,
      setShowConfirmationModal
    },
    k8sModalData: {
      showK8sModal,
      closeK8sModal,
      k8sFormErrorMsg,
      isK8sFormValid,
      imiK8sFormData,
      k8sFormEvents,
      k8sFormFooterEvents
    },
    instanceTypeGridData: {
      data: instanceTypesGridItems,
      columns: instanceTypeGridColumns,
      emptyGrid: instanceTypeEmptyGrid,
      loading: instanceTypeLoading,
      onRefreshGridData: onRefreshInstanceTypeGridData
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
    backToHome,
    filterText,
    setFilter,
    openCreateModal
  } = props

  let gridItems = []

  if (filterText !== '' && data) {
    const input = filterText.toLowerCase()
    gridItems = data.filter((item) => {
      const imiName = item.name
      return imiName === undefined || imiName.toLowerCase().includes(input)
    })
    if (gridItems.length === 0) {
      gridItems = data.filter((item) => item.provider.toLowerCase().includes(input))
    }
    if (gridItems.length === 0) {
      gridItems = data.filter((item) => item.state.toLowerCase().includes(input))
    }
    if (gridItems.length === 0) {
      gridItems = data.filter((item) => item.type.toLowerCase().includes(input))
    }
    if (gridItems.length === 0) {
      gridItems = data.filter((item) => item.version.includes(input))
    }
  } else {
    gridItems = data
  }

  function buildCustomInput(data, events, index = null) {
    return (
      <CustomInput
        key={data.id}
        type={data.type}
        fieldSize={data.fieldSize}
        placeholder={data.placeholder}
        isRequired={data.validationRules.isRequired}
        label={data.label}
        value={data.value}
        onChanged={(event) => { events.onChanged(event, data.id, index) }}
        isValid={data.isValid}
        isTouched={data.isTouched}
        helperMessage={data.helperMessage}
        isReadOnly={data.isReadOnly}
        options={data.options}
        validationMessage={data.validationMessage}
        readOnly={data.readOnly}
        refreshButton={data.refreshButton}
        isMultiple={data.isMultiple}
        onChangeSelectValue={data.onChangeSelectValue}
        extraButton={data.extraButton}
        emptyOptionsMessage={data.emptyOptionsMessage}
      />
    )
  }

  return (
    <>
      <div className='section'>
        <Button variant='link' className='p-s0' onClick={() => backToHome()}>
          ⟵ Back to Home
        </Button>
      </div>

      <div className="filter">
        <Button onClick={() => openCreateModal()}>Create</Button>
        <SearchBox
          intc-id="searchImagess"
          value={filterText}
          onChange={setFilter}
          placeholder="Search images..."
          aria-label="Type to search images..."
        />
      </div>
      <div className='section'>
        <GridPagination
          data={gridItems}
          columns={columns}
          emptyGrid={emptyGridObject}
          loading={loading}
        />
      </div>

      <Modal
        id={IMIFormID}
        className='modal-lg'
        show={showModal}
        onHide={() => {
          closeModal()
        }}
      >
        <Modal.Header closeButton>
          <Modal.Title>{'IMI (Intel Machine Image)'}</Modal.Title>
        </Modal.Header>

        <Modal.Body className='m-3'>
          <div>
            <div className='row'>
              <div className='col-12 col-lg-6'>
                {buildCustomInput(imiFormData.name, formEvents)}
              </div>
              <div className='col-12 col-lg-6'>
                {buildCustomInput(imiFormData.upstreamReleaseName, formEvents)}
              </div>
              <div className='col-12 col-lg-6'>
                {buildCustomInput(imiFormData.provider, formEvents)}
              </div>
              <div className='col-12 col-lg-6'>
                {buildCustomInput(imiFormData.type, formEvents)}
              </div>
              <div className='col-12 col-lg-6'>
                {buildCustomInput(imiFormData.runtime, formEvents)}
              </div>
              <div className='col-12 col-lg-6'>
                {buildCustomInput(imiFormData.os, formEvents)}
              </div>
              <div className='col-12 col-lg-6'>
                {buildCustomInput(imiFormData.state, formEvents)}
              </div>
              {primaryLabel !== 'Create'
                ? (
                <div className='col-12 col-lg-6'>
                  {buildCustomInput(imiFormData.artifact, formEvents)}
                </div>
                  )
                : null}
              <div className='col-12 col-lg-6'>
                {buildCustomInput(imiFormData.family, formEvents)}
              </div>
              <div className='col-12 col-lg-6'>
                {buildCustomInput(imiFormData.category, formEvents)}
              </div>
            </div>

            <div className='p-3'></div>

            <section>
              <h5>Components</h5>
              <table className='table'>
                <thead>
                  <tr className='row m-0'>
                    <td className='col-1 d-flex align-items-center'>
                      <button
                        type='button'
                        className='btn btn-primary'
                        disabled={primaryLabel === 'Edit'}
                        onClick={onAddComponentClick}
                      >
                        {' + '}
                      </button>
                    </td>
                    <td className='col-3 d-flex align-items-center'>
                      <small>Name *</small>
                    </td>
                    <td className='col-3 d-flex align-items-center'>
                      <small>Version *</small>
                    </td>
                    <td className='col-5 d-flex align-items-center'>
                      <small>Artifact *</small>
                    </td>
                  </tr>
                </thead>
                <tbody>
                  {imiComponentsData && Array.isArray(imiComponentsData)
                    ? imiComponentsData.map((component, index) => {
                      return (
                          <tr key={index} className='row m-0'>
                            <td className='col-1 d-flex align-items-center'>
                              <button
                                type='button'
                                className='btn btn-danger'
                                disabled={primaryLabel === 'Edit'}
                                onClick={() => {
                                  onRemoveComponentClick(index)
                                }}
                              >
                                <strong> {' - '} </strong>
                              </button>
                            </td>
                            <td className='col-3 d-flex align-items-center'>
                              <div className='mt-3'>
                                {buildCustomInput(component.name, componentFormEvents, index)}
                              </div>
                            </td>
                            <td className='col-3 d-flex align-items-center'>
                              <div className='mt-3'>
                                {buildCustomInput(component.version, componentFormEvents, index)}
                              </div>
                            </td>
                            <td className='col-5 d-flex align-items-center'>
                              <div className='mt-3 w-100'>
                                {buildCustomInput(component.artifact, componentFormEvents, index)}
                              </div>
                            </td>
                          </tr>
                      )
                    })
                    : 'No Components Exist'}
                </tbody>
              </table>
            </section>
          </div>

          <div>
            <div className='invalid-feedback d-block pt-2' intc-id='terminateInstanceError'>
              {formErrorMsg}
            </div>
          </div>

          <div className='p-3'></div>

          <div>
            <div className='row nowrap justify-content-between align-items-center'>
              <div className='col-md-4'>
                <div className='text-start'>
                  <h5 className='my-3 p-0'>Instance Types</h5>
                </div>
              </div>

              {primaryLabel === 'Create'
                ? (
                  <div className='col-md-8'>
                    <div className='text-end'>
                      <Button className='w-auto' onClick={() => onRefreshInstanceTypeGridData()}>Refresh</Button>
                    </div>
                  </div>
                  )
                : null}
            </div>

            <GridPagination
              data={instanceTypesGridItems}
              columns={instanceTypeGridColumns}
              emptyGrid={instanceTypeEmptyGrid}
              loading={instanceTypeLoading}
            />
          </div>
        </Modal.Body>

        <Modal.Footer>
          <Button
            disabled={primaryLabel === 'Edit' ? false : !isFormValid}
            variant={'primary'}
            className='btn'
            onClick={formFooterEvents.onPrimaryBtnClick}
          >
            {primaryLabel}
          </Button>

          {primaryLabel === 'Edit'
            ? (
            <Button
              variant={'danger'}
              className='btn'
              onClick={() => {
                setShowConfirmationModal(true)
              }}
            >
              Delete
            </Button>
              )
            : null}

          <Button variant={'link'} className='btn' onClick={formFooterEvents.onSecondaryBtnClick}>
            Cancel
          </Button>
        </Modal.Footer>
      </Modal>

      <ConfirmationModal
        show={showConfirmationModal}
        onHide={() => { setShowConfirmationModal(false) }}
        headerData={{
          title: 'Delete IMI'
        }}
        bodyData={<p>Are you sure you want to delete?</p>}
        footerData={{
          primary: {
            label: 'Delete',
            onClick: () => {
              setShowConfirmationModal(false)
              formFooterEvents.onDangerBtnClick()
            }
          },
          secondary: {
            onClick: () => {
              setShowConfirmationModal(false)
            }
          }
        }}
      />

      <Modal
        className='modal-lg'
        show={showK8sModal}
        onHide={() => {
          closeK8sModal()
        }}
      >
        <Modal.Header closeButton>
          <Modal.Title>{'IMI to K8s Compatibility'}</Modal.Title>
        </Modal.Header>

        <Modal.Body className='m-3'>
          <div>
            <div className='row'>
              <div className='col-12 col-lg-6'>
                {buildCustomInput(imiK8sFormData.name, k8sFormEvents)}
              </div>
              <div className='col-12 col-lg-6'>
                {buildCustomInput(imiK8sFormData.upstreamReleaseName, k8sFormEvents)}
              </div>
              <div className='col-12 col-lg-6'>
                {buildCustomInput(imiK8sFormData.provider, k8sFormEvents)}
              </div>
              <div className='col-12 col-lg-6'>
                {buildCustomInput(imiK8sFormData.runtime, k8sFormEvents)}
              </div>
              <div className='col-12 col-lg-6'>
                {buildCustomInput(imiK8sFormData.os, k8sFormEvents)}
              </div>
              <div className='col-12 col-lg-6'>
                {buildCustomInput(imiK8sFormData.family, k8sFormEvents)}
              </div>
              <div className='col-12 col-lg-6'>
                {buildCustomInput(imiK8sFormData.category, k8sFormEvents)}
              </div>
            </div>
          </div>

          <div className='py-3'>
            <div className='invalid-feedback d-block pt-2' intc-id='terminateInstanceError'>
              {k8sFormErrorMsg}
            </div>
          </div>

          {imiK8sFormData.type.value === 'worker'
            ? <div>
                <div className='row nowrap justify-content-between align-items-center'>
                  <div className='col-md-4'>
                    <div className='text-start'>
                      <h5 className='my-3 p-0'>Instance Types</h5>
                    </div>
                  </div>
                </div>

                <GridPagination
                  data={instanceTypesGridItems}
                  columns={instanceTypeGridColumns}
                  emptyGrid={instanceTypeEmptyGrid}
                  loading={instanceTypeLoading}
                />
              </div>
            : null
          }
        </Modal.Body>

        <Modal.Footer>
          <Button
            disabled={!isK8sFormValid}
            variant={'primary'}
            className='btn'
            onClick={k8sFormFooterEvents.onPrimaryBtnClick}
          >
            Update K8s
          </Button>

          <Button variant={'link'} className='btn' onClick={k8sFormFooterEvents.onSecondaryBtnClick}>
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

export default IMI
