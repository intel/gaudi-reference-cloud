// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useMemo } from 'react'
import { ReactComponent as ExternalLink } from '../../assets/images/ExternalLink.svg'

interface ImageComponentsProps {
  components: any[]
}

const ImageComponents: React.FC<ImageComponentsProps> = ({ components }) => {
  const componentsLength = components.length

  const feedback = useMemo(
    () =>
      components.map((comp: any, index: number) => {
        return (
          <React.Fragment key={index}>
            {componentsLength > 1 && index === componentsLength - 1 ? <>&nbsp;and&nbsp;</> : index > 0 && <>,&nbsp;</>}
            <div className="d-flex flex-row align-items-center">
              <a href={comp?.infoUrl} target="_blank" rel="noreferrer" className="d-flex gap-s4 align-items-center">
                &nbsp;{comp?.name} ({comp?.type})
                <ExternalLink />
              </a>
            </div>
          </React.Fragment>
        )
      }),
    []
  )

  return (
    <div className="d-flex flex-row align-items-center flex-wrap">
      {feedback.length > 0 ? 'Image equipped with: ' : ''} {feedback}
    </div>
  )
}

export default ImageComponents
