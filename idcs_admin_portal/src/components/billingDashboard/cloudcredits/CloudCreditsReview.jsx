import React from 'react'

const CloudCreditsReview = (props) => {
  const data = props.stage.data
  const formElements = props.stage.data[0].formElements

  return (
    <div className="col-md-12">
      <br />

      <div className="grid row">
        <div className="input-group">
          <h3 className="pl-0">{data[0].sectionTitle}</h3>
          <button
            type="button"
            className="btn btn-link btn-sm"
            onClick={() => props.setCurrentStage(0)}
          >
            Edit
          </button>
        </div>
        <div className="col-12 ps-3 pb-2 pt-4">
          {formElements.map((element, indexElement) => (
            <div key={indexElement} className="grid row">
              <b>{element.label}</b>
              <p>
                {element.value.toString() === 'wildcard'
                  ? props.wildcard
                  : element.value.toString()}
              </p>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

export default CloudCreditsReview
