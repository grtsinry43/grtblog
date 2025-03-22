import { defineConfig } from '@umijs/max';

export default defineConfig({
  base: process.env.NODE_ENV === 'production' ? '/admin/' : '/',
  publicPath: process.env.NODE_ENV === 'production' ? '/admin/' : '/',
  antd: {
    // themes
    // dark: true,
    // less or css, default less
    // style: 'less',
    // // shortcut of `configProvider.theme`
    // // use to configure theme token, antd v5 only
    // theme: {},
    // // antd <App /> valid for version 5.1.0 or higher, default: undefined
    // appConfig: {},
    // // Transform DayJS to MomentJS
    // momentPicker: true,
    // // Add StyleProvider for legacy browsers
    // styleProvider: {
    //   hashPriority: 'high',
    //   legacyTransformer: true,
    // },
  },
  access: {},
  dva: {},
  model: {},
  initialState: {},
  request: {},
  // ssr: {},
  layout: {
    title: 'Grtblog Admin',
    favicon: '/favicon.ico',
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
      icon: 'PicLeftOutlined',
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
      icon: 'SnippetsOutlined',
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
      icon: 'FileTextOutlined',
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
      name: '相册',
      icon: 'PictureOutlined',
      path: '/album',
      routes: [
        {
          name: '添加图片',
          path: '/album/add',
          component: './Album/AddPhoto',
        },
      ],
    },
    {
      name: '评论管理',
      icon: 'comment',
      path: '/comment',
      component: './Comment',
    },
    {
      name: '常用工具',
      icon: 'ToolOutlined',
      path: '/tools',
      routes: [
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
      icon: 'LinkOutlined',
      component: './Friend',
    },
    {
      name: '自定义',
      path: '/custom',
      icon: 'SettingOutlined',
      routes: [
        {
          name: '导航菜单',
          path: '/custom/nav',
          component: './Custom/NavMenu',
        },
        {
          name: '页脚设置',
          path: '/custom/footer',
          component: './Custom/Footer',
        },
      ],
    },
    {
      name: '插件',
      path: '/plugin',
      icon: 'AppstoreAddOutlined',
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
      icon: 'SettingOutlined',
      path: '/config',
      component: './Config',
    },
    {
      name: '系统信息',
      icon: 'InfoCircleOutlined',
      path: '/system',
      component: './System',
    },
    { path: '/*', component: '@/pages/404.tsx' },
  ],
  npmClient: 'pnpm',
  proxy: {
    '/api/v1': {
      target: 'http://127.0.0.1:8082',
      changeOrigin: true,
      pathRewrite: { '^/api/v1': '' },
    },
    '/captcha': {
      target: 'http://127.0.0.1:8082',
      changeOrigin: true,
    },
    '/uploads': {
      target: 'http://127.0.0.1:8082',
      changeOrigin: true,
    },
    '/api/frontend': {
      target: 'http://127.0.0.1:3000',
      changeOrigin: true,
      pathRewrite: { '^/api/frontend': '' },
    },
  },
});
