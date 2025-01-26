import { defineConfig } from '@umijs/max';

export default defineConfig({
  base: process.env.NODE_ENV === 'production' ? '/admin/' : '/',
  publicPath: process.env.NODE_ENV === 'production' ? '/admin/' : '/',
  antd: {},
  access: {},
  dva: {},
  model: {},
  initialState: {},
  request: {},
  layout: {
    title: 'Grtblog 后台管理系统',
  },
  routes: [
    {
      path: '/',
      redirect: '/home',
    },
    {
      name: '首页',
      icon: 'Home',
      path: '/home',
      component: './Home',
    },
    {
      name: '文章',
      icon: '',
      path: '/article',
      routes: [
        {
          name: '文章列表',
          path: '/article/list',
          component: './Article',
        },
        {
          name: '添加文章',
          path: '/article/add',
          component: './Article/AddArticle',
        },
        {
          name: '编辑文章',
          path: '/article/edit/:id',
          component: './Article/EditArticle',
          hideInMenu: true,
        },
      ],
    },
    {
      name: '随手记录',
      path: '/moment',
      routes: [
        {
          name: '随手记录列表',
          path: '/moment/list',
          component: './Moment',
        },
        {
          name: '添加随手记录',
          path: '/moment/add',
          component: './Moment/AddMoment',
        },
        {
          name: '编辑随手记录',
          path: '/moment/edit/:id',
          component: './Moment/EditMoment',
          hideInMenu: true,
        },
      ],
    },
    {
      name: '页面管理',
      path: '/pages',
      routes: [
        {
          name: '全部页面',
          path: '/pages/list',
          component: './Page',
        },
        {
          name: '添加页面',
          path: '/pages/add',
          component: './Page/AddPage',
        },
        {
          name: '编辑页面',
          path: '/pages/edit/:id',
          component: './Page/EditPage',
          hideInMenu: true,
        },
      ],
    },
    {
      name: '常用工具',
      path: '/tools',
      routes: [
        {
          name: '全部工具',
          path: '/tools/list',
          component: './Tool',
        },
        {
          name: '从文件导入',
          path: '/tools/addArticleFromFile',
          component: './Tool/ImportMd',
        },
      ],
    },
    {
      name: '友链',
      path: '/friend',
      routes: [
        {
          name: '友链列表',
          path: '/friend/list',
          component: './Friend/FriendList',
        },
      ],
    },
    {
      name: '插件',
      path: '/plugin',
      routes: [
        {
          name: '插件列表',
          path: '/plugin/list',
          component: './Plugin/PluginList',
        },
        {
          name: '插件市场',
          path: '/plugin/market',
          component: './Plugin/PluginMarket',
        },
        {
          name: '插件设置',
          path: '/plugin/setting/:pluginName',
          component: './Plugin/PluginSetting',
          hideInMenu: true,
        },
      ],
    },
    {
      name: '登录',
      path: '/login',
      component: './Login',
      hideInMenu: true,
      menuRender: false,
    },
    {
      name: '配置信息',
      path: '/config',
      component: './Config',
    },
  ],
  npmClient: 'pnpm',
  proxy: {
    '/api/v1': {
      target: 'http://127.0.0.1:8080',
      changeOrigin: true,
      pathRewrite: { '^/api/v1': '' },
    },
    '/captcha': {
      target: 'http://127.0.0.1:8080',
      changeOrigin: true,
    },
    '/uploads': {
      target: 'http://127.0.0.1:8080',
      changeOrigin: true,
    },
    '/api/frontend': {
      target: 'http://127.0.0.1:3000',
      changeOrigin: true,
      pathRewrite: { '^/api/frontend': '' },
    },
  },
});
