// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'
import idcConfig from '../../config/configurator'
import { v4 as uuid } from 'uuid'

export interface ToastDefinition {
  id: string
  variant: string
  bodyMessage: string
  autohide: boolean
  delay: number
}

interface ToastStore {
  toasts: ToastDefinition[]
  showSuccess: (message: string, keepAlive: boolean) => void
  showWarning: (message: string, keepAlive: boolean) => void
  showInfo: (message: string, keepAlive: boolean) => void
  showError: (message: string, keepAlive: boolean) => void
  removeToast: (id: string) => void
}

const addToast = (
  toasts: ToastDefinition[],
  variant: string,
  message: string,
  keepAlive: boolean
): ToastDefinition[] => {
  const newToast: ToastDefinition = {
    id: uuid(),
    variant,
    bodyMessage: message,
    delay: variant === 'danger' ? idcConfig.REACT_APP_TOAST_ERROR_DELAY : idcConfig.REACT_APP_TOAST_DELAY,
    autohide: !keepAlive
  }
  toasts.push(newToast)
  return [...toasts]
}

const useToastStore = create<ToastStore>()((set, get) => ({
  toasts: [],
  showSuccess: (message: string, keepAlive: boolean = false) => {
    const toasts = get().toasts
    set({ toasts: addToast(toasts, 'success', message, keepAlive) })
  },
  showWarning: (message: string, keepAlive: boolean = false) => {
    const toasts = get().toasts
    set({ toasts: addToast(toasts, 'warning', message, keepAlive) })
  },
  showInfo: (message: string, keepAlive: boolean = false) => {
    const toasts = get().toasts
    set({ toasts: addToast(toasts, 'info', message, keepAlive) })
  },
  showError: (message: string, keepAlive: boolean = false) => {
    const toasts = get().toasts
    set({ toasts: addToast(toasts, 'danger', message, keepAlive) })
  },
  removeToast: (id: string) => {
    const toasts = get().toasts.filter((x) => x.id !== id)
    set({ toasts: [...toasts] })
  }
}))

export default useToastStore
