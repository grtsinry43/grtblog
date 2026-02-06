import { merge } from 'lodash-es'
import { darkTheme, lightTheme } from 'naive-ui'
import { computed } from 'vue'

import { toRefsPreferencesStore } from '@/stores/preferences'
import { commonThemeOverrides } from '@/theme/common'
import { baseDarkThemeOverrides } from '@/theme/dark'
import { baseLightThemeOverrides } from '@/theme/light'
import { cah } from '@/utils/chromaHelper'

import type { GlobalThemeOverrides } from 'naive-ui'

function applyGlassEffect(overrides: GlobalThemeOverrides, glassOpacity: number): GlobalThemeOverrides {
  const alpha = glassOpacity / 100
  const cardColor = overrides.common?.cardColor
  if (cardColor) {
    const transparentCardColor = cah(cardColor, alpha)
    overrides.common!.cardColor = transparentCardColor
  }
  return overrides
}

export function useTheme() {
  const { themeColor, isDark, backgroundImage } = toRefsPreferencesStore()

  const isGlassActive = computed(() =>
    backgroundImage.value.show && backgroundImage.value.url && backgroundImage.value.glassEffect.enable,
  )

  const getLightThemeOverrides = (primaryColor = themeColor.value) => {
    return merge(commonThemeOverrides(primaryColor), baseLightThemeOverrides(primaryColor))
  }

  const getDarkThemeOverrides = (primaryColor = themeColor.value) => {
    return merge(commonThemeOverrides(primaryColor), baseDarkThemeOverrides(primaryColor))
  }

  const themeOverrides = computed(() => {
    const overrides = isDark.value
      ? getDarkThemeOverrides(themeColor.value)
      : getLightThemeOverrides(themeColor.value)

    if (isGlassActive.value) {
      return applyGlassEffect(overrides, backgroundImage.value.glassEffect.opacity)
    }

    return overrides
  })

  const theme = computed(() => {
    return isDark.value ? darkTheme : lightTheme
  })

  return {
    theme,
    themeOverrides,
    getLightThemeOverrides,
    getDarkThemeOverrides,
  }
}
