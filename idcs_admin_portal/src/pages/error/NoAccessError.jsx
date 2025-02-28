import React from 'react'
import { AppRolesEnum } from '../../utility/Enums'
import idcConfig from '../../config/configurator'

const NoAccessError = ({ allowedRoles }) => {
  const getAGSRoleNameItem = (role) => {
    switch (role) {
      case AppRolesEnum.GlobalAdmin:
        return idcConfig.REACT_APP_AGS_GLOBAL_ADMIN
      case AppRolesEnum.SREAdmin:
        return idcConfig.REACT_APP_AGS_SRE_ADMIN
      case AppRolesEnum.IKSAdmin:
        return idcConfig.REACT_APP_AGS_IKS_ADMIN
      case AppRolesEnum.ComputeAdmin:
        return idcConfig.REACT_APP_AGS_COMPUTE_ADMIN
      case AppRolesEnum.ProductAdmin:
        return idcConfig.REACT_APP_AGS_SLURM_ADMIN
      case AppRolesEnum.SlurmAdmin:
        return idcConfig.react_app_ags
      default:
        return null
    }
  }

  const getRequiredRoles = () => {
    return (
      <ul>
        {allowedRoles?.map((r, i) => (
          <li key={i} style={{ listStyleType: 'none' }}>
            {getAGSRoleNameItem(r)}
          </li>
        ))}
      </ul>
    )
  }

  return (
    <div className="section text-center align-items-center">
      <h1>Intel AI Cloud Admin Console</h1>
      <h2>Unauthorized</h2>
      <br/>
      <p>You need to request one of these roles in AGS:</p>
      {getRequiredRoles()}
    </div>
  )
}

export default NoAccessError
