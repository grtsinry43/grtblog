// .vitepress/theme/index.ts
import DefaultTheme from 'vitepress/theme'
import './style.css'

// Fonts
import '@fontsource/google-sans/index.css';
import '@fontsource/noto-serif-sc/index.css';
import '@fontsource-variable/victor-mono/index.css';

export default {
  extends: DefaultTheme,
  enhanceApp({ app }) {
    // register your custom global components
  }
}