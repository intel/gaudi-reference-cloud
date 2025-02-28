// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import moment from 'moment'

const mainDateFormat = (date) => {
  return moment(date).format('dd/MM/YYYY [at] hh:mm A')
}

export default mainDateFormat
