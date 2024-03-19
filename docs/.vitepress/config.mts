import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'Azukiiro',
  description: '聚合评测方案',
  themeConfig: {
    logo: '/azukiiro.svg',

    nav: [
      { text: '首页', link: '/' },
      { text: '开始使用', link: '/guides/' },
      { text: '适配器文档', link: '/adapters/' }
    ],

    sidebar: [
      {
        text: '开始使用',
        link: '/guides/',
        items: [
          { text: '运维指南', link: '/guides/admin-guide' },
          { text: '开发指南', link: '/guides/dev-guide' }
        ]
      },
      {
        text: '适配器文档',
        link: '/adapters/',
        items: [
          { text: 'Dummy', link: '/adapters/dummy' },
          { text: 'UOJ', link: '/adapters/uoj' },
          { text: 'Glue', link: '/adapters/glue' },
          { text: 'VJudge', link: '/adapters/vjudge' }
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
