// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'
import CloudAccountService from '../../services/CloudAccountService'
import useProductStore, { type Product } from '../productStore/ProductStore'
import useUserStore from '../userStore/UserStore'
import idcConfig from '../../config/configurator'

export interface SshKeys {
  resourceId: string
  name: string
  type: 'rsa'
  value: string
  cloudAccountId: string
  sshPublicKey: string
  ownerEmail: string
  allowDelete?: boolean
  trustedCustomer?: boolean
  email?: string
  instances?: string
}

export interface Vnet {
  name: string
  availabilityZone: string
  prefixLength: string
  region: string
}

export interface Interfaces {
  name: string
  vNet: string
  dnsName: string
  addresses: string[]
  prefixLength: string
  subnet: string
  gateway: string
}

export interface Reservation {
  cloudAccountId: string
  name: string
  instanceGroup: string
  resourceId: string
  resourceVersion: string
  availabilityZone: string
  instanceType: string
  machineImage: string
  runStrategy: string
  sshPublicKey: string[]
  status: string
  userName: string
  message: string
  interfaces: Interfaces[]
  sshProxyUser: string
  sshProxyAddress: string
  sshProxyPort: string
  creationTimestamp: string
  expirationTimestamp?: string
  instanceTypeDetails: Product | null
  sshUrl?: string | null
  nodegroupType?: string | null
  requestedExtensionDays?: number
  labels: instanceLabels
  quickConnectEnabled?: boolean
  objectStorage?: boolean
  primaryOwnerEmail?: string
  secondaryOwnerEmail?: string
}

export type instanceLabels = Record<string, string>

export interface GroupReservation {
  cloudAccountId: string
  name: string
  instanceCount: number
  readyCount: number
  availabilityZone: string
  instanceType: string
  machineImage: string
  runStrategy: string
  sshPublicKey: string[]
  interfaces: Interfaces[]
  instanceTypeDetails: Product | null
  quickConnectEnabled: boolean
}

export interface StorageReservation {
  cloudAccountId: string
  name: string
  description: string
  resourceId: string
  encrypted: boolean
  accessModes: string
  availabilityZone: string
  mountProtocol: string
  storage: string
  size: string
  storageClass: string
  status: string
  message: string
  mountClusterName: string
  mountClusterVersion: string
  mountFilesystemName: string
  mountNamespace: string
  creationTimestamp: string
  user: string
  password: string
  securityGroup: SecurityGroup[] | []
}

export interface SecurityGroup {
  gateway: string
  prefixLength: string
  subnet: string
}

export interface CloudAccount {
  publicKeys: SshKeys[] | []
  instances: Reservation[] | []
  instanceGroups: GroupReservation[] | []
  instanceGroupInstances: Reservation[] | []
  storages: StorageReservation[] | []
  instanceTypeSelected: Product | null
  loading: boolean
  shouldRefreshInstances: boolean
  shouldRefreshInstanceGroups: boolean
  shouldRefreshStorages: boolean
  vnets: Vnet[] | []
  setPublickeys: () => Promise<void>
  setInstances: (isBackground: boolean) => Promise<void>
  setInstanceGroups: (isBackground: boolean) => Promise<void>
  setInstanceGroupInstances: (instanceGroup?: string) => Promise<void>
  setStorages: (isBackground: boolean) => Promise<void>
  setInstanceTypeSelected: (name: string) => void
  setVnets: () => Promise<void>
  setShouldRefreshInstances: (value: boolean) => void
  setShouldRefreshInstanceGroups: (value: boolean) => void
  setShouldRefreshStorages: (value: boolean) => void
  generateCloudConnection: (resourceId: string) => Promise<void>
  setInstancesOptionsList: () => Promise<any>
  setStorageOptionsList: () => Promise<any>
  reset: () => void
}

