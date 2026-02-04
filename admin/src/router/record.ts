import type { MenuMixedOptions } from './interface'

export const routeRecordRaw: MenuMixedOptions[] = [
  {
    path: 'dashboard',
    name: 'dashboard',
    icon: 'icon-[mage--dashboard-chart]',
    label: '仪表板',
    meta: {
      componentName: 'Dashboard',
      pinned: true,
      showTab: true,
    },
    component: 'dashboard/index',
  },
  {
    path: 'articles',
    name: 'articleManagement',
    icon: 'iconify ph--article',
    label: '文章管理',
    redirect: 'articles/list',
    children: [
      {
        path: 'list',
        name: 'articleList',
        label: '文章列表',
        icon: 'iconify ph--list-bullets',
        meta: {
          componentName: 'ArticleList',
          showTab: true,
        },
        component: 'articles/index',
      },
      {
        path: 'edit/new',
        name: 'articleCreate',
        label: '新建文章',
        show: false,
        meta: {
          componentName: 'ArticleEdit',
          showTab: true,
          enableMultiTab: true,
          renderTabTitle() {
            return '新建文章'
          },
        },
        component: 'articles/edit',
      },
      {
        path: 'edit/:id',
        name: 'articleEdit',
        label: '编辑文章',
        show: false,
        meta: {
          componentName: 'ArticleEdit',
          showTab: true,
          enableMultiTab: true,
          renderTabTitle({ id }) {
            return `编辑文章${id ? `-${id}` : ''}`
          },
        },
        component: 'articles/edit',
      },
    ],
  },
  {
    path: 'notes',
    name: 'noteManagement',
    icon: 'iconify ph--aperture-thin',
    label: '手记管理',
    redirect: 'notes/list',
    children: [
      {
        path: 'list',
        name: 'noteList',
        label: '手记列表',
        icon: 'iconify ph--note',
        meta: {
          componentName: 'NoteList',
          showTab: true,
        },
        component: 'notes/index',
      },
      {
        path: 'edit/new',
        name: 'noteCreate',
        label: '新建手记',
        show: false,
        meta: {
          componentName: 'NoteEdit',
          showTab: true,
          enableMultiTab: true,
          renderTabTitle() {
            return '新建手记'
          },
        },
        component: 'notes/edit',
      },
      {
        path: 'edit/:id',
        name: 'noteEdit',
        label: '编辑手记',
        show: false,
        meta: {
          componentName: 'NoteEdit',
          showTab: true,
          enableMultiTab: true,
          renderTabTitle({ id }) {
            return `编辑手记${id ? `-${id}` : ''}`
          },
        },
        component: 'notes/edit',
      },
    ],
  },
  {
    path: 'thinkings',
    name: 'thinkingManagement',
    icon: 'iconify ph--lightbulb-filament',
    label: '思考管理',
    redirect: 'thinkings/list',
    children: [
      {
        path: 'list',
        name: 'thinkingList',
        label: '思考列表',
        icon: 'iconify ph--list-bullets',
        meta: {
          componentName: 'ThinkingList',
          showTab: true,
        },
        component: 'thinking/index',
      },
      {
        path: 'create',
        name: 'thinkingCreate',
        label: '新建思考',
        show: false,
        meta: {
          componentName: 'ThinkingEdit',
          showTab: true,
          enableMultiTab: true,
          renderTabTitle() {
            return '新建思考'
          },
        },
        component: 'thinking/edit',
      },
      {
        path: 'edit/:id',
        name: 'thinkingEdit',
        label: '编辑思考',
        show: false,
        meta: {
          componentName: 'ThinkingEdit',
          showTab: true,
          enableMultiTab: true,
          renderTabTitle({ id }) {
            return `编辑思考${id ? `-${id}` : ''}`
          },
        },
        component: 'thinking/edit',
      },
    ],
  },
  {
    path: 'pages',
    name: 'pageManagement',
    icon: 'iconify ph--layout',
    label: '页面管理',
    redirect: 'pages/list',
    children: [
      {
        path: 'list',
        name: 'pageList',
        label: '页面列表',
        icon: 'iconify ph--file-text',
        meta: {
          componentName: 'PageList',
          showTab: true,
        },
        component: 'pages/index',
      },
      {
        path: 'create',
        name: 'pageCreate',
        label: '新建页面',
        show: false,
        meta: {
          componentName: 'PageEdit',
          showTab: true,
          enableMultiTab: true,
          renderTabTitle() {
            return '新建页面'
          },
        },
        component: 'pages/edit',
      },
      {
        path: 'edit/:id',
        name: 'pageEdit',
        label: '编辑页面',
        show: false,
        meta: {
          componentName: 'PageEdit',
          showTab: true,
          enableMultiTab: true,
          renderTabTitle({ id }) {
            return `编辑页面${id ? `-${id}` : ''}`
          },
        },
        component: 'pages/edit',
      },
    ],
  },
  {
    path: 'albums',
    name: 'albumManagement',
    icon: 'iconify ph--image',
    label: '相册管理',
    redirect: 'albums/list',
    children: [
      {
        path: 'list',
        name: 'albumList',
        label: '相册列表',
        icon: 'iconify ph--image',
        meta: {
          componentName: 'AlbumList',
          showTab: true,
        },
        component: 'albums/index',
      },
    ],
  },
  {
    path: 'comments',
    name: 'commentManagement',
    icon: 'iconify ph--chat-circle-text',
    label: '评论管理',
    meta: {
      componentName: 'CommentList',
      showTab: true,
    },
    component: 'comments/index',
  },
  {
    path: 'friend-links',
    name: 'friendLinkManagement',
    icon: 'iconify ph--link',
    label: '友链',
    redirect: 'friend-links/list',
    children: [
      {
        path: 'list',
        name: 'friendLinkList',
        label: '友链列表',
        icon: 'iconify ph--link',
        meta: {
          componentName: 'FriendLinkList',
          showTab: true,
        },
        component: 'friend-links/index',
      },
    ],
  },
  {
    path: 'federation',
    name: 'unionManagement',
    icon: 'iconify ph--circles-three',
    label: '联合',
    redirect: 'federation/debug',
    children: [
      {
        path: 'debug',
        name: 'federationDebug',
        label: '联邦调试',
        icon: 'iconify ph--bug',
        meta: {
          componentName: 'FederationDebug',
          showTab: true,
        },
        component: 'federation/debug',
      },
      {
        path: 'instances',
        name: 'federationInstances',
        label: '联邦实例',
        icon: 'iconify ph--network',
        meta: {
          componentName: 'FederationInstances',
          showTab: true,
        },
        component: 'federation/instances',
      },
      {
        path: 'settings',
        name: 'unionSettings',
        label: '联合设置',
        icon: 'iconify ph--circles-three',
        meta: {
          componentName: 'FederationSettings',
          showTab: true,
        },
        component: 'federation/settings',
      },
    ],
  },
  {
    path: 'files',
    name: 'fileManagement',
    icon: 'icon-[fluent--cloud-arrow-up-24-regular]',
    label: '文件管理',
    redirect: 'files/list',
    children: [
      {
        path: 'list',
        name: 'fileList',
        label: '文件列表',
        icon: 'icon-[fluent--cloud-arrow-up-24-regular]',
        meta: {
          componentName: 'FileList',
          showTab: true,
        },
        component: 'uploads/index',
      },
    ],
  },
  {
    path: 'plugins',
    name: 'pluginManagement',
    icon: 'iconify ph--puzzle-piece',
    label: '插件与云函数',
    redirect: 'plugins/list',
    children: [
      {
        path: 'list',
        name: 'pluginList',
        label: '插件与云函数',
        icon: 'iconify ph--puzzle-piece',
        meta: {
          componentName: 'PluginList',
          showTab: true,
        },
        component: 'plugins/index',
      },
    ],
  },
  {
    path: 'webhooks',
    name: 'webhookList',
    icon: 'iconify ph--webhooks-logo',
    label: 'Webhook',
    meta: {
      componentName: 'WebhookList',
      showTab: true,
    },
    component: 'webhooks/index',
  },
  {
    path: 'email',
    name: 'emailManagement',
    icon: 'iconify ph--envelope',
    label: '邮件管理',
    redirect: 'email/templates',
    children: [
      {
        path: 'templates',
        name: 'emailTemplateList',
        label: '邮件模版',
        icon: 'iconify ph--scroll',
        meta: {
          componentName: 'EmailTemplateList',
          showTab: true,
        },
        component: 'email/templates/index',
      },
      {
        path: 'templates/new',
        name: 'emailTemplateCreate',
        label: '新建模版',
        show: false,
        meta: {
          componentName: 'EmailTemplateEdit',
          showTab: true,
          enableMultiTab: true,
          renderTabTitle() {
            return '新建模版'
          },
        },
        component: 'email/templates/edit',
      },
      {
        path: 'templates/:code',
        name: 'emailTemplateEdit',
        label: '编辑模版',
        show: false,
        meta: {
          componentName: 'EmailTemplateEdit',
          showTab: true,
          enableMultiTab: true,
          renderTabTitle({ code }) {
            return `编辑模版-${code}`
          },
        },
        component: 'email/templates/edit',
      },
      {
        path: 'subscriptions',
        name: 'emailSubscriptionList',
        label: '订阅管理',
        icon: 'iconify ph--users',
        meta: {
          componentName: 'EmailSubscriptionList',
          showTab: true,
        },
        component: 'email/subscriptions/index',
      },
      {
        path: 'test',
        name: 'emailTest',
        label: '邮件测试',
        icon: 'iconify ph--paper-plane-tilt',
        meta: {
          componentName: 'EmailTest',
          showTab: true,
        },
        component: 'email/test/index',
      },
    ],
  },
  {
    path: 'navigation',
    name: 'navMenuManagement',
    icon: 'iconify ph--list',
    label: '导航菜单',
    meta: {
      componentName: 'NavMenuManagement',
      showTab: true,
    },
    component: 'navigation/index',
  },
  {
    path: 'settings',
    name: 'settings',
    icon: 'iconify ph--gear-six',
    label: '设置',
    redirect: 'settings/site-info',
    children: [
      {
        path: 'site-info',
        name: 'siteInfo',
        label: '站点信息',
        icon: 'iconify ph--globe-hemisphere-west',
        meta: {
          componentName: 'SiteInfo',
          showTab: true,
        },
        component: 'settings/site-info/index',
      },
      {
        path: 'login-methods',
        name: 'loginMethods',
        label: '登录方式',
        icon: 'iconify ph--shield-check',
        meta: {
          componentName: 'LoginMethods',
          showTab: true,
        },
        component: 'settings/login-methods/index',
      },
      {
        path: 'system',
        name: 'systemSettings',
        label: '系统设置',
        icon: 'iconify ph--gear',
        meta: {
          componentName: 'SystemSettings',
          showTab: true,
        },
        component: 'sysconfig/index',
      },
    ],
  },
  {
    path: 'advanced',
    name: 'advancedInfo',
    icon: 'iconify ph--info',
    label: '高级信息',
    redirect: 'advanced/overview',
    children: [
      {
        path: 'overview',
        name: 'advancedOverview',
        label: '高级信息',
        icon: 'iconify ph--info',
        meta: {
          componentName: 'AdvancedInfo',
          showTab: true,
        },
        component: 'advanced/index',
      },
    ],
  },
  {
    path: 'monitoring',
    name: 'systemMonitor',
    icon: 'iconify ph--activity',
    label: '系统监控',
    redirect: 'monitoring/overview',
    children: [
      {
        path: 'overview',
        name: 'systemMonitorOverview',
        label: '系统监控',
        icon: 'iconify ph--activity',
        meta: {
          componentName: 'SystemMonitor',
          showTab: true,
        },
        component: 'monitoring/index',
      },
      {
        path: 'logs',
        name: 'systemLogs',
        label: '系统日志',
        icon: 'iconify ph--scroll',
        meta: {
          componentName: 'SystemLogs',
          showTab: true,
        },
        component: 'monitoring/logs',
      },
    ],
  },
  {
    path: '/about',
    key: 'about',
    name: 'about',
    icon: 'iconify ph--info',
    label: '关于',
    component: 'about/index',
    meta: {
      showTab: true,
    },
  },
]
