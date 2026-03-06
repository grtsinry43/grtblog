export interface ImageExtInfoItem {
  id: string
  width?: number
  height?: number
  color?: string
}

export interface ContentExtInfo {
  images?: ImageExtInfoItem[]
  is_year_summary?: number
  _federation_delivery_registry?: {
    mentions?: Record<string, string>
    citations?: Record<string, string>
  }
  [key: string]: unknown
}
