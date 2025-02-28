// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'

const ComingSoonBanner = (props) => {
  const message = props.message

  return (
    <div className="section text-center align-items-center">
      <h1>{props.title}</h1>
      <h2>Coming Soon!</h2>
      <p intc-id="text-comming-soon">{message}</p>
    </div>
  )
}

export default ComingSoonBanner
