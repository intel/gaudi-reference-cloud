// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'

interface DarkModeStore {
  isDarkMode: boolean
  isChecked: boolean
  setDarkMode: (status: boolean) => void
}

const useDarkModeStore = create<DarkModeStore>()((set) => ({
  isDarkMode: false,
  isChecked: false,
  setDarkMode: (status: boolean = false) => {
    const isDarkModeLocalStorage = localStorage.getItem('dark-mode')
    const darkModeStatus = isDarkModeLocalStorage === null ? status : isDarkModeLocalStorage === 'true'
    set({ isDarkMode: darkModeStatus })
    set({ isChecked: darkModeStatus })
  }
}))

export default useDarkModeStore
