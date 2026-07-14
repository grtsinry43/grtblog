export interface ImageExtInfoItem {
  id: string
  width?: number
  height?: number
  color?: string
}

export type MomentWeather = 'sunny' | 'cloudy' | 'overcast' | 'rainy' | 'snowy' | 'windy' | 'foggy'

export type MomentMood = 'joyful' | 'calm' | 'excited' | 'tired' | 'sad'

export interface MomentExtInfo {
  weather?: MomentWeather
  mood?: MomentMood
  [key: string]: unknown
}

export interface ContentExtInfo {
  images?: ImageExtInfoItem[]
  moment?: MomentExtInfo
  is_year_summary?: number
  _federation_delivery_registry?: {
    mentions?: Record<string, string>
    citations?: Record<string, string>
  }
  [key: string]: unknown
}
