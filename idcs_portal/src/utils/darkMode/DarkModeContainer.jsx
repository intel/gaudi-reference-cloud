// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useEffect } from 'react'
import useDarkModeStore from '../../store/darkModeStore/DarkModeStore'

const DarkModeContainer = () => {
  // Global State
  const isDarkMode = useDarkModeStore((state) => state.isDarkMode)
  const setDarkMode = useDarkModeStore((state) => state.setDarkMode)

  useEffect(() => {
    const isDarkMode = window.matchMedia('(prefers-color-scheme: dark)')
    const detectionDiv = document.querySelector('#detection')
    const isAutoDark = getComputedStyle(detectionDiv).backgroundColor !== 'rgb(255, 255, 255)'

    if (isDarkMode.matches || isAutoDark) {
      setDarkMode(true)
    } else {
      setDarkMode(false)
    }

    // This callback will fire if the perferred color scheme changes without a reload
    isDarkMode.addEventListener('change', (evt) => setDarkMode(evt.matches))
  }, [])

  useEffect(() => {
    if (isDarkMode) {
      document.body.setAttribute('data-bs-theme', 'dark')
    } else {
      document.body.setAttribute('data-bs-theme', 'light')
    }
  }, [isDarkMode])

  return <></>
}

export default DarkModeContainer
