import type { TreeOption } from 'naive-ui'

export interface RouteCatalogTreeOption extends TreeOption {
  children?: RouteCatalogTreeOption[]
  label: string
  routeCount: number
  routePath: string
  terminal: boolean
}

interface MutableRouteNode {
  children: Map<string, MutableRouteNode>
  label: string
  routePath: string
  terminal: boolean
}

export function buildRouteCatalogTree(routes: string[]): RouteCatalogTreeOption[] {
  const root: MutableRouteNode = {
    children: new Map(),
    label: '/',
    routePath: '/',
    terminal: false,
  }

  for (const route of routes) {
    const normalizedRoute = normalizeRoute(route)
    if (!normalizedRoute) continue

    if (normalizedRoute === '/') {
      root.terminal = true
      continue
    }

    let parent = root
    for (const segment of normalizedRoute.slice(1).split('/')) {
      const routePath = parent === root ? `/${segment}` : `${parent.routePath}/${segment}`
      let node = parent.children.get(segment)
      if (!node) {
        node = {
          children: new Map(),
          label: segment,
          routePath,
          terminal: false,
        }
        parent.children.set(segment, node)
      }
      parent = node
    }
    parent.terminal = true
  }

  const options = Array.from(root.children.values(), toTreeOption).sort(compareRouteOptions)
  if (root.terminal) {
    options.unshift({
      isLeaf: true,
      key: '/',
      label: '/',
      routeCount: 1,
      routePath: '/',
      terminal: true,
    })
  }
  return options
}

export function collectBranchKeys(options: RouteCatalogTreeOption[]): Array<string | number> {
  const keys: Array<string | number> = []

  function visit(nodes: RouteCatalogTreeOption[]) {
    for (const node of nodes) {
      if (!node.children?.length) continue
      keys.push(node.key as string | number)
      visit(node.children)
    }
  }

  visit(options)
  return keys
}

export function collectTopLevelBranchKeys(
  options: RouteCatalogTreeOption[],
): Array<string | number> {
  return options
    .filter((option) => option.children?.length)
    .map((option) => option.key as string | number)
}

function normalizeRoute(route: string): string | null {
  const trimmed = route.trim()
  if (!trimmed) return null

  const segments = trimmed.split('/').filter(Boolean)
  return segments.length ? `/${segments.join('/')}` : '/'
}

function toTreeOption(node: MutableRouteNode): RouteCatalogTreeOption {
  const children = Array.from(node.children.values(), toTreeOption).sort(compareRouteOptions)
  const routeCount = children.reduce(
    (total, child) => total + child.routeCount,
    node.terminal ? 1 : 0,
  )

  return {
    children: children.length ? children : undefined,
    isLeaf: children.length === 0,
    key: node.routePath,
    label: node.label,
    routeCount,
    routePath: node.routePath,
    terminal: node.terminal,
  }
}

function compareRouteOptions(left: RouteCatalogTreeOption, right: RouteCatalogTreeOption): number {
  return left.label.localeCompare(right.label, undefined, { numeric: true })
}
