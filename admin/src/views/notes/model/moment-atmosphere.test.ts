import { describe, expect, it } from 'vitest'

import {
  mergeMomentAtmosphere,
  normalizeMomentMood,
  normalizeMomentWeather,
} from './moment-atmosphere'

describe('moment atmosphere metadata', () => {
  it('merges atmosphere without overwriting existing ext info', () => {
    expect(
      mergeMomentAtmosphere(
        {
          images: [{ id: 'cover.webp' }],
          moment: { custom: 'preserved', weather: 'sunny' },
        },
        'rainy',
        'calm',
      ),
    ).toEqual({
      images: [{ id: 'cover.webp' }],
      moment: { custom: 'preserved', weather: 'rainy', mood: 'calm' },
    })
  })

  it('removes only atmosphere fields when both selections are cleared', () => {
    expect(
      mergeMomentAtmosphere(
        { moment: { weather: 'sunny', mood: 'joyful', custom: true } },
        null,
        null,
      ),
    ).toEqual({ moment: { custom: true } })
    expect(mergeMomentAtmosphere({ moment: { weather: 'sunny' } }, null, null)).toBeNull()
  })

  it('normalizes unsupported persisted values to empty selections', () => {
    expect(normalizeMomentWeather('rainy')).toBe('rainy')
    expect(normalizeMomentWeather('storm')).toBeNull()
    expect(normalizeMomentMood('calm')).toBe('calm')
    expect(normalizeMomentMood(1)).toBeNull()
  })
})
