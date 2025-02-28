import React from 'react'
import Navbar from 'react-bootstrap/Navbar'
import { Dropdown, Button, Nav } from 'react-bootstrap'
import { BsBoxArrowRight, BsPerson } from 'react-icons/bs'
import useUserStore from '../../store/userStore/UserStore'
import RegionMenu from './RegionMenu'
import useLogout from '../../hooks/useLogout'
import DarkModeContainerSwitch from '../../utility/darkMode/DarkModeContainerSwitch'
import { useNavigate } from 'react-router'
import TopToolbarContainer from '../../containers/navigation/TopToolbarContainer'

const Header = () => {
  // Navigation
  const navigate = useNavigate()

  const { user } = useUserStore()
  const { logoutHandler } = useLogout()

  const userName = user ? user.displayName : ''
  const email = user ? user.email : ''
  const logout = async() => {
    await logoutHandler('/')
  }

  function backToHome() {
    navigate('/')
  }

  return (
    <>
      <Navbar
        className="w-100 siteNavbar" expand aria-label="Site NavBar"
        fixed="top"
      >
        <div className='h4' onClick={backToHome} style={{ cursor: 'pointer' }}>
          Intel® Tiber™ AI Cloud Admin Console
        </div>
        <Navbar.Collapse className="justify-content-end user-image">
          <Nav>
            <div className="d-flex align-items-start" role="search">
              <RegionMenu />
            </div>

            <Dropdown intc-id="userMenu" className="align-self-center">
              <Dropdown.Toggle
                id="dropdown-header-user-toggle"
                role="combobox"
                className="d-flex align-items-center"
                variant="icon-simple"
                as={Button}
                data-bs-toggle="dropdown"
                aria-controls="dropdown-header-user-menu"
                aria-expanded="false"
                aria-label="User Menu"
              >
                <BsPerson intc-id="userIcon" />
              </Dropdown.Toggle>
              <Dropdown.Menu
                id="dropdown-header-user-menu"
                renderOnMount
                align="end"
                aria-labelledby="dropdown-header-user-toggle"
                style={{ whiteSpace: 'nowrap' }}
              >
                <div className="px-3 py-2 defaultBackground">
                  <div>
                    <span className="h6 mb-0">{userName}</span>
                  </div>
                  <div>
                    <small>{email}</small>
                  </div>
                </div>
                <Dropdown.Divider className="mt-0" />
                <DarkModeContainerSwitch />
                <Dropdown.Item as="button" onClick={logout}>
                  <BsBoxArrowRight/>Sign-out
                </Dropdown.Item>
              </Dropdown.Menu>
            </Dropdown>
          </Nav>
        </Navbar.Collapse>
      </Navbar>
      <TopToolbarContainer />
    </>
  )
}

export default Header
