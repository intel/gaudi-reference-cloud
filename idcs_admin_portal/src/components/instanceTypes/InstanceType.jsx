import React from 'react'
import GridPagination from '../../utility/gridPagination/gridPagination'
import { Button, Modal } from 'react-bootstrap'
import CustomInput from '../../utility/customInput/CustomInput'
import ConfirmationModal from '../../utility/modals/onConfirmModal/OnConfirmModal'
import { MdDangerous } from 'react-icons/md'
import SearchBox from '../../utility/searchBox/SearchBox'

function InstanceType(props) {
  const {
    gridData: { data, columns, emptyGridObject, loading },
    modalData: {
      id: instanceTypeFormID,
      showModal,
      closeModal,
      primaryLabel,
      formErrorMsg,
      isFormValid,
      instanceTypeFormData,
      computeInstanceFilterText,
      setComputeInstanceFilter,
      formEvents,
      formFooterEvents,
      showConfirmationModal,
      setShowConfirmationModal
    },
    k8sModalData: {
      showK8sModal,
      closeK8sModal,
      isK8sFormValid,
      k8sFormErrorMsg,
      k8sFormFooterEvents
    },
    iksInstanceTypeData,
    imiK8sGridData: {
      data: imiK8sGridItems,
      columns: imiK8sGridColumns,
      emptyGrid: imiK8sEmptyGrid,
      loading: imiK8sLoading
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
    imiGridData,
    computeInstanceTypeData,
    computeInstanceGridData,
    backToHome,
    filterText,
    setFilter,
    openCreateModal
  } = props

  let gridItems = []
  if (filterText !== '' && data) {
    const input = filterText.toLowerCase()

    gridItems = data.filter((item) => {
      const instancetypename = item.instancetypename
      return instancetypename === undefined || instancetypename.toLowerCase().includes(input)
    })
    if (gridItems.length === 0) {
      gridItems = data.filter((item) => item.status.toLowerCase().includes(input))
    }
    if (gridItems.length === 0) {
      gridItems = data.filter((item) => item.displayname.toLowerCase().includes(input))
    }
  } else {
    gridItems = data
  }

  let computeInstanceGridItems = []
  if (computeInstanceGridData.data && computeInstanceFilterText !== '') {
    const query = computeInstanceFilterText.toLowerCase()

    computeInstanceGridItems = computeInstanceGridData.data.filter((item) => {
      const instancetypename = item.instancetypename

      return (instancetypename === undefined || instancetypename.toLowerCase().includes(query))
    })
  } else {
    computeInstanceGridItems = computeInstanceGridData.data
  }

  function buildCustomInput(data, events) {
    return (
      <CustomInput
        key={data.id}
        type={data.type}
        fieldSize={data.fieldSize}
        placeholder={data.placeholder}
        isRequired={data.validationRules.isRequired}
        label={data.label}
        value={data.value}
        onChanged={(event) => {
          events.onChanged(event, data.id)
        }}
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
          intc-id="searchInstanceTypes"
          value={filterText}
          onChange={setFilter}
          placeholder="Search instance types..."
          aria-label="Type to instance types..."
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
        id={instanceTypeFormID}
        className='modal-lg'
        show={showModal}
        onHide={() => {
          closeModal()
        }}
      >
        <Modal.Header closeButton>
          <Modal.Title>{'Instance Types'}</Modal.Title>
        </Modal.Header>

        <Modal.Body>
          {primaryLabel === 'Create'
            ? (
              <section>
                <div className="mb-s6 d-flex flex-lg-row flex-xs-column gap-s4 justify-content-lg-between justify-content-xs-start align-items-lg-center">
                  <h5>Compute Instance Types</h5>
                  <div className="d-inline-flex">
                      <SearchBox
                        intc-id="searchComputeInstanceTypes"
                        value={computeInstanceFilterText}
                        onChange={setComputeInstanceFilter}
                        placeholder="Search compute instance types..."
                        aria-label="Type to compute instance types.."
                      />
                  </div>
                </div>

                <GridPagination
                  data={computeInstanceGridItems}
                  columns={computeInstanceGridData.columns}
                  emptyGrid={computeInstanceGridData.emptyGrid}
                  loading={computeInstanceGridData.loading}
                />
              </section>
              )
            : null
          }

          {primaryLabel !== 'Create'
            ? (
              <section>
                {computeInstanceTypeData?.instancetypename
                  ? (
                    <>
                      <div className='row nowrap justify-content-between align-items-center'>
                        <div className='col-md-4'>
                          <div className='text-start'>
                            <h5 className='my-3 p-0'>Compute Instance Type</h5>
                          </div>
                        </div>
                      </div>

                      <div className="m-3">
                        <div className="row">
                          <div className="col-12 col-lg-4 border fw-bold">Instance Type Name</div>
                          <div className="col-12 col-lg-8 border">{computeInstanceTypeData.instancetypename}</div>

                          <div className="col-12 col-lg-4 border fw-bold">Memory (GB)</div>
                          <div className="col-12 col-lg-8 border">{computeInstanceTypeData.memory}</div>

                          <div className="col-12 col-lg-4 border fw-bold">CPU Cores</div>
                          <div className="col-12 col-lg-8 border">{computeInstanceTypeData.cpu}</div>

                          <div className="col-12 col-lg-4 border fw-bold">Node Provider Name</div>
                          <div className="col-12 col-lg-8 border">{computeInstanceTypeData.nodeprovidername}</div>

                          <div className="col-12 col-lg-4 border fw-bold">Storage (GB)</div>
                          <div className="col-12 col-lg-8 border">{computeInstanceTypeData.storage}</div>

                          <div className="col-12 col-lg-4 border fw-bold">Status</div>
                          <div className="col-12 col-lg-8 border">{computeInstanceTypeData.status}</div>

                          <div className="col-12 col-lg-4 border fw-bold">Display Name</div>
                          <div className="col-12 col-lg-8 border">{computeInstanceTypeData.displayname}</div>

                          <div className="col-12 col-lg-4 border fw-bold">IMI Override</div>
                          <div className="col-12 col-lg-8 border">{computeInstanceTypeData.imioverride ? 'Yes' : 'No'}</div>

                          <div className="col-12 col-lg-4 border fw-bold">Description</div>
                          <div className="col-12 col-lg-8 border">{computeInstanceTypeData.description}</div>

                          <div className="col-12 col-lg-4 border fw-bold">Family</div>
                          <div className="col-12 col-lg-8 border">{computeInstanceTypeData.family}</div>

                          <div className="col-12 col-lg-4 border fw-bold">Category</div>
                          <div className="col-12 col-lg-8 border">{computeInstanceTypeData.category}</div>
                        </div>
                      </div>
                    </>
                    )
                  : (
                    <h2>No Compute Instance Found Matching {instanceTypeFormData.instancetypename.value}</h2>
                    )
                }
              </section>
              )
            : null
          }

          <div className='p-3'></div>

          <div className='row nowrap justify-content-between align-items-center'>
            <div className='col-md-4'>
              <div className='text-start'>
                <h5 className='my-3 p-0'>IKS Instance Type</h5>
              </div>
            </div>
          </div>

          <div>
            <div className='row'>
              <div className='col-12 col-lg-6'>
                {buildCustomInput(instanceTypeFormData.instancetypename, formEvents)}
              </div>
              <div className='col-12 col-lg-6'>
                {buildCustomInput(instanceTypeFormData.memory, formEvents)}
              </div>
              <div className='col-12 col-lg-6'>
                {buildCustomInput(instanceTypeFormData.cpu, formEvents)}
              </div>
              <div className='col-12 col-lg-6'>
                {buildCustomInput(instanceTypeFormData.nodeprovidername, formEvents)}
              </div>
              <div className='col-12 col-lg-6'>
                {buildCustomInput(instanceTypeFormData.storage, formEvents)}
              </div>
              <div className='col-12 col-lg-6'>
                {buildCustomInput(instanceTypeFormData.status, formEvents)}
              </div>
              <div className='col-12 col-lg-6'>
                {buildCustomInput(instanceTypeFormData.displayname, formEvents)}
              </div>
              <div className='col-12 col-lg-6'>
                <div className='h-100 d-inline-flex align-items-center mt-2'>
                  {buildCustomInput(instanceTypeFormData.imioverride, formEvents)}
                </div>
              </div>
              <div className='col-12 col-lg-6'>
                {buildCustomInput(instanceTypeFormData.description, formEvents)}
              </div>
              <div className='col-12 col-lg-6'>
                {buildCustomInput(instanceTypeFormData.category, formEvents)}
              </div>
              <div className='col-12 col-lg-6'>
                {buildCustomInput(instanceTypeFormData.family, formEvents)}
              </div>
              {primaryLabel === 'Create'
                ? (
                  <div className='col-12 col-lg-6'>
                    <div className='h-100 d-inline-flex align-items-center mt-2'>
                      {buildCustomInput(instanceTypeFormData.allowManualInsert, formEvents)}
                    </div>
                  </div>
                  )
                : null
              }
            </div>
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
                  <h5 className='my-3 p-0'>IMIS</h5>
                </div>
              </div>

              {primaryLabel === 'Create'
                ? (
                  <div className='col-md-8'>
                    <div className='text-end'>
                      <Button className='w-auto' onClick={() => imiGridData.onRefreshGridData()}>
                        Refresh
                      </Button>
                    </div>
                  </div>
                )
                : null }
            </div>

            <GridPagination
              data={imiGridData.data}
              columns={imiGridData.columns}
              emptyGrid={imiGridData.emptyGrid}
              loading={imiGridData.loading}
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

      <Modal
        id={instanceTypeFormID}
        className='modal-lg'
        show={showK8sModal}
        onHide={() => {
          closeK8sModal()
        }}
      >
        <Modal.Header closeButton>
          <Modal.Title>{'Title'}</Modal.Title>
        </Modal.Header>

        <Modal.Body className='m-3'>
          <section>
            {iksInstanceTypeData?.instancetypename
              ? (
                <>
                  <div className='row nowrap justify-content-between align-items-center'>
                    <div className='col-md-4'>
                      <div className='text-start'>
                        <h5 className='my-3 p-0'>IKS Instance Type</h5>
                      </div>
                    </div>
                  </div>

                  <div className="m-3">
                    <div className="row">
                      <div className="col-12 col-lg-4 border fw-bold">Instance Type Name</div>
                      <div className="col-12 col-lg-8 border">{iksInstanceTypeData.instancetypename}</div>

                      <div className="col-12 col-lg-4 border fw-bold">Memory (GB)</div>
                      <div className="col-12 col-lg-8 border">{iksInstanceTypeData.memory}</div>

                      <div className="col-12 col-lg-4 border fw-bold">CPU Cores</div>
                      <div className="col-12 col-lg-8 border">{iksInstanceTypeData.cpu}</div>

                      <div className="col-12 col-lg-4 border fw-bold">Node Provider Name</div>
                      <div className="col-12 col-lg-8 border">{iksInstanceTypeData.nodeprovidername}</div>

                      <div className="col-12 col-lg-4 border fw-bold">Storage (GB)</div>
                      <div className="col-12 col-lg-8 border">{iksInstanceTypeData.storage}</div>

                      <div className="col-12 col-lg-4 border fw-bold">Status</div>
                      <div className="col-12 col-lg-8 border">{iksInstanceTypeData.status}</div>

                      <div className="col-12 col-lg-4 border fw-bold">Display Name</div>
                      <div className="col-12 col-lg-8 border">{iksInstanceTypeData.displayname}</div>

                      <div className="col-12 col-lg-4 border fw-bold">IMI Override</div>
                      <div className="col-12 col-lg-8 border">{iksInstanceTypeData.imioverride ? 'Yes' : 'No'}</div>

                      <div className="col-12 col-lg-4 border fw-bold">Description</div>
                      <div className="col-12 col-lg-8 border">{iksInstanceTypeData.description}</div>

                      <div className="col-12 col-lg-4 border fw-bold">Family</div>
                      <div className="col-12 col-lg-8 border">{iksInstanceTypeData.family}</div>

                      <div className="col-12 col-lg-4 border fw-bold">Category</div>
                      <div className="col-12 col-lg-8 border">{iksInstanceTypeData.category}</div>
                    </div>
                  </div>
                </>
                )
              : null
            }
          </section>

          <div className='py-3'>
            <div className='invalid-feedback d-block pt-2' intc-id='terminateInstanceError'>
              {k8sFormErrorMsg}
            </div>
          </div>

          <div>
            <div className='row nowrap justify-content-between align-items-center'>
              <div className='col-md-4'>
                <div className='text-start'>
                  <h5 className='my-3 p-0'>Tag K8s Compatable IMIs</h5>
                </div>
              </div>
            </div>

            <GridPagination
              data={imiK8sGridItems}
              columns={imiK8sGridColumns}
              emptyGrid={imiK8sEmptyGrid}
              loading={imiK8sLoading}
            />
          </div>
        </Modal.Body>

        <Modal.Footer>
          <Button
            disabled={!isK8sFormValid}
            variant={'primary'}
            className='btn'
            onClick={k8sFormFooterEvents.onPrimaryBtnClick}
          >
            Update
          </Button>

          <Button
            variant={'link'}
            className='btn'
            onClick={k8sFormFooterEvents.onSecondaryBtnClick}
          >
            Cancel
          </Button>
        </Modal.Footer>
      </Modal>

      <ConfirmationModal
        confirmModalData={{
          isShow: showConfirmationModal,
          onClose: () => {
            setShowConfirmationModal(false)
          },
          title: 'Delete Instance Type',
          data: [
            {
              col: 'Instance Type Name',
              value: instanceTypeFormData.instancetypename.value
            }
          ]
        }}
        onSubmit={() => {
          setShowConfirmationModal(false)
          formFooterEvents.onDangerBtnClick()
        }}
      />

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

export default InstanceType
