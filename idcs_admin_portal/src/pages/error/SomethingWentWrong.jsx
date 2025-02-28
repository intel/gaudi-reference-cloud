// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

const SomethingWentWrong = ({ error }) => {
  return (
    <div className="section text-center align-items-center">
      <h1>Something went wrong</h1>
      <p>{error}</p>
    </div>
  )
}

export default SomethingWentWrong
