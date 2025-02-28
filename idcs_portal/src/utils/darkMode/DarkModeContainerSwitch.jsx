// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import Form from 'react-bootstrap/Form'
import Wrapper from '../Wrapper'
import { BsMoon } from 'react-icons/bs'
import useDarkModeStore from '../../store/darkModeStore/DarkModeStore'

const DarkModeContainerSwitch = (props) => {
  // Global State
  const isChecked = useDarkModeStore((state) => state.isChecked)
  const setDarkMode = useDarkModeStore((state) => state.setDarkMode)

  const handleToggle = () => {
    localStorage.setItem('dark-mode', !isChecked)
    setDarkMode(!isChecked)
  }

  const darkModeSwitch = (
    <div className="d-flex flex-row align-items-center gap-s4 py-s3 px-s6">
      <BsMoon />
      Toggle Dark Mode
      <Form.Check type="switch" checked={isChecked} onChange={handleToggle} />
    </div>
  )

  return <Wrapper>{darkModeSwitch}</Wrapper>
}

export default DarkModeContainerSwitch
