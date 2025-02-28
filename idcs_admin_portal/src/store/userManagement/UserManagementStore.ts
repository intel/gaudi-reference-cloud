import { create } from 'zustand'
import moment from 'moment'

import type UserModel from '../models/userManagement/UserModel'
import CloudAccountService from '../../services/CloudAccountService'

const dateFormat = 'MM/DD/YYYY hh:mm a'

interface UserManagementStore {
  userList: UserModel[] | null
  loading: boolean
  setUserList: () => Promise<void>
}

const useUserManagementStore = create<UserManagementStore>()((set) => ({
  userList: [],
  loading: false,
  setUserList: async () => {
    set({ loading: true })
    // eslint-disable-next-line @typescript-eslint/await-thenable
    const { data } = await CloudAccountService.getCloudAccounts()
    const newData = parseJson(data)

    set({ userList: buildUsersResponse(newData) })
    set({ loading: false })
  }
}))

const buildUsersResponse = (data: any): UserModel[] => {
  return data?.map((result: any) => {
    const item = result.result
    return {
      id: item?.id ? item.id : '',
      name: item?.name ? item.name : '',
      owner: item?.owner ? item.owner : '',
      type: item?.type ? item.type : '',
      created: item?.created ? moment(item.created).format(dateFormat) : '',
      restricted: item?.restricted ? 'Yes' : 'No',
      adminName: item?.adminName ? item?.adminName : '',
      accessLimitedTimestamp: item?.accessLimitedTimestamp && item?.adminName !== '' ? moment(item.accessLimitedTimestamp).format(dateFormat) : ''
    }
  })
}

const parseJson = (data: any): any => {
  data = data.replace('\n', '', 'g')

  let start = data.indexOf('{')
  let open = 0
  let i: number = start
  const len = data.length
  const result = []

  for (; i < len; i++) {
    if (data[i] === '{') {
      open++
    } else if (data[i] === '}') {
      open--
      if (open === 0) {
        result.push(JSON.parse(data.substring(start, i + 1)))
        start = i + 1
      }
    }
  }

  return result
}

export default useUserManagementStore
