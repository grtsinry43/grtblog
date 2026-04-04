/**
 * Upgrade Guide Registry
 *
 * To add a guide for a new release:
 *   1. Append the version string to `allUpgradeGuideVersions` in
 *      server/internal/app/setupstate/service.go
 *   2. Add an UpgradeGuideVersion entry below with features and config keys.
 *
 * The init wizard and upgrade-guide page both read from this registry
 * and render the UI dynamically — no extra components needed.
 */

export interface UpgradeGuideFeatureConfig {
  /** sys_config key, e.g. "federation.enabled" */
  key: string
  /** Which config endpoint to call */
  endpoint: 'federation' | 'activitypub' | 'sysconfig'
  /** Value to write when the feature is toggled on */
  enableValue: unknown
}

export interface UpgradeGuideFeature {
  /** Unique id within this guide version */
  id: string
  /** Iconify class name */
  icon: string
  /** Display label */
  label: string
  /** Short description shown below the label */
  description: string
  /** Config keys to set when the user enables this feature */
  configs: UpgradeGuideFeatureConfig[]
  /**
   * If set, also write the site's public_url into this config key
   * when the feature is enabled (useful for instanceURL fields).
   */
  autoFillInstanceURL?: {
    key: string
    endpoint: 'federation' | 'activitypub' | 'sysconfig'
  }
}

export interface UpgradeGuideVersion {
  /** Must match the version string registered on the backend */
  version: string
  /** Tag shown above the title, e.g. "v2.1 新功能" */
  tag: string
  /** Section heading */
  title: string
  /** Paragraph below the heading */
  description: string
  /** Info alert text shown below the feature list */
  hint: string
  /** Toggleable features */
  features: UpgradeGuideFeature[]
}

// ---------------------------------------------------------------------------
// REGISTRY — add new version entries here
// ---------------------------------------------------------------------------

export const upgradeGuideRegistry: UpgradeGuideVersion[] = [
  {
    version: '2.1',
    tag: 'v2.1 新功能',
    title: '联合与互联',
    description: '选择是否为您的站点启用联合功能。这些选项可以随时在设置中更改。',
    hint: '联合功能开启后，系统会自动生成签名密钥，更多高级选项可在「设置 > 联合」中配置。遥测功能可在「设置 > 遥测」中随时查看数据详情或关闭。所有选项均可在设置中随时更改。',
    features: [
      {
        id: 'federation',
        icon: 'ph--circles-three',
        label: '站点联合',
        description:
          '启用后，您的站点可以与其他博客实例或支持联合协议的博客系统建立连接，互相交换友链申请、文章引用和提及通知。',
        configs: [{ key: 'federation.enabled', endpoint: 'federation', enableValue: true }],
        autoFillInstanceURL: { key: 'federation.instanceURL', endpoint: 'federation' },
      },
      {
        id: 'activitypub',
        icon: 'ph--broadcast',
        label: 'ActivityPub',
        description:
          '启用后，Mastodon、Misskey 等 Fediverse 平台的用户可以直接搜索并关注您的站点，新文章和手记将自动推送到他们的时间线。',
        configs: [{ key: 'activitypub.enabled', endpoint: 'activitypub', enableValue: true }],
        autoFillInstanceURL: { key: 'activitypub.instanceURL', endpoint: 'activitypub' },
      },
      {
        id: 'telemetry',
        icon: 'ph--heartbeat',
        label: '帮助我们变得更好',
        description:
          '匿名发送脱敏后的错误摘要和基础运行指标，帮助开发团队更快发现并修复问题。我们承诺：不收集任何个人信息、文章内容或访客数据，您可以随时在设置中查看将要发送的完整数据并关闭此功能。GrtBlog 是开源项目，遥测相关的所有代码均可在 GitHub 上查看、审计和提出问题。',
        configs: [{ key: 'telemetry.enabled', endpoint: 'sysconfig', enableValue: true }],
      },
    ],
  },

  // -- Future example (uncomment and fill in when 2.2 ships) ----------------
  // {
  //   version: '2.2',
  //   tag: 'v2.2 新功能',
  //   title: 'AI 辅助写作',
  //   description: '...',
  //   hint: '...',
  //   features: [
  //     {
  //       id: 'ai-summary',
  //       icon: 'ph--brain',
  //       label: '自动摘要',
  //       description: '...',
  //       configs: [{ key: 'ai.autoSummary.enabled', endpoint: 'sysconfig', enableValue: true }],
  //     },
  //   ],
  // },
]

/** Look up a single version entry. */
export function getGuideByVersion(version: string): UpgradeGuideVersion | undefined {
  return upgradeGuideRegistry.find((g) => g.version === version)
}

/** Return guide entries for the given pending version list, in registry order. */
export function getPendingGuides(pendingVersions: string[]): UpgradeGuideVersion[] {
  const set = new Set(pendingVersions)
  return upgradeGuideRegistry.filter((g) => set.has(g.version))
}

/** Return ALL guide entries — used by the init wizard to show every feature. */
export function getAllGuides(): UpgradeGuideVersion[] {
  return upgradeGuideRegistry
}