const BuildReservations = async (instancesResponse: any): Promise<Reservation[]> => {
  const instances: Reservation[] = []

  if (!instancesResponse.data.items) {
    return instances
  }

  for (const index in instancesResponse.data.items) {
    const instanceItem = { ...instancesResponse.data.items[index] }

    const interfaces: Interfaces[] = []

    for (const indexInteface in instanceItem.status.interfaces) {
      const interfaceItem = {
        ...instanceItem.status.interfaces[indexInteface]
      }

      const newInterface: Interfaces = {
        name: interfaceItem.name,
        vNet: interfaceItem.vNet,
        dnsName: interfaceItem.dnsName,
        gateway: interfaceItem.gateway,
        prefixLength: interfaceItem.prefixLength,
        addresses: [...interfaceItem.addresses],
        subnet: interfaceItem.subnet
      }

      interfaces.push(newInterface)
    }

    const products = useProductStore.getState().products

    if (products.length === 0) {
      await useProductStore.getState().setProducts()
    }

    const instanceType = useProductStore.getState().getProductByName(instanceItem.spec.instanceType)
    const sshProxy = { ...instanceItem.status.sshProxy }

    const labels = instanceItem.metadata.labels
    const nodegroupType = labels?.nodegroupType ? labels.nodegroupType : ''

    const instance: Reservation = {
      cloudAccountId: instanceItem.metadata.cloudAccountId,
      name: instanceItem.metadata.name,
      instanceGroup: instanceItem.spec.instanceGroup,
      creationTimestamp: instanceItem.metadata.creationTimestamp,
      resourceId: instanceItem.metadata.resourceId,
      resourceVersion: instanceItem.metadata.resourceVersion,
      availabilityZone: instanceItem.spec.availabilityZone,
      instanceType: instanceItem.spec.instanceType,
      machineImage: instanceItem.spec.machineImage,
      runStrategy: instanceItem.spec.runStrategy,
      sshPublicKey: instanceItem.spec.sshPublicKeyNames,
      quickConnectEnabled: instanceItem.spec.quickConnectEnabled === 'True',
      status: instanceItem.status.phase,
      userName: instanceItem.status.userName,
      message: instanceItem.status.message,
      interfaces,
      instanceTypeDetails: instanceType,
      sshProxyUser: sshProxy.proxyUser,
      sshProxyPort: sshProxy.proxyPort,
      sshProxyAddress: sshProxy.proxyAddress,
      nodegroupType,
      labels: instanceItem.metadata.labels
    }

    instances.push(instance)
  }

  instances.sort((p1, p2) => (p1.name < p2.name ? 1 : p1.name > p2.name ? -1 : 0))

  return instances
}

const getSshKeyOwnerEmail = (sshKeyRawData: any): string => {
  if (sshKeyRawData.spec.ownerEmail) {
    return sshKeyRawData.spec.ownerEmail
  } else {
    const currentUser = useUserStore.getState().user
    if (currentUser === null) {
      return ''
    }
    const isOwnCloudAccount = useUserStore.getState().isOwnCloudAccount
    return isOwnCloudAccount ? currentUser.email : currentUser.accountOwnerEmail
  }
}

const initialState = {
  publicKeys: [],
  instances: [],
  storages: [],
  instanceGroups: [],
  instanceGroupInstances: [],
  instanceTypeSelected: null,
  loading: false,
  shouldRefreshInstances: false,
  shouldRefreshInstanceGroups: false,
  shouldRefreshStorages: false,
  vnets: [],
  labels: {}
}

