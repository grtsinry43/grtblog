import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "GrtBlog",
  description: "GrtBlog v2 — 现代化静态博客系统文档",
  cleanUrls: true,
  lang: 'zh-CN',
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    nav: [
      { text: '首页', link: '/' },
      { text: '使用指南', link: '/guide/introduction' },
      { text: '开发文档', link: '/dev/architecture' },
    ],

    sidebar: {
      '/guide/': [
        {
          text: '使用指南',
          items: [
            { text: '什么是 GrtBlog？', link: '/guide/introduction' },
            { text: '快速部署', link: '/guide/deployment' },
            { text: '写作指南', link: '/guide/writing' },
            { text: '个性化配置', link: '/guide/configuration' },
          ]
        }
      ],
      '/dev/': [
        {
          text: '开发文档',
          items: [
            { text: '架构总览', link: '/dev/architecture' },
            { text: '本地开发', link: '/dev/getting-started' },
            { text: '后端架构', link: '/dev/backend' },
            { text: '前端架构', link: '/dev/frontend' },
            { text: 'svatoms 数据流', link: '/dev/svatoms' },
            { text: 'svmarkdown 渲染', link: '/dev/svmarkdown' },
            { text: '管理后台', link: '/dev/admin' },
            { text: '贡献指南', link: '/dev/contributing' },
          ]
        }
      ],
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/grtsinry43/grtblog-v2' }
    ],

    outline: {
      label: '目录',
    },

    docFooter: {
      prev: '上一页',
      next: '下一页',
    },
  }
})
