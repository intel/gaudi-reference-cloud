const loadWap = () => {
    const shouldDisableWap = window.location.host.startsWith('localhost') || window.location.host.startsWith('127.0.0.1')
    if (shouldDisableWap) {
      return
    }
    
    // Configure TMS settings
    window.wapProfile = 'profile-microsite' // This is mapped by WAP authorize value
    window.wapLocalCode = 'us-en' // Dynamically set per localized site, see mapping table for values
    window.wapSection = 'intel-dev-cloud' // WAP team will give you a unique section for your site
    window.wapEnv = 'prod' // environment to be use in Adobe Tags.
    // eslint-disable-next-line no-unused-vars, no-var
    var wapSinglePage = false // Include this variable only if your site is a single page application, such as one developed with the React framework
    // Load TMS
    const url = 'https://www.intel.com/content/dam/www/global/wap/main/wap-microsite.js'
    const po = document.createElement('script')
    po.type = 'text/javascript'
    po.async = true
    po.src = url
    const s = document.getElementsByTagName('script')[0]
    s.parentNode.insertBefore(po, s)
  }