const useCloudAccountStore = create<CloudAccount>()((set, get) => ({
  ...initialState,
  reset: () => {
    set(initialState)
  },
  setPublickeys: async () => {
    set({ loading: true })
    const sshResponse = await CloudAccountService.getSshByCloud()

    const publicKeys: SshKeys[] = []

    for (const index in sshResponse.data.items) {
      const sshItem = { ...sshResponse.data.items[index] }

      const publicKey: SshKeys = {
        cloudAccountId: sshItem.metadata.cloudAccountId,
        resourceId: sshItem.metadata.resourceId,
        type: 'rsa',
        name: sshItem.metadata.name,
        value: sshItem.metadata.name,
        sshPublicKey: sshItem.spec.sshPublicKey,
        ownerEmail: getSshKeyOwnerEmail(sshItem),
        allowDelete: sshItem.metadata.allowDelete
      }

      publicKeys.push(publicKey)
    }

    publicKeys.sort((p1, p2) => (p1.name < p2.name ? 1 : p1.name > p2.name ? -1 : 0))

    set({ loading: false, publicKeys })
  },
  setStorages: async (isBackground = false) => {
    try {
      if (!isBackground) {
        set({ loading: true })
      }

      const storagesResponse = await CloudAccountService.getStoragesByCloudAccount()
      const storages: StorageReservation[] = []

      const getStorageStatusPhase = (status: any): string => {
        if (!status.phase) {
          return ''
        }
        const phase: string = status.phase.replace('FS', '')
        return `${phase.charAt(0)}${phase.substring(1)}`
      }

      const getStorageGBSize = (spec: any): string => {
        return spec.request.storage.replace(/\D/g, '')
      }

      for (const index in storagesResponse.data.items) {
        const storageItem = { ...storagesResponse.data.items[index] }

        const { metadata, spec, status } = storageItem

        const securityGroup = { ...status.securityGroup }

        const securityGroupItems: SecurityGroup[] = []

        for (const index in securityGroup.networkFilterAllow) {
          const networkFilterAllow = securityGroup.networkFilterAllow[index]
          securityGroupItems.push({
            gateway: networkFilterAllow.gateway,
            prefixLength: networkFilterAllow.prefixLength,
            subnet: networkFilterAllow.subnet
          })
        }

        const storageReservation: StorageReservation = {
          cloudAccountId: metadata.cloudAccountId,
          name: metadata.name,
          description: metadata.description,
          resourceId: metadata.resourceId,
          encrypted: spec.encrypted,
          accessModes: spec.accessModes,
          availabilityZone: spec.availabilityZone,
          mountProtocol: spec.mountProtocol,
          storage: spec.request.storage,
          size: getStorageGBSize(spec),
          storageClass: spec.storageClass,
          status: getStorageStatusPhase(status),
          message: status.message,
          mountClusterName: status.mount?.clusterName,
          mountClusterVersion: status.mount?.clusterVersion,
          mountFilesystemName: status.mount?.filesystemName,
          mountNamespace: status.mount?.namespace,
          creationTimestamp: metadata.creationTimestamp,
          user: status.mount?.username,
          password: '',
          securityGroup: securityGroupItems
        }
        storages.push(storageReservation)
      }
      storages.sort((p1, p2) => (p1.name < p2.name ? 1 : p1.name > p2.name ? -1 : 0))
      set({ loading: false, storages })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  setInstanceGroups: async (isBackground = false) => {
    try {
      if (!isBackground) {
        set({ loading: true })
      }

      const instanceGroupsResponse = await CloudAccountService.getInstanceGroupsByCloudAccount()

      const instanceGroups: GroupReservation[] = []

      for (const index in instanceGroupsResponse.data.items) {
        const instanceItem = { ...instanceGroupsResponse.data.items[index] }

        const interfaces: Interfaces[] = []

        for (const indexInterface in instanceItem.spec.instanceSpec.interfaces) {
          const interfaceItem = {
            ...instanceItem.spec.instanceSpec.interfaces[indexInterface]
          }

          const newInterface: Interfaces = {
            name: interfaceItem.name,
            vNet: interfaceItem.vNet,
            prefixLength: '',
            dnsName: '',
            gateway: '',
            addresses: [],
            subnet: ''
          }

          interfaces.push(newInterface)
        }

        const products = useProductStore.getState().products

        if (products.length === 0) {
          await useProductStore.getState().setProducts()
        }

        const instanceType = useProductStore.getState().getProductByName(instanceItem.spec.instanceSpec.instanceType)

        const instanceGroup: GroupReservation = {
          cloudAccountId: instanceItem.metadata.cloudAccountId,
          name: instanceItem.metadata.name,
          instanceCount: instanceItem.spec.instanceCount,
          readyCount: instanceItem.status.readyCount,
          availabilityZone: instanceItem.spec.instanceSpec.availabilityZone,
          instanceType: instanceItem.spec.instanceSpec.instanceType,
          machineImage: instanceItem.spec.instanceSpec.machineImage,
          runStrategy: instanceItem.spec.instanceSpec.runStrategy,
          sshPublicKey: instanceItem.spec.instanceSpec.sshPublicKeyNames,
          quickConnectEnabled: instanceItem.spec.instanceSpec.quickConnectEnabled === 'True',
          interfaces,
          instanceTypeDetails: instanceType
        }

        instanceGroups.push(instanceGroup)
      }

      instanceGroups.sort((p1, p2) => (p1.name < p2.name ? 1 : p1.name > p2.name ? -1 : 0))

      set({ loading: false })
      set({ instanceGroups })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  setInstances: async (isBackground = false) => {
    try {
      if (!isBackground) {
        set({ loading: true })
      }

      const instancesResponse = await CloudAccountService.getInstancesByCloudAccount()

      const instances = await BuildReservations(instancesResponse)

      set({ loading: false })
      set({ instances })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  setInstanceGroupInstances: async (instanceGroup?: string) => {
    if (!instanceGroup) {
      set({ instanceGroupInstances: [] })
      return
    }

    const instancesResponse = await CloudAccountService.getInstanceGroupInstances(instanceGroup)

    const instanceGroupInstances = await BuildReservations(instancesResponse)

    set({ instanceGroupInstances })
  },
  setInstanceTypeSelected: (name: string) => {
    let product = null

    if (name !== null) {
      product = useProductStore.getState().getProductByName(name)
    }

    set({ instanceTypeSelected: product })
  },
  setVnets: async () => {
    const response = await CloudAccountService.getMyVnets()

    const vnets: Vnet[] = []

    for (const index in response.data.items) {
      const vnetItem = { ...response.data.items[index] }

      const vnet: Vnet = {
        name: vnetItem.metadata.name,
        availabilityZone: vnetItem.spec.availabilityZone,
        prefixLength: vnetItem.spec.prefixLength,
        region: vnetItem.spec.region
      }

      vnets.push(vnet)
    }

    set({ vnets })
  },
  setShouldRefreshInstances: (value: boolean) => {
    set({ shouldRefreshInstances: value })
  },
  setShouldRefreshInstanceGroups: (value: boolean) => {
    set({ shouldRefreshInstanceGroups: value })
  },
  setShouldRefreshStorages: (value: boolean) => {
    set({ shouldRefreshStorages: value })
  },
  generateCloudConnection: async (resourceId: string) => {
    const currentUser = useUserStore.getState().user
    const cloudAccountId = currentUser?.cloudAccountNumber
    if (cloudAccountId) {
      const cloudConnectUrl = `${idcConfig.REACT_APP_CLOUD_CONNECT_URL.replace('$UUID', resourceId).replace('$CLOUDACCOUNT', cloudAccountId).replace('$UUID', resourceId).replace('$REGION', idcConfig.REACT_APP_SELECTED_REGION)}`
      window.open(cloudConnectUrl, '_blank', 'noreferrer')
    }
  },
  setInstancesOptionsList: async () => {
    if (get().instances.length === 0) {
      await get().setInstances(false)
    }

    return buildInstanceOptionsList(get().instances)
  },
  setStorageOptionsList: async () => {
    if (get().storages.length === 0) {
      await get().setStorages(false)
    }

    return buildStorageOptionsList(get().storages)
  }
}))

const allowedStates = ['Provisioning', 'Stopping', 'Terminating', 'Deleting', 'Starting']

const needToRefreshInstances = (): boolean => {
  const instances = useCloudAccountStore.getState().instances
  const shouldRefreshInstances = useCloudAccountStore.getState().shouldRefreshInstances
  return shouldRefreshInstances && instances?.some((instance: Reservation) => allowedStates.includes(instance.status))
}

const needToRefreshInstanceGroups = (): boolean => {
  const shouldRefreshInstanceGroups = useCloudAccountStore.getState().shouldRefreshInstanceGroups
  const instanceGroups = useCloudAccountStore.getState().instanceGroups
  const instanceGroupInstances = useCloudAccountStore.getState().instanceGroupInstances
  return (
    shouldRefreshInstanceGroups &&
    (instanceGroups?.some((x) => x.readyCount < x.instanceCount) ||
      instanceGroupInstances?.some((instance: Reservation) => allowedStates.includes(instance.status)))
  )
}

const needToRefreshStorages = (): boolean => {
  const storages = useCloudAccountStore.getState().storages
  const shouldRefreshStorages = useCloudAccountStore.getState().shouldRefreshStorages
  return (
    shouldRefreshStorages && storages?.some((storage: StorageReservation) => allowedStates.includes(storage.status))
  )
}

const buildInstanceOptionsList = (instancesResponse: any): any[] => {
  return instancesResponse.map((instance: any) => {
    const instanceTypeDetails = instance.instanceTypeDetails
    const displayName = instanceTypeDetails ? instanceTypeDetails.displayName : ''
    return {
      name: `${instance.name} (${displayName})`,
      value: instance.resourceId
    }
  })
}

const buildStorageOptionsList = (response: any): any[] => {
  return response.map((record: any) => {
    return {
      name: record.name,
      value: record.resourceId
    }
  })
}

setInterval(() => {
  if (needToRefreshInstances()) {
    void (async () => {
      await useCloudAccountStore.getState().setInstances(true)
    })()
  }

  if (needToRefreshInstanceGroups()) {
    void (async () => {
      await useCloudAccountStore.getState().setInstanceGroups(true)
    })()
  }
  if (needToRefreshStorages()) {
    void (async () => {
      await useCloudAccountStore.getState().setStorages(true)
    })()
  }
}, 10000)

export default useCloudAccountStore
