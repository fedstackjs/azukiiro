import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'Azukiiro',
  description: '聚合评测方案',
  themeConfig: {
    logo: '/azukiiro.svg',

    nav: [
      { text: '首页', link: '/' },
      { text: '开始使用', link: '/getting-started' }
    ],

    sidebar: [
      {
        text: '开始使用',
        items: [{ text: '开始使用', link: '/getting-started' }]
      },
      {
        text: '教程',
        items: [
          { text: '运维指南', link: '/admin-guide' },
          { text: '开发指南', link: '/dev-guide' }
        ]
      }
    ],

    socialLinks: [{ icon: 'github', link: 'https://github.com/fedstackjs/azukiiro' }],

    editLink: {
      pattern: 'https://github.com/fedstackjs/azukiiro/edit/main/docs/:path',
      text: 'Edit this page on GitHub'
    },

    footer: {
      message: 'Released under the AGPL-3.0 License.',
      copyright: 'Copyright © 2022-present FedStack'
    }
  }
})
