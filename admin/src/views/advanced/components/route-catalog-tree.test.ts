import { describe, expect, it } from 'vitest'

import {
  buildRouteCatalogTree,
  collectBranchKeys,
  collectTopLevelBranchKeys,
} from './route-catalog-tree'

describe('route catalog tree', () => {
  it('groups flat routes by path segment and counts terminal routes', () => {
    const tree = buildRouteCatalogTree(['/', '/posts', '/posts/2', '/posts/10', '/moments/1'])

    expect(tree.map((node) => node.routePath)).toEqual(['/', '/moments', '/posts'])
    expect(tree[1]?.routeCount).toBe(1)
    expect(tree[2]?.routeCount).toBe(3)
    expect(tree[2]?.terminal).toBe(true)
    expect(tree[2]?.children?.map((node) => node.label)).toEqual(['2', '10'])
  })

  it('normalizes duplicate separators and ignores duplicate or empty routes', () => {
    const tree = buildRouteCatalogTree([' /posts//hello/ ', '/posts/hello', '', '   '])

    expect(tree).toHaveLength(1)
    expect(tree[0]?.routeCount).toBe(1)
    expect(tree[0]?.children?.[0]?.routePath).toBe('/posts/hello')
  })

  it('collects first-level and all expandable branches', () => {
    const tree = buildRouteCatalogTree(['/posts/archive/2025', '/moments/1'])

    expect(collectTopLevelBranchKeys(tree)).toEqual(['/moments', '/posts'])
    expect(collectBranchKeys(tree)).toEqual(['/moments', '/posts', '/posts/archive'])
  })
})
