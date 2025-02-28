import React from 'react'

const CloudCreditsStepper = (props) => {
  // const stages = props.stages;
  const currentStage = props.currentStage || 0
  const setCurrentStage = props.setCurrentStage
  // const forms = props.forms;
  const newStages = groupStagesWithForms(props.stages, props.forms)
  const instanceSelected = props.instanceSelected

  function groupStagesWithForms (stages, forms) {
    const newStages = []
    for (let i = 0; i < stages.length; i++) {
      if (
        stages[i].sectionGroup === 'cloudcredits' ||
        stages[i].sectionGroup === 'review'
      ) {
        newStages.push({ stage: stages[i], form: null })
      } else {
        const newForms = []
        for (const form in forms) {
          if (forms[form].formGroup === stages[i].sectionGroup) {
            newForms.push(forms[form])
          }
        }
        let isValid = true
        let isTouch = false
        for (let j = 0; j < newForms.length; j++) {
          if (!newForms[j].isValid) {
            isValid = false
          }
          if (newForms[j].isTouched) {
            isTouch = true
          }
        }

        if (newForms.length > 0) {
          newStages.push({
            stage: stages[i],
            form: newForms,
            isValid,
            isTouched: isTouch
          })
        }
      }
    }
    return newStages
  }

  return (
    <>
      <div className="stepper-wrapper" intc-id="reservationStepper">
        <div className="stepper-head  inline-stepper-head">
          <div className="stepper-head  inline-stepper-head">
            {newStages.map((stage, index) => (
              <div
                key={index}
                onClick={
                  instanceSelected ? () => setCurrentStage(index) : null
                }
                className={`stepper-step ${
                  currentStage >= index ? 'is-active' : null
                }`}
                intc-id={stage.stage.title.replaceAll(' ', '') + 'Stepper'}
              >
                <div className="stepper-indicator">
                  <span className="stepper-indicator-info" intc-id="stepperNumber">{index + 1}</span>
                  <div className="stepper-label" intc-id="stepperLabel">{stage.stage.title}</div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </>
  )
}

export default CloudCreditsStepper
