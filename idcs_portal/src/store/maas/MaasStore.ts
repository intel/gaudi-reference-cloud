// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'
import MaasService from '../../services/MaasService'

export interface ChatItem {
  isResponse: boolean
  text: string
}

export interface MaasStore {
  loading: boolean
  generatedMessage: string
  chat: ChatItem[]
  reset: () => void
  generateStream: (params: any) => Promise<void>
  addChatItem: (chatItem: ChatItem) => void
}

const initialState = {
  generatedMessage: '',
  loading: false,
  chat: []
}

const useMaasStore = create<MaasStore>()((set, get) => ({
  ...initialState,
  reset: () => {
    set(initialState)
  },
  generateStream: async (params: any) => {
    set({ loading: true, generatedMessage: '' })
    try {
      await MaasService.generateStream(params, (chunk: string) => {
        set((state) => ({ generatedMessage: state.generatedMessage + chunk }))
      })
      const state = get()
      const newMessage = state.generatedMessage
      set((state) => ({ chat: [...state.chat, { isResponse: true, text: newMessage }] }))
      set({ generatedMessage: '' })
      set({ loading: false })
    } catch (error) {
      set({ generatedMessage: '' })
      set({ loading: false })
      throw error
    }
  },
  addChatItem: (chatItem: ChatItem) => {
    set((state) => ({ chat: [...state.chat, chatItem] }))
  }
}))

export default useMaasStore
