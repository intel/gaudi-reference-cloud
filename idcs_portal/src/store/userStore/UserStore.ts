// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'
import { EnrollAccountType, EnrollActionResponse, InvitationStateSelection } from '../../utils/Enums'
import CloudAccountService from '../../services/CloudAccountService'
import AppSettingsService from '../../services/AppSettingsService'
import { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'
export interface User {
  displayName: string
  firstName: string
  lastName: string
  email: string
  accountOwnerEmail: string
  countryCode: string
  enterpriseId: string
  ownCloudAccountNumber: string
  cloudAccountNumber: string
  ownCloudAccountType: string
  cloudAccountType: string
  groups: string[]
  idp: string
  idToken: string
  authenticated: boolean
  hasInvitations: boolean
}

interface CloudAccountInvitation {
  cloudAccountId: string
  cloudAccountType: string
  invitationState: string
  email: string
  isApproved: boolean
}

interface Invitations {
  expiry: string | null
  invitationState: string | null
  memberEmail: string | null
  note: string | null
}

export interface EnrollResponse {
  action: string
  enrolled: boolean
  haveBillingAccount: boolean
  haveCloudAccount: boolean
  registered: boolean
  cloudAccountEmail: string
}

interface UserStore {
  user: User | null
  setUser: (idTokenClaims: any, idToken: string) => void
  enroll: (isPremium: boolean, acceptTermAndConditions: boolean) => Promise<void>
  getUserRoles: () => string[]
  isStandardUser: () => boolean
  isPremiumUser: () => boolean
  isEnterprisePendingUser: () => boolean
  isEnterpriseUser: () => boolean
  isIntelUser: () => boolean
  isOwnCloudAccount: boolean
  enrollResponse: EnrollResponse | null
  isLogoutInProgress: boolean
  setIsLogoutInProgress: (isInProgress: boolean) => void
  cloudAccounts: CloudAccountInvitation[] | []
  setcloudAccounts: () => Promise<void>
  setSelectedCloudAccount: (cloudAccount: CloudAccountInvitation) => void
  loading: boolean
  adminInvitation: Invitations[] | []
  invitationLoading: boolean
  setInvitations: () => Promise<void>
  shouldShowAccessManagement: () => boolean
  checkMemberAccess: () => Promise<boolean>
}

const useUserStore = create<UserStore>()((set, get) => ({
  user: null,
  enrollResponse: null,
  isOwnCloudAccount: true,
  setUser: (idTokenClaims: any, idToken: string) => {
    const user: User | null = get().user
    if (user?.authenticated) {
      return
    }
    const newUser: User = {
      displayName: idTokenClaims.displayName || '',
      firstName: idTokenClaims.firstName || '',
      lastName: idTokenClaims.lastName || '',
      email: idTokenClaims.email || '',
      accountOwnerEmail: '',
      countryCode: idTokenClaims.countryCode || '',
      enterpriseId: idTokenClaims.enterpriseId || '',
      ownCloudAccountNumber: '',
      cloudAccountNumber: '',
      ownCloudAccountType: '',
      cloudAccountType: '',
      groups: idTokenClaims.groups || [],
      idp: idTokenClaims.idp || '',
      idToken: idToken || '',
      authenticated: false,
      hasInvitations: false
    }
    set(() => ({ user: newUser }))
  },
  enroll: async (shouldEnrollPremium, acceptTermAndConditions) => {
    const user: User | null = get().user
    if (user !== null) {
      const { data } = await CloudAccountService.enrollCloudAccountDetails(shouldEnrollPremium, acceptTermAndConditions)
      const newEnroll: EnrollResponse = {
        action: data.action || '',
        enrolled: data.enrolled || '',
        haveBillingAccount: data.haveBillingAccount || '',
        haveCloudAccount: data.haveCloudAccount || '',
        registered: data.registered || '',
        cloudAccountEmail: data?.cloudAccountEmail || ''
      }
      if (data.action !== EnrollActionResponse.ENROLL_ACTION_RETRY) {
        user.hasInvitations = data.isMember
        user.cloudAccountNumber = data.isMember ? '' : data.cloudAccountId
        user.cloudAccountType = data.isMember ? '' : data.cloudAccountType
        user.accountOwnerEmail = data.isMember ? '' : user.email
        user.ownCloudAccountNumber = data.cloudAccountId
        user.ownCloudAccountType = data.cloudAccountType
        user.authenticated = true
      }
      const isOwnCloudAccount = user.cloudAccountNumber === user.ownCloudAccountNumber
      set(() => ({ enrollResponse: newEnroll, user: { ...user }, isOwnCloudAccount }))
      set({ loading: false })
    }
  },
  getUserRoles: () => {
    const user: User | null = get().user
    return user !== null ? user.groups : []
  },
  isStandardUser: () => {
    const user = get().user
    return user !== null ? user.cloudAccountType === EnrollAccountType.standard.toString() : false
  },
  isPremiumUser: () => {
    const user = get().user
    return user !== null ? user.cloudAccountType === EnrollAccountType.premium.toString() : false
  },
  isEnterprisePendingUser: () => {
    const user = get().user
    return user !== null ? user.cloudAccountType === EnrollAccountType.enterprise_pending.toString() : false
  },
  isEnterpriseUser: () => {
    const user = get().user
    return user !== null ? user.cloudAccountType === EnrollAccountType.enterprise.toString() : false
  },
  isIntelUser: () => {
    const user = get().user
    return user !== null ? user.cloudAccountType === EnrollAccountType.intel.toString() : false
  },
  isLogoutInProgress: false,
  setIsLogoutInProgress: (isInProgress: boolean) => {
    if (isInProgress) {
      AppSettingsService.clearDefaultCloudAccount()
    }
    set(() => ({ isLogoutInProgress: isInProgress }))
  },
  cloudAccounts: [],
  loading: false,
  setcloudAccounts: async () => {
    const user: User | null = get().user

    set({ loading: true })
    const { data } = await CloudAccountService.getUserCloudAccountsList(user?.email)

    const invitations: CloudAccountInvitation[] = []

    for (const index in data.memberAccount) {
      const invite = { ...data.memberAccount[index] }
      const inviteItem: CloudAccountInvitation = {
        cloudAccountId: invite.id,
        cloudAccountType: invite.type,
        invitationState: invite.invitationState,
        email: invite.name,
        isApproved:
          invite.invitationState === InvitationStateSelection.INVITE_STATE_ACCEPTED ||
          invite.invitationState === InvitationStateSelection.INVITE_STATE_UNSPECIFIED
      }
      invitations.push(inviteItem)
    }

    set({ cloudAccounts: invitations })
    set({ loading: false })
  },
  adminInvitation: [],
  invitationLoading: false,
  setInvitations: async () => {
    set({ invitationLoading: true })

    const { data } = await CloudAccountService.multiUserAdminInvitationList()

    set({ adminInvitation: data.invites })
    set({ invitationLoading: false })
  },
  setSelectedCloudAccount: (cloudAccount: CloudAccountInvitation) => {
    const user = get().user as User
    const isOwnCloudAccount = cloudAccount.cloudAccountId === user.ownCloudAccountNumber
    set({
      user: {
        ...user,
        cloudAccountNumber: cloudAccount.cloudAccountId,
        cloudAccountType: cloudAccount.cloudAccountType,
        accountOwnerEmail: cloudAccount.email
      },
      isOwnCloudAccount
    })
  },
  shouldShowAccessManagement: () => {
    const isOwnCloudAccount = get().isOwnCloudAccount
    const isEnterpriseUser = get().isEnterpriseUser()
    const isIntelUser = get().isIntelUser()
    const isPremiumUser = get().isPremiumUser()
    return (
      isOwnCloudAccount &&
      isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_MULTIUSER) &&
      (isPremiumUser || isEnterpriseUser || isIntelUser)
    )
  },
  checkMemberAccess: async () => {
    const user: User | null = get().user
    const isOwnCloudAccount = get().isOwnCloudAccount
    const skipValidation = !user || isOwnCloudAccount || !user.authenticated || !user.accountOwnerEmail

    if (skipValidation) {
      return true
    }

    const { data } = await CloudAccountService.getUserCloudAccountsList(user?.email)

    const invitations: CloudAccountInvitation[] = []

    for (const index in data.memberAccount) {
      const invite = { ...data.memberAccount[index] }
      const inviteItem: CloudAccountInvitation = {
        cloudAccountId: invite.id,
        cloudAccountType: invite.type,
        invitationState: invite.invitationState,
        email: invite.name,
        isApproved:
          invite.invitationState === InvitationStateSelection.INVITE_STATE_ACCEPTED ||
          invite.invitationState === InvitationStateSelection.INVITE_STATE_UNSPECIFIED
      }
      invitations.push(inviteItem)
    }

    const isMemberOf = invitations.some((x) => x.isApproved && x.email === user.accountOwnerEmail)
    return isMemberOf
  }
}))

export default useUserStore
