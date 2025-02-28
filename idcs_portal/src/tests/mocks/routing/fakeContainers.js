// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useLocation } from 'react-router-dom'

/**
 * Array that specify all containers to be mocked that are used on RoutesMain component
 * Please update every time you add a route
 * Used exclusevely for test routes and access to them
 */
const allRoutingContainersToMock = [
  '../../../containers/billing/CloudCreditsContainers',
  '../../../containers/billing/InvoicesContainers',
  '../../../containers/billing/ManageCouponCodeContainers',
  '../../../containers/billing/ManagePaymentMethodsContainer',
  '../../../containers/billing/ManageCreditCardContainers',
  '../../../containers/billing/PremiumContainers',
  '../../../containers/billing/UpgradeAccountContainers',
  '../../../containers/billing/UsagesContainer',
  '../../../containers/billing/paymentMethods/CreditCardResponseContainers',
  '../../../containers/billing/paymentMethods/ManagePaymentContainers',
  '../../../containers/cluster/ClusterAddNodeGroupContainer',
  '../../../containers/cluster/ClusterHomePageContainer',
  '../../../containers/cluster/ClusterMyReservationsContainer',
  '../../../containers/cluster/ClusterNodeGroupContainer',
  '../../../containers/cluster/ClusterReserveContainer',
  '../../../containers/cluster/LoadbalancerReservationsContainer',
  '../../../containers/compute/ComputeEditReservationContainer',
  '../../../containers/compute/ComputeLaunchContainer',
  '../../../containers/compute/ComputeReservationsContainers',
  '../../../containers/compute-groups/ComputeGroupsEditReservationContainer',
  '../../../containers/compute-groups/ComputeGroupsLaunchContainer',
  '../../../containers/compute-groups/ComputeGroupsReservationsContainers',
  '../../../containers/hardwareCatalog/HardwareCatalogContainer',
  '../../../containers/homePage/HomePageContainer',
  '../../../containers/keypairs/ImportKeysConstainer',
  '../../../containers/keypairs/KeyPairsContainer',
  '../../../containers/notifications/NotificationsContainer',
  '../../../containers/profile/ApiKeysContainer',
  '../../../containers/software/SoftwareCatalogContainer',
  '../../../containers/trainingAndWorkshops/TrainingAndWorkshopsContainer',
  '../../../containers/trainingAndWorkshops/TrainingDetailContainer',
  '../../../containers/storage/StorageReservationsContainer',
  '../../../containers/storage/StorageLaunchContainer'
]

const mockContainerComponent = () => {
  const location = useLocation()
  const title = location.pathname === '/' ? 'HomePage' : location.pathname
  return <h1 intc-id={title}>{title}</h1>
}

allRoutingContainersToMock.forEach((route) => {
  jest.mock(route, () => () => {
    return mockContainerComponent()
  })
})

/**
 * Must be imported before the import of Routing component so that import takes the mock containers
 */
const EmptyImport = () => {
  return null
}

export default EmptyImport
