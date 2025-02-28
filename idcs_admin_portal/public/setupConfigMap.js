function loadConfig() {
  const xobj = new XMLHttpRequest()
  xobj.overrideMimeType('application/json')
  xobj.open('GET', `/configMap.json?${Date.now()}`, false)
  xobj.onreadystatechange = function () {
    if (xobj.readyState === 4 && xobj.status === 200) {
      window._env_ = JSON.parse(xobj.responseText)
    }
  }
  xobj.send(null)
}
loadConfig()
