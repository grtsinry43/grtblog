export function checkVersion() {
  const storageKey = 'version'
  const currentVersion = __APP_VERSION__
  const storedVersion = localStorage.getItem(storageKey)

  if (storedVersion !== currentVersion) {
    localStorage.clear()
    sessionStorage.clear()

    localStorage.setItem(storageKey, currentVersion)
    window.location.reload()
  }
}
