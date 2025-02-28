// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

export async function makePDF(content, invoiceId) {
  const printWindow = window.open()
  printWindow.document.write(content)
  printWindow.document.title = 'Invoice' + invoiceId
  printWindow.document.close()
}

export default makePDF
