interface Region {
  name: string
  friendly_name: string
  type: string
  api_dns: string
  availability_zone: string
  subnet: string
  prefix: string
  is_default: boolean
  created_at: string
  updated_at: string
  adminName: string
}

export interface AccountWhitelist {
  cloudaccountId: string
  regionName: string
  adminName: string
  created: string
}

export default Region
