import React from 'react'
import CustomInput from '../CustomInput'

// CloudCredtsForm: generates form using customInput component
// Dependencies: CustomInput
// Props
// props.stage.data: array of input elements
// props.stage.titleStage: stage title
// props.stage.subTitleStage: stage subTitleStage
// props.stage.stageParam; stage param

const FormWidget = (props) => {
  const data = props.stage.data
  const titleStage = props.stage.titleStage
  const subTitleStage = props.stage.subTitleStage
  const param = props.stage.stageParam ? props.stage.stageParam : null
  return (
    <div className="p-1">
      <h3 intc-id="stageTitle">{titleStage}</h3>

      <br />

      {subTitleStage ? <h4 intc-id="stageSubtitle">subTitleStage</h4> : ''}

      {param}

      <div className="row">
        <div className="col-md-4">
          {data.map((element) =>
            element.config?.type === 'select'
              ? (
              <div className="row col-10" key={element.id}>
                <CustomInput
                  machinename={element.config?.machinename}
                  key={element.id}
                  type={element.config?.type}
                  fieldSize={element.config.fieldSize}
                  description={element.config?.description}
                  placeholder={element.config.placeholder}
                  isRequired={element.config.validationRules.isRequired}
                  label={
                    element.config.validationRules.isRequired
                      ? element.config.label + ' *'
                      : element.config.label
                  }
                  value={element.config.value}
                  onChanged={(event) => props.onChangeInput(event, element.id)}
                  isValid={element.config.isValid}
                  isTouched={element.config.isTouched}
                  isReadOnly={element.config.isReadOnly}
                  selectableOptions={element.config?.options}
                  validationMessage={element.config.validationMessage}
                  readOnly={element.config.readOnly}
                  maxLength={element.config.maxLength}
                />
              </div>
                )
              : (
              <>
                <CustomInput
                  machinename={element.config?.machinename}
                  key={element.id}
                  type={element.config?.type}
                  wildcard={element.config?.wildcard}
                  fieldSize={element.config?.fieldSize}
                  placeholder={element.config?.placeholder}
                  description={element.config?.description}
                  isRequired={element.config?.validationRules.isRequired}
                  label={
                    element.config?.validationRules.isRequired
                      ? element.config?.label + ' *'
                      : element.config?.label
                  }
                  value={element.config?.value}
                  onChanged={(event) => props.onChangeInput(event, element.id)}
                  isValid={element.config?.isValid}
                  isTouched={element.config?.isTouched}
                  isReadOnly={element.config?.isReadOnly}
                  selectableOptions={element.config?.options}
                  validationMessage={element.config?.validationMessage}
                  readOnly={element.config?.readOnly}
                  maxLength={element.config?.maxLength}
                  button={element.config.button}
                  checked={
                    element.config.value ===
                    element.config.options?.input?.value
                  }
                />
              </>
                )
          )}
        </div>
      </div>
    </div>
  )
}

export default FormWidget
