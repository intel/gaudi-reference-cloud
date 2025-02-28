// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Button, OverlayTrigger, Popover } from 'react-bootstrap'
import { BsInfoCircle } from 'react-icons/bs'

interface ImageComponentsWithOverlayProps {
  components: any[]
}

const ImageComponentsWithOverlay: React.FC<ImageComponentsWithOverlayProps> = ({ components }) => {
  const componentsLength = components.length

  const popover = (comp: any): JSX.Element => {
    return (
      <Popover className="p-3">
        {comp?.description} Version: {comp?.version}
      </Popover>
    )
  }

  const feedback = components.map((comp: any, index: number) => {
    return (
      <React.Fragment key={index}>
        {componentsLength > 1 && index === componentsLength - 1 ? <>&nbsp;and&nbsp;</> : index > 0 && <>,&nbsp;</>}
        <div className="d-flex flex-row align-items-center">
          <a href={comp?.infoUrl} target="_blank" rel="noreferrer">
            &nbsp;{comp?.name} ({comp?.type})
          </a>
          <OverlayTrigger trigger="focus" placement="right" overlay={popover(comp)}>
            <Button variant="icon-simple">
              <BsInfoCircle />
            </Button>
          </OverlayTrigger>
        </div>
      </React.Fragment>
    )
  })

  return (
    <div className="d-flex flex-row align-items-center flex-wrap">
      {feedback.length > 0 ? 'Image equipped with: ' : ''} {feedback}
    </div>
  )
}

export default ImageComponentsWithOverlay
