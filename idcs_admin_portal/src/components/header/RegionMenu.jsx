import React from 'react'
import { Dropdown, Nav } from 'react-bootstrap'
import { BsGlobe, BsCheck } from 'react-icons/bs'
import idcConfig from '../../config/configurator'
import useAppStore from '../../store/appStore/AppStore'

const RegionMenu = ({ hideLeftMenu }) => {
  const changeRegion = useAppStore((state) => state.changeRegion)

  const getRegions = () => {
    const regions = idcConfig.REACT_APP_DEFAULT_REGIONS
    return regions.map(function (item) {
      const isSelected = item === idcConfig.REACT_APP_SELECTED_REGION
      return (
        <Dropdown.Item
          key={item}
          intc-id={`region-option-${item}`}
          onClick={() => {
            changeRegion(item)
          }}

        >
          <BsCheck className={isSelected ? 'me-1' : 'invisible'} size="16" intc-id={`selectedIcon-${item}`} />
          <span className={isSelected ? 'fw-bold' : ''} aria-label={`Connect to ${item} Region`}> { item } </span>
        </Dropdown.Item>
      )
    })
  }

  const MenuOptions = React.forwardRef(
    ({ children, style, className, 'aria-labelledby': labeledBy }, ref) => {
      return (
        <div
          ref={ref}
          style={style}
          className={className}
        >
          <span className="list-unstyled mb-0">{children}</span>
        </div>
      )
    }
  )

  MenuOptions.displayName = 'RegionMenu'

  return (
    <Dropdown intc-id="regionMenu" as={Nav.Item} className="align-self-center">
      <Dropdown.Toggle
        id="dropdown-header-region-toggle"
        role="combobox"
        variant="simple"
        aria-label="switch region"
        data-bs-toggle="dropdown"
        aria-expanded="false"
        aria-controls="dropdown-header-menu-region"
        className="d-flex align-items-center"
      >
        <div className="d-sm-none">
          <BsGlobe intc-id="regionIcon" title="Region Icon" />
        </div>
        <div className="d-none d-sm-flex">
          <span intc-id="regionLabel" role="note" aria-label={`Selected ${idcConfig.REACT_APP_SELECTED_REGION} region`}>
            {' '}
            {idcConfig.REACT_APP_SELECTED_REGION}{' '}
          </span>
        </div>
      </Dropdown.Toggle>
      <Dropdown.Menu align="end" as={MenuOptions} id="dropdown-header-menu-region" renderOnMount aria-labelledby="dropdown-header-region-toggle">{getRegions()}</Dropdown.Menu>
    </Dropdown>
  )
}

export default RegionMenu
