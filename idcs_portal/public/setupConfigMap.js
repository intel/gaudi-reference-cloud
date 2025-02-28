'use strict'

const loadConfig = () => {
  const xobj = new XMLHttpRequest()
  xobj.overrideMimeType('application/json')
  xobj.open('GET', `/configMap.json?${Date.now()}`, false)
  xobj.onreadystatechange = function () {
    if (xobj.readyState === 4 && xobj.status === 200) {
      window._env_ = JSON.parse(xobj.responseText)
      loadSiteBanners()
      loadDocData()
    }
  }
  xobj.send(null)
}

const loadSiteBanners = () => {
  try {
    const xobj = new XMLHttpRequest()
    xobj.overrideMimeType('application/json')
    xobj.open('GET', `/banners/siteBannerMessages.json?${Date.now()}`, false)
    xobj.onreadystatechange = function () {
      if (xobj.readyState === 4 && xobj.status === 200) {
        const siteBannersObj = JSON.parse(xobj.responseText)
        window._env_.REACT_APP_SITE_BANNERS = siteBannersObj.REACT_APP_SITE_BANNERS
      }
    }
    xobj.send(null)
  } catch (error) {
    console.error('Cannot load site banners')
    window._env_.REACT_APP_SITE_BANNERS = []
  }
}

const loadDocData = () => {
  try {
    const xobj = new XMLHttpRequest()
    xobj.overrideMimeType('application/json')
    xobj.open('GET', `/docs/docData.json?${Date.now()}`, false)
    xobj.onreadystatechange = function () {
      if (xobj.readyState === 4 && xobj.status === 200) {
        const docDataObj = JSON.parse(xobj.responseText)
        window._env_.REACT_APP_LEARNING_DOCS = docDataObj
      }
    }
    xobj.send(null)
  } catch (error) {
    console.error('Cannot load doc data')
    window._env_.REACT_APP_LEARNING_DOCS = []
  }
}

loadConfig()
