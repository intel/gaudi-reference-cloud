class RegionService {
  lastChangedRegionKey = 'lastChangedRegion'

  saveLastChangedRegion(region) {
    sessionStorage.setItem(this.lastChangedRegionKey, region)
  }

  getLastChangedRegion() {
    const currentSelectedRegion = sessionStorage.getItem(this.lastChangedRegionKey)
    return currentSelectedRegion
  }

  clearLastChangedRegion() {
    sessionStorage.removeItem(this.lastChangedRegionKey)
  }
}

export default new RegionService()
