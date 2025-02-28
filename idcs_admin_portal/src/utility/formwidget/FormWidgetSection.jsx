import React from 'react'
import Wrapper from '../wrapper/Wrapper'

import FormWidget from './FormWidget'
import CloudCreditsReview from '../../components/billingDashboard/cloudcredits/CloudCreditsReview'

/// Cloud credit FormWidgetSection: Component to display Cloud credit information
/// Props
/// sectionType : Changes the behavior to display information, available options +> card, table, form, review
/// stage: Contains the object to display de information for each stage
// onChangeInput: function needed only for form section type
// sectionTitle: h3 to display at the top of the section
// sectionSubtile: h4 to display as subtitle for the section
const FormWidgetSection = (props) => {
  const stage = props.stage
  const sectionType = props.sectionType.toLowerCase()
  const sectionTitle = props.sectionTitle
  const sectionSubtile = props.sectionSubtile
  let contentView = null
  const setCurrentStage = (stage) => {
    props.setCurrentStage(stage)
  }

  switch (sectionType) {
    case 'form':
      contentView = (
        <FormWidget stage={stage} onChangeInput={props.onChangeInput} />
      )
      break
    case 'review':
      contentView = (
        <CloudCreditsReview
          stage={stage}
          wildcard={props.wildcard}
          setCurrentStage={setCurrentStage}
        />
      )
      break
    default:
      contentView = (
        <FormWidget stage={stage} onChangeInput={props.onChangeInput} />
      )
      break
  }

  return (
    <Wrapper>
      {sectionTitle ? <h3>{sectionTitle}</h3> : null}
      {sectionSubtile ? <h4>{sectionSubtile}</h4> : null}

      {contentView}
    </Wrapper>
  )
}

export default FormWidgetSection
