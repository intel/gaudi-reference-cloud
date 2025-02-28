// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import ProductCatalogueUnavailable from '../../../pages/error/ProductCatalogueUnavailable'
import TrainingList from './TrainingList'
import JupiterLaunchModal from '../jupiterLaunchModal/JupiterLaunchModal'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'
import ProductFilter from '../../../utils/productFilter/ProductFilter'
import PublicService from '../../../services/PublicService'
import { IDCVendorFamilies } from '../../../utils/Enums'
import { DropdownButton, ButtonGroup, Dropdown } from 'react-bootstrap'
import Spinner from '../../../utils/spinner/Spinner'

const TrainingAndWorkshops = (props) => {
  // props
  const loading = props.loading
  const trainings = props.trainings
  const onSelectTraining = props.onSelectTraining
  const onClickLaunch = props.onClickLaunch
  const enrolling = props.enrolling
  const setEnrolling = props.setEnrolling
  const showErrorModal = props.showErrorModal
  const setShowErrorModal = props.setShowErrorModal
  const errorMessage = props.errorMessage
  const jupyterRes = props.jupyterRes
  const setTagFilter = props.setTagFilter
  const trainingTextFilter = props.trainingTextFilter
  const setTrainingTextFilter = props.setTrainingTextFilter
  const searchFilter = props.searchFilter

  // Variables
  let filterTrainings = []
  let familyFilterList = []

  const getStartedGaudiTrainingName = 'tr-gi-start'
  const getStartedGaudiTrainingId = getTrainingId(getStartedGaudiTrainingName)

  // Distinct to display in Tags filter
  familyFilterList = [...new Set(trainings.map((x) => x.familyDisplayName))].sort(sortFamilyList)

  // Tags filters
  for (const fieldToFilter in searchFilter) {
    const filter = searchFilter[fieldToFilter]
    if (filter.length === 0 || filter.some((x) => x === 'All')) {
      Array.prototype.push.apply(filterTrainings, trainings)
    } else {
      const productFilter = trainings.filter((item) => filter.some((x) => x === item[fieldToFilter]))
      Array.prototype.push.apply(filterTrainings, productFilter)
    }
  }

  if (filterTrainings.length === 0) {
    filterTrainings = trainings
  }

  if (trainingTextFilter) {
    filterTrainings = filterTrainings.filter(
      (item) =>
        item.familyDisplayName.toLowerCase().includes(trainingTextFilter.toLowerCase()) ||
        item.displayCatalogDesc.toLowerCase().includes(trainingTextFilter.toLowerCase()) ||
        item.displayName.toLowerCase().includes(trainingTextFilter.toLowerCase())
    )
  }

  function sortFamilyList(a, b) {
    return (
      PublicService.getCatalogOrder(IDCVendorFamilies.Training, a) -
      PublicService.getCatalogOrder(IDCVendorFamilies.Training, b)
    )
  }

  function getTrainingId(name) {
    if (!name) return null
    const result = trainings.find((training) => training.name === name)
    return result?.id || null
  }

  return (
    <>
      <JupiterLaunchModal jupyterRes={jupyterRes} show={enrolling} onClose={() => setEnrolling(false)} />
      <ErrorModal
        showModal={showErrorModal}
        message={errorMessage}
        titleMessage={'Could not launch notebook'}
        description={'There was an error while processing your request.'}
        onClickCloseErrorModal={() => setShowErrorModal(false)}
      />
      <ProductFilter
        setTagFilter={setTagFilter}
        productsCount={trainings.length}
        onChangeSearchBox={setTrainingTextFilter}
        availableFilters={familyFilterList}
        filterField="familyDisplayName"
        title="Available notebooks"
        extraButton={
          <DropdownButton
            as={ButtonGroup}
            variant="primary"
            title="Connect now"
            intc-id="launchJupiterLabDropdownButton"
          >
            <Dropdown.Item
              intc-id="dropdown-item-training-launchJupterLab-AI-accelerator"
              data-wap_ref="dropdown-item-training-launchJupterLab-AI-accelerator"
              className="text-nowrap"
              onClick={(e) => onClickLaunch(e, getStartedGaudiTrainingId)}
            >
              AI Accelerator
            </Dropdown.Item>
            <Dropdown.Item
              intc-id="dropdown-item-training-launchJupterLab-gpu"
              data-wap_ref="dropdown-item-training-launchJupterLab-gpu"
              className="text-nowrap"
              onClick={(e) => onClickLaunch(e, null)}
            >
              GPU
            </Dropdown.Item>
          </DropdownButton>
        }
      />
      <>
        {loading ? (
          <Spinner />
        ) : (
          <>
            {trainings.length === 0 ? (
              <ProductCatalogueUnavailable />
            ) : (
              <div className="section flex-xs-column flex-md-row gap-s8">
                <TrainingList
                  onClickLaunch={onClickLaunch}
                  trainings={filterTrainings}
                  onSelectTraining={onSelectTraining}
                />
              </div>
            )}
          </>
        )}
      </>
    </>
  )
}

export default TrainingAndWorkshops
