// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import Form from 'react-bootstrap/Form'
import Wrapper from '../Wrapper'
import { BsMoonStars } from 'react-icons/bs'
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
    <div className="px-2 px-xl-3 d-flex flex-row">
      <div>
        <BsMoonStars className="me-1" /> Dark Theme
      </div>
      <div className="text-end">
        <Form.Check type="switch" checked={isChecked} onChange={handleToggle} />
      </div>
    </div>
  )

  return <Wrapper>{darkModeSwitch}</Wrapper>
}

export default DarkModeContainerSwitch
