// .vitepress/theme/index.ts
import { h } from 'vue'
import DefaultTheme from 'vitepress/theme'
import './style.css'
import AnnouncementBar from './components/AnnouncementBar.vue'

// Fonts
import '@fontsource/google-sans/index.css';
import '@fontsource/noto-serif-sc/index.css';
import '@fontsource-variable/victor-mono/index.css';

export default {
  extends: DefaultTheme,
  Layout() {
    return h(DefaultTheme.Layout, null, {
      'layout-top': () => h(AnnouncementBar)
    })
  },
  enhanceApp({ app }) {
    // register your custom global components
  }
}