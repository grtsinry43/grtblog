import type { ContentExtInfo, MomentMood, MomentWeather } from '@/types/ext-info'

interface AtmosphereOption<T extends string> {
  value: T
  label: string
  iconClass: string
}

export const momentWeatherOptions: readonly AtmosphereOption<MomentWeather>[] = [
  { value: 'sunny', label: '晴朗', iconClass: 'ph--sun' },
  { value: 'cloudy', label: '多云', iconClass: 'ph--cloud-sun' },
  { value: 'overcast', label: '阴天', iconClass: 'ph--cloud' },
  { value: 'rainy', label: '下雨', iconClass: 'ph--cloud-rain' },
  { value: 'snowy', label: '下雪', iconClass: 'ph--cloud-snow' },
  { value: 'windy', label: '有风', iconClass: 'ph--wind' },
  { value: 'foggy', label: '有雾', iconClass: 'ph--cloud-fog' },
]

export const momentMoodOptions: readonly AtmosphereOption<MomentMood>[] = [
  { value: 'joyful', label: '开心', iconClass: 'ph--smiley' },
  { value: 'calm', label: '平静', iconClass: 'ph--heart' },
  { value: 'excited', label: '兴奋', iconClass: 'ph--sparkle' },
  { value: 'tired', label: '疲惫', iconClass: 'ph--moon-stars' },
  { value: 'sad', label: '低落', iconClass: 'ph--smiley-sad' },
]

const momentWeatherValues = new Set(momentWeatherOptions.map((option) => option.value))
const momentMoodValues = new Set(momentMoodOptions.map((option) => option.value))

export function normalizeMomentWeather(value: unknown): MomentWeather | null {
  return typeof value === 'string' && momentWeatherValues.has(value as MomentWeather)
    ? (value as MomentWeather)
    : null
}

export function normalizeMomentMood(value: unknown): MomentMood | null {
  return typeof value === 'string' && momentMoodValues.has(value as MomentMood)
    ? (value as MomentMood)
    : null
}

export function mergeMomentAtmosphere(
  base: ContentExtInfo | null | undefined,
  weather: MomentWeather | null,
  mood: MomentMood | null,
): ContentExtInfo | null {
  const next: ContentExtInfo = { ...(base ?? {}) }
  const moment = { ...(next.moment ?? {}) }

  if (weather) moment.weather = weather
  else delete moment.weather

  if (mood) moment.mood = mood
  else delete moment.mood

  if (Object.keys(moment).length) next.moment = moment
  else delete next.moment

  return Object.keys(next).length ? next : null
}
