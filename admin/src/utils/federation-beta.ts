import type { RouteLocationNormalized } from 'vue-router'

type FederationBetaConfirmHandler = () => Promise<boolean>

const FEDERATION_BETA_TITLE = '联合功能 Beta 说明'
const FEDERATION_BETA_CONTENT =
  'Blog Federation 与 ActivityPub 兼容目前仍处于 Beta 阶段。当前版本已开放体验，但在稳定性、兼容性与交互细节上仍会持续改进与修复，预计在 2.1.0 版本趋于稳定。'

let federationBetaConfirmHandler: FederationBetaConfirmHandler | null = null

const FEDERATION_BETA_ACK_KEY = 'grtblog:federation-beta-ack'

export function isFederationBetaRoute(to: RouteLocationNormalized) {
  if (to.name === 'settings' && to.query.tab === 'federation') {
    return true
  }

  return (
    typeof to.name === 'string' &&
    [
      'unionManagement',
      'federationInstances',
      'federationOutbound',
      'activityPubOutbox',
      'federationReviews',
      'federationDebug',
      'unionSettingsLegacy',
      'activityPubSettingsLegacy',
    ].includes(to.name)
  )
}

export function registerFederationBetaConfirmHandler(handler: FederationBetaConfirmHandler | null) {
  federationBetaConfirmHandler = handler
}

export function hasAcknowledgedFederationBeta() {
  if (typeof window === 'undefined') {
    return false
  }

  return window.localStorage.getItem(FEDERATION_BETA_ACK_KEY) === '1'
}

export function markFederationBetaAcknowledged() {
  if (typeof window === 'undefined') {
    return
  }

  window.localStorage.setItem(FEDERATION_BETA_ACK_KEY, '1')
}

export async function showFederationBetaDialog() {
  if (hasAcknowledgedFederationBeta()) {
    return true
  }

  if (!federationBetaConfirmHandler) {
    return true
  }

  return federationBetaConfirmHandler()
}

export const federationBetaTitle = FEDERATION_BETA_TITLE
export const federationBetaMessage = FEDERATION_BETA_CONTENT
