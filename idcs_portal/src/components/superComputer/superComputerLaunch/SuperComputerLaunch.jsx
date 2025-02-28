// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import CustomInput from '../../../utils/customInput/CustomInput'
import ModalInstanceCompare from '../../../utils/modals/modalInstanceCompare/ModalInstanceCompare'
import ModalCreatePublicKey from '../../../utils/modals/modalCreatePublicKey/ModalCreatePublicKey'
import { formatCurrency } from '../../../utils/numberFormatHelper/NumberFormatHelper'
import { BsNodePlusFill } from 'react-icons/bs'
import Button from 'react-bootstrap/Button'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'
import ReservationSubmit from '../../../utils/modals/reservationSubmit/ReservationSubmit'
import { ButtonGroup } from 'react-bootstrap'
import EmptyCatalogModal from '../../../utils/modals/emptyCatalogModal/EmptyCatalogModal'
import LineDivider from '../../../utils/lineDivider/LineDivider'
import UpgradeNeededModal from '../../../utils/modals/upgradeNeededModal/UpgradeNeededModal'
import { Link } from 'react-router-dom'
import Spinner from '../../../utils/spinner/Spinner'

const SuperComputerLaunch = ({
  mainTitle,
  form,
  onChangeInput,
  onChangeDropdownMultiple,
  showInstanceCompareModal,
  onShowHideInstanceCompareModal,
  afterInstanceSelected,
  aiProducts,
  coreComputeProducts,
  navigationBottom,
  isValidForm,
  onSubmit,
  costEstimate,
  showPublicKeyModal,
  onShowHidePublicKeyModal,
  afterPubliKeyCreate,
  onClickComputeAction,
  submitModal,
  errorModal,
  onCloseErrorModal,
  emptyCatalogModal,
  showUpgradeNeededModal,
  setShowUpgradeNeededModal,
  clusterResourceLimit,
  isGeneralComputeAvailable,
  sizeUnit,
  loading
}) => {
  // *****
  // functions
  // *****
  function createCustomInput(element, key, index) {
    return (
      <CustomInput
        key={key}
        type={element.configInput.type}
        fieldSize={element.configInput.fieldSize}
        placeholder={element.configInput.placeholder}
        isRequired={element.configInput.validationRules.isRequired}
        label={
          element.configInput.validationRules.isRequired && element.id !== 'fileVolumneFlag'
            ? element.configInput.label + ' *'
            : element.configInput.label
        }
        maxLength={element.configInput.maxLength}
        value={element.configInput.value}
        onChanged={(event) => onChangeInput(event, element.id, element.idParent, element.nodeIndex)}
        onChangeDropdownMultiple={onChangeDropdownMultiple}
        isValid={element.configInput.isValid}
        isTouched={element.configInput.isTouched}
        helperMessage={element.configInput.helperMessage}
        isReadOnly={element.configInput.isReadOnly}
        options={element.configInput.options}
        validationMessage={element.configInput.validationMessage}
        readOnly={element.configInput.readOnly}
        refreshButton={element.configInput.refreshButton}
        extraButton={element.configInput.extraButton}
        selectAllButton={element.configInput.selectAllButton}
        labelButton={
          element.id === 'computeInstanceType'
            ? { label: 'Delete node group', buttonFunction: () => onClickComputeAction(index, 'Delete') }
            : element.configInput.labelButton
        }
        emptyOptionsMessage={element.configInput.emptyOptionsMessage}
        hidden={element.configInput.hidden}
      />
    )
  }
  // *****
  // Variables
  // *****
  const submitButtonLabels = ['Launch']
  const formAiTypeElements = []
  const formAiConfigElements = []
  const formComputeNodeElements = []
  const fromInstanceDetailElements = []
  const formStorageElements = []
  const formKeysElements = []
  const maxNodeGroupsPerCluster = clusterResourceLimit?.maxnodegroupspercluster || 5

  for (const key in form) {
    const formItem = {
      ...form[key]
    }

    if (key === 'computeNodes') {
      const computeNodes = formItem.items
      for (const index in computeNodes) {
        const node = { ...computeNodes[index] }
        const computeNodeTypeElements = []
        const computeNodeConfigElements = []
        for (const nodeKey in node) {
          const nodeItem = {
            ...node[nodeKey]
          }
          if (nodeItem.subGroup === 'computeNodesType') {
            computeNodeTypeElements.push({
              id: nodeKey,
              idParent: key,
              nodeIndex: index,
              configInput: nodeItem
            })
          }
          if (nodeItem.subGroup === 'computeNodesConfig') {
            computeNodeConfigElements.push({
              id: nodeKey,
              idParent: key,
              nodeIndex: index,
              configInput: nodeItem
            })
          }
        }
        formComputeNodeElements.push({
          id: key,
          typeItems: computeNodeTypeElements,
          configItems: computeNodeConfigElements
        })
      }
    }

    if (formItem.sectionGroup === 'configuration') {
      if (formItem.subGroup === 'instanceType') {
        formAiTypeElements.push({
          id: key,
          configInput: formItem
        })
      }
      if (formItem.subGroup === 'config') {
        formAiConfigElements.push({
          id: key,
          configInput: formItem
        })
      }
    }

    if (formItem.sectionGroup === 'clusterProperties') {
      fromInstanceDetailElements.push({
        id: key,
        configInput: formItem
      })
    }

    if (formItem.sectionGroup === 'storage') {
      formStorageElements.push({
        id: key,
        configInput: formItem
      })
    }

    if (formItem.sectionGroup === 'keys') {
      formKeysElements.push({
        id: key,
        configInput: formItem
      })
    }
  }

  return (
    <>
      <ModalInstanceCompare
        showModalActionConfirmation={showInstanceCompareModal.show}
        closeInstanceCompareModal={onShowHideInstanceCompareModal}
        afterInstanceSelected={afterInstanceSelected}
        products={showInstanceCompareModal.type === 'computeInstance' ? coreComputeProducts : aiProducts}
      />
      <ModalCreatePublicKey
        showModalActionConfirmation={showPublicKeyModal}
        closeCreatePublicKeyModal={onShowHidePublicKeyModal}
        afterPubliKeyCreate={afterPubliKeyCreate}
        isModal={true}
      />
      <ReservationSubmit showReservationCreateModal={submitModal.show} />
      <ErrorModal showModal={errorModal.show} message={errorModal.message} onClickCloseErrorModal={onCloseErrorModal} />
      <UpgradeNeededModal showModal={showUpgradeNeededModal} onClose={() => setShowUpgradeNeededModal(false)} />
      <EmptyCatalogModal
        show={emptyCatalogModal.show}
        product={emptyCatalogModal.product}
        goBackPath={emptyCatalogModal.goBackPath}
        extraExplanation={emptyCatalogModal.extraExplanation}
        extraActions={
          <Link to={'/compute/reserve'} className="btn btn-outline-primary">
            Launch instance
          </Link>
        }
      />
      {loading ? (
        <Spinner />
      ) : (
        <>
          <div className="section">
            <h2 intc-id="superComputeTitle">{mainTitle}</h2>
          </div>
          <div className="d-flex flex-column flex-lg-row align-self-stretch gap-lg-s8">
            <div className="section d-flex flex-column w-100 gap-s6">
              <h3 intc-id="SCSubtitle">1- Cluster configuration</h3>
              <h4 intc-id="AINodesSubtitle">AI nodes</h4>
              {formAiTypeElements.map((element, index) => (
                <div key={index} className="d-flex flex-row w-100 gap-s6">
                  {createCustomInput(element, index)}
                </div>
              ))}
              <div className="d-flex flex-column flex-md-row gap-s6 w-100">
                {formAiConfigElements.map((element, index) => createCustomInput(element, index))}
              </div>
              {isGeneralComputeAvailable ? (
                <>
                  <h4 intc-id="GCNodesSubtitle">General compute nodes</h4>
                  {formComputeNodeElements.map((compute, index) => {
                    return (
                      <React.Fragment key={index}>
                        {compute.typeItems.map((item, subindex) => {
                          return createCustomInput(item, subindex, index)
                        })}
                        <div className="d-flex flex-column flex-md-row gap-s6 w-100">
                          {compute.configItems.map((element, elementIndex) =>
                            createCustomInput(element, elementIndex, index)
                          )}
                        </div>
                      </React.Fragment>
                    )
                  })}
                  <ButtonGroup>
                    <Button
                      intc-id="btn-iksMyClusters-addNode"
                      data-wap_ref="btn-iksMyClusters-addNode"
                      variant="outline-primary"
                      disabled={formComputeNodeElements.length >= maxNodeGroupsPerCluster}
                      onClick={() => onClickComputeAction(null, 'Add')}
                    >
                      <BsNodePlusFill className="me-1" />
                      Add node group
                    </Button>
                    <span>{`Up to ${maxNodeGroupsPerCluster} node groups max. (${
                      maxNodeGroupsPerCluster - formComputeNodeElements.length
                    } remaining)`}</span>
                  </ButtonGroup>
                </>
              ) : null}
              <h3 intc-id="SCProperties">2- Cluster properties</h3>
              {fromInstanceDetailElements.map((element, index) => {
                return createCustomInput(element, index)
              })}
              <h4 intc-id="SCStorage">Storage</h4>
              {formStorageElements.map((element, index) => {
                return createCustomInput(element, index)
              })}
              {formKeysElements.map((element, index) => {
                return createCustomInput(element, index)
              })}
            </div>
            <LineDivider className="h-auto" vertical />
            <div className="section d-flex flex-column gap-s4">
              <h4>Hourly cost estimation</h4>
              <div className="d-flex justify-content-end">
                <span className="fw-semibold" intc-id="SCCostTotal">
                  Total:&nbsp; {costEstimate.costTotal ? formatCurrency(costEstimate.costTotal) : '$ 0'}
                </span>
              </div>
              <div className="d-flex flex-row gap-s4">
                <span className="fw-normal">Control Plane:</span>
                <span className="fw-normal ms-auto" intc-id="SCControlPlaneCost">
                  {` ${costEstimate.controlPlaneHourlyCost ? formatCurrency(costEstimate.controlPlaneHourlyCost) : '$ 0'}`}
                </span>
              </div>
              <div className="d-flex flex-row gap-s4">
                <span className="fw-normal">
                  {`${costEstimate.aiNodeCount ? costEstimate.aiNodeCount : ''} AI nodes: `}
                </span>
                <span className="fw-normal ms-auto" intc-id="SCAINodesCost">
                  {` ${costEstimate.aiHourlyCost ? formatCurrency(costEstimate.aiHourlyCost) : '$0'}`}
                </span>
              </div>
              {isGeneralComputeAvailable ? (
                <div className="d-flex flex-row gap-s4">
                  <span className="fw-normal">
                    {`${costEstimate.computeNodeCount ? costEstimate.computeNodeCount : ''} GC nodes: `}
                  </span>
                  <span className="fw-normal ms-auto" intc-id="SCGCNodesCost">
                    {`${costEstimate.computeHourlyCost ? formatCurrency(costEstimate.computeHourlyCost) : '$0'} `}
                  </span>
                </div>
              ) : null}

              <div className="d-flex flex-row gap-s4">
                <span className="fw-normal">
                  {`${costEstimate.storageGbCount ? costEstimate.storageGbCount : ''} ${sizeUnit} storage: `}
                </span>
                <span className="fw-normal ms-auto" intc-id="SCStorageCost">
                  {` ${costEstimate.storageGbCount ? formatCurrency(costEstimate.storageHourlyCost) : '$0'} `}
                </span>
              </div>
            </div>
          </div>
          <div className="section">
            <ButtonGroup>
              {navigationBottom.map((item, index) => (
                <Button
                  intc-id={`btn-computelaunch-navigationBottom ${item.buttonLabel}`}
                  data-wap_ref={`btn-computelaunch-navigationBottom ${item.buttonLabel}`}
                  aria-label={item.buttonLabel}
                  key={index}
                  variant={item.buttonVariant}
                  onClick={submitButtonLabels.includes(item.buttonLabel) ? onSubmit : item.buttonFunction}
                >
                  {item.buttonLabel}
                </Button>
              ))}
            </ButtonGroup>
          </div>
        </>
      )}
    </>
  )
}

export default SuperComputerLaunch
