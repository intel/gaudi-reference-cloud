// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import useToastStore from '../store/toastStore/ToastStore'

export const useCopy = () => {
  const { showSuccess } = useToastStore((state) => state)

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text)
    showSuccess('Copied to clipboard!')
  }

  return {
    copyToClipboard
  }
}
