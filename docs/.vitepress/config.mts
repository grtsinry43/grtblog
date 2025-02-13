import {defineConfig} from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
    title: "Grtblog",
    description: "博客站点的一站式解决方案，使用 Nextjs + SpringBoot 构建的现代化博客管理系统。",
    themeConfig: {
        // https://vitepress.dev/reference/default-theme-config
        nav: [
            {text: '首页', link: '/'},
            {text: '快速开始', link: '/quick-start'},
            {text: 'API', link: '/api-examples'},
        ],

        sidebar: [
            {
                text: '快速开始',
                items: [
                    {text: '使用Docker Compose部署', link: '/quick-start/docker-compose'},
                    {text: '搜索与推荐', link: '/quick-start/search-and-recommend'},
                    {text: '后端', link: '/quick-start/backend'},
                    {text: 'Nginx配置', link: '/quick-start/nginx-config'},
                    {text: '前端', link: '/quick-start/frontend'},
                    {text: '管理后台', link: '/quick-start/admin'},
                ]
            },
            {
                text: 'API 参考',
                items: [
                    {text: 'Markdown Examples', link: '/markdown-examples'},
                    {text: '完整接口文档', link: '/api-examples'}
                ]
            }
        ],

        socialLinks: [
            {icon: 'github', link: 'https://github.com/grtsinry43/grtblog'},
        ]
    }
})
