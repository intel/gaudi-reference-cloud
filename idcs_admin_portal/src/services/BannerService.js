import { fromWebToken } from '@aws-sdk/credential-providers'
import { S3Client, GetObjectCommand, PutObjectCommand } from '@aws-sdk/client-s3'
import idcConfig from '../config/configurator'
import { getAccessToken, parseJwt } from '../utility/axios/AxiosInstance'

const BANNER_FILE = 'banners/siteBannerMessages.json'
let s3Client = null
let accessToken = ''

const initializeS3Connection = async () => {
  const decodedJwt = parseJwt(accessToken)
  const timeStampInSeconds = Math.floor(Date.now() / 1000)
  if (!s3Client || !decodedJwt || decodedJwt?.exp < timeStampInSeconds) {
    accessToken = await getAccessToken()
    const credentials = fromWebToken({
      roleArn: idcConfig.REACT_APP_BANNER_AWS_ROLE_ARN,
      webIdentityToken: accessToken
    })

    s3Client = new S3Client({
      region: idcConfig.REACT_APP_AWS_REGION,
      credentials
    })
  }
}

const bannerFileLoading = async () => {
  await initializeS3Connection()
  const getObjCmd = new GetObjectCommand({
    Bucket: idcConfig.REACT_APP_AWS_S3_BUCKET_NAME,
    Key: BANNER_FILE,
    ResponseCacheControl: 'no-cache'
  })

  const response = await s3Client.send(getObjCmd)
  return await response.Body.transformToString()
}

const bannerFileUpdating = async (payload) => {
  await initializeS3Connection()
  const uploadData = { REACT_APP_SITE_BANNERS: payload }
  const command = new PutObjectCommand({
    Bucket: idcConfig.REACT_APP_AWS_S3_BUCKET_NAME,
    Key: BANNER_FILE,
    ContentType: 'application/json',
    Body: JSON.stringify(uploadData)
  })
  await s3Client.send(command)
}

class BannerService {
  async getBanners() {
    try {
      const response = await bannerFileLoading()
      const siteBannersObj = JSON.parse(response)
      return siteBannersObj.REACT_APP_SITE_BANNERS || []
    } catch (error) {
      console.error(error)
      throw Error(error)
    }
  }

  async postBanner(payload, banners) {
    try {
      const updatedData = [...banners]
      updatedData.unshift(payload)
      await bannerFileUpdating(updatedData)
    } catch (error) {
      console.error(error)
      throw Error(error)
    }
  }

  async updateBanner(payload, banners) {
    try {
      const updatedData = [...banners]
      const idx = updatedData.findIndex((banner) => banner.id === payload.id)
      if (idx > -1) {
        updatedData[idx] = payload
        await bannerFileUpdating(updatedData)
      } else throw Error(`Unable to update the banner with ID: ${payload.id}. Banner not found!`)
    } catch (error) {
      console.error(error)
      throw Error(error)
    }
  }

  async removeBanner(id, banners) {
    try {
      const updatedData = [...banners]
      const idx = updatedData.findIndex((banner) => banner.id === id)
      if (idx > -1) {
        updatedData.splice(idx, 1)
        await bannerFileUpdating(updatedData)
      } else throw Error(`Unable to delete the banner with ID: ${id}. Banner not found!`)
    } catch (error) {
      console.error(error)
      throw Error(error)
    }
  }
}

export default new BannerService()
