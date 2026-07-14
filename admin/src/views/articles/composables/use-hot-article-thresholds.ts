import { onMounted, shallowRef } from 'vue'

import { listSysConfigs } from '@/services/sysconfig'

import {
  formatHotArticleThresholds,
  resolveHotArticleThresholds,
} from '../model/hot-article-thresholds'

const HOT_ARTICLE_CONFIG_KEYS = ['article.hot.views', 'article.hot.likes', 'article.hot.comments']

export function useHotArticleThresholds() {
  const description = shallowRef('正在读取热门文章配置…')

  async function load() {
    try {
      const tree = await listSysConfigs(HOT_ARTICLE_CONFIG_KEYS)
      description.value = formatHotArticleThresholds(resolveHotArticleThresholds(tree))
    } catch {
      description.value = '当前文章已满足后台配置的热门标准'
    }
  }

  onMounted(load)

  return { description }
}
