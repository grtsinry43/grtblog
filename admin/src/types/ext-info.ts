export interface ImageExtInfoItem {
  id: string
  width?: number
  height?: number
  color?: string
}

export interface ContentExtInfo {
  images?: ImageExtInfoItem[]
  is_year_summary?: number
  [key: string]: unknown
}
