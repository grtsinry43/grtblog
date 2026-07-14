import { describe, expect, it } from 'vitest'

import {
  DEFAULT_HOT_ARTICLE_THRESHOLDS,
  formatHotArticleThresholds,
  resolveHotArticleThresholds,
} from './hot-article-thresholds'

import type { SysConfigItem, SysConfigTreeResponse } from '@/services/sysconfig'

function configItem(key: string, value: unknown): SysConfigItem {
  return {
    key,
    value,
    valueType: 'number',
    enumOptions: [],
    visibleWhen: [],
    sort: 0,
    meta: {},
    isSensitive: false,
    createdAt: '',
    updatedAt: '',
  }
}

describe('hot article thresholds', () => {
  it('reads current values from root and nested config items', () => {
    const tree: SysConfigTreeResponse = {
      items: [configItem('article.hot.views', '320')],
      groups: [
        {
          key: 'article',
          path: 'article',
          label: '文章',
          children: [
            {
              key: 'hot',
              path: 'article/hot',
              label: '热门文章',
              items: [configItem('article.hot.likes', 24), configItem('article.hot.comments', '8')],
            },
          ],
        },
      ],
    }

    expect(resolveHotArticleThresholds(tree)).toEqual({ views: 320, likes: 24, comments: 8 })
  })

  it('falls back to backend defaults for missing or invalid values', () => {
    expect(
      resolveHotArticleThresholds({
        groups: [],
        items: [configItem('article.hot.views', 'invalid')],
      }),
    ).toEqual(DEFAULT_HOT_ARTICLE_THRESHOLDS)
  })

  it('formats all three configured conditions with the actual inclusive comparison', () => {
    expect(formatHotArticleThresholds({ views: 100, likes: 10, comments: 5 })).toBe(
      '热门标准：浏览量 ≥ 100、点赞数 ≥ 10 或评论数 ≥ 5',
    )
  })
})
