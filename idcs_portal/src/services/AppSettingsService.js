// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

class AppSettingsService {
  lastChangedRegionKey = 'lastChangedRegion'
  defaultCloudAccountKey = 'defaultCloudAccount'

  setDefaultCloudAccount(cloudAccount) {
    sessionStorage.setItem(this.defaultCloudAccountKey, cloudAccount)
  }

  getDefaultCloudAccount() {
    return sessionStorage.getItem(this.defaultCloudAccountKey)
  }

  clearDefaultCloudAccount() {
    sessionStorage.removeItem(this.defaultCloudAccountKey)
  }
}

export default new AppSettingsService()
