declare global {
  interface Window {
    __GRTBLOG_RUNTIME_CONFIG__?: {
      FEDERATION_ENABLED?: boolean
    }
  }
}

export const isFederationEnabled: boolean =
  window.__GRTBLOG_RUNTIME_CONFIG__?.FEDERATION_ENABLED === true
