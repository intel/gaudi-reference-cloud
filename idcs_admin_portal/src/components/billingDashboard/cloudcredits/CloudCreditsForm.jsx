import { Button } from 'react-bootstrap'

import FormWidgetSection from '../../../utility/formwidget/FormWidgetSection'
import CloudCreditsStepper from '../../../utility/steppers/CloudCreditsStepper'
import Wrapper from '../../../utility/wrapper/Wrapper'

const CloudCreditsForm = (props) => {
  // stagesData object contains all the elements to be displayed in the page
  const stagesData = []
  const currentStage = props.currentStage

  const instanceSelected = props.instanceSelected

  // Instance Type Name
  const instanceDisplayNameTxt = instanceSelected
    ? instanceSelected.instanceDisplayNameTxt
    : null

  // Iteration to prepare the  information to be display in each stage
  for (const index in props.stages) {
    // Get copy of the stage to update with additional data
    const stage = { ...props.stages[index] }

    // section group helps to organize the form elements to be display on each stage
    const sectionGroup = stage.sectionGroup

    if (sectionGroup === 'cloudcredits') {
      buildCloudCreditsDetailsSection(stage, sectionGroup)
    }

    if (sectionGroup === 'review') {
      // Change the type base on instance type selection
      stage.subTitleStage = stage.subTitleStage + instanceDisplayNameTxt
      buildReviewSection(stage)
    }

    // Add the new objet to the stagesData array to be sent to the generic component
    stagesData.push(stage)
  }

  // Function to build publicKeys items to display
  function buildCloudCreditsDetailsSection (stage, sectionGroup) {
    stage.data = buildFormElements(sectionGroup)
  }

  // Function to build form elements to a specific stage
  function buildFormElements (sectionGroup) {
    // Array of data that contain all the information
    const sectionElements = []

    // Go through the form elements to select the items based on the formGroup
    for (const element in props.form) {
      if (props.form[element].formGroup === sectionGroup) {
        // Form elements for the section
        sectionElements.push({
          id: element,
          config: props.form[element],
          onChange: props.onChangeInput
        })
      }
    }

    return sectionElements
  }

  // Function to build review items to display
  function buildReviewSection (stage) {
    const reviewElements = []

    const reviewElement = {}

    reviewElement.sectionTitle = 'Cloud Credit Details'

    // index to be used a parameter for the sectionTitleFunction
    reviewElement.sectionTitleIndex = 1
    // const setCurrentStage = props.setCurrentStage;
    // function to change current stage
    // reviewElement.sectionTitleFunction = props?.setCurrentStage;
    reviewElement.sectionTitleFunction = props.currentStage

    reviewElement.formElements = buildReviewFormElements('cloudcredits')

    reviewElements.push(reviewElement)

    stage.data = reviewElements
  }

  // Function to build input elements for review section
  function buildReviewFormElements (sectionGroup) {
    const formElements = []
    for (const element in props.form) {
      if (props.form[element].formGroup === sectionGroup) {
        formElements.push({
          label: props.form[element].label,
          value:
            !props.form[element].isValid &&
            // && !props.form[element].value
            props.form[element].isTouched
              ? props.form[element].validationMessage
                ? props.form[element].validationMessage
                : props.form[element].label + ' is required'
              : props.form[element].value
        })
      }
    }

    return formElements
  }

  function checkCurrectStageFormValidation (data) {
    const filter = data.filter((e) => e.config.isValid === false)
    return filter.length !== 0
  }

  return (
    <Wrapper>
      {/* <Row className="ps-2 mb-2 col-12">
        <span className="col-6 pt-4 pl-2">
          <h2 className="pt-2">Create Cloud Credits</h2>
          <p className="pl-2">Specific the needed information</p>
        </span>
        <span className="col-6 text-end pt-4">
          <Button
            variant="link"
            onClick={(e) => props.whichPageToOpen("cloudcredits")}
            className="btn-sm"
          >
            Back to list
          </Button>
        </span>
      </Row> */}

      <div className="m-3">
        <h2>Create Cloud Credits</h2>

        <p className="lead"> {stagesData[1].description}</p>
        <CloudCreditsStepper
          instanceSelected={instanceSelected}
          stages={stagesData}
          currentStage={currentStage}
          setCurrentStage={props.setCurrentStage}
          forms={props.form}
        />
        <div className="row">
          {props.loading
            ? (
            <div className="col-12">
              <div className="spinner-border text-primary center"></div>
            </div>
              )
            : (
            <Wrapper>
              <div
                className={`${
                  stagesData[currentStage].show_right_details ? 'col-10' : ''
                }`}
              >
                <FormWidgetSection
                  sectionType={stagesData[currentStage].sectionType}
                  stage={stagesData[currentStage]}
                  onChangeInput={props.onChangeInput}
                  setCurrentStage={props.setCurrentStage}
                  wildcard={props.form.audience.wildcard}
                />

                {stagesData[currentStage].show_nav_buttons
                  ? (
                  <div className="mt-4">
                    {stagesData.length - 1 === currentStage
                      ? (
                      <Button
                        variant="primary"
                        className="btn-sm me-3"
                        onClick={() => props.onSubmitHandler()}
                        intc-id="button-cloudcreditform-create"
                      >
                        Create Code
                      </Button>
                        )
                      : (
                      <Button
                        variant="primary"
                        className="btn-sm me-3"
                        intc-id="button-cloudcreditform-next"
                        disabled={checkCurrectStageFormValidation(
                          stagesData[currentStage].data
                        )}
                        onClick={() => props.setCurrentStage(currentStage + 1)}
                      >
                        Next
                      </Button>
                        )}

                    {stagesData[currentStage].show_nav_buttons &&
                    currentStage >= 1
                      ? (
                      <Button
                        variant="outline-primary"
                        className="btn-sm me-3"
                        intc-id="button-cloudcreditform-back"
                        onClick={() => props.setCurrentStage(currentStage - 1)}
                      >
                        Back to {stagesData[currentStage - 1].title}
                      </Button>
                        )
                      : null}
                    <Button
                      variant="link"
                      className="btn-sm"
                      intc-id="button-cloudcreditform-cancel"
                      onClick={() => props.onCancelHandler()}
                    >
                      Cancel
                    </Button>
                  </div>
                    )
                  : null}
              </div>
            </Wrapper>
              )}
        </div>
      </div>
    </Wrapper>
  )
}

export default CloudCreditsForm
