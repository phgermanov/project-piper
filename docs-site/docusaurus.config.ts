import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'Project Piper Documentation',
  tagline: 'Enterprise CI/CD for SAP Ecosystem',
  favicon: 'img/favicon.ico',

  future: {
    v4: true,
  },

  // Set the production url of your site here
  url: 'https://phgermanov.github.io',
  // For GitHub pages deployment
  baseUrl: '/project-piper/',

  // GitHub pages deployment config
  organizationName: 'phgermanov',
  projectName: 'project-piper',

  onBrokenLinks: 'warn',
  onBrokenAnchors: 'warn',

  markdown: {
    format: 'mdx',
    mermaid: true,
    hooks: {
      onBrokenMarkdownLinks: 'warn',
    },
  },

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: {
          routeBasePath: '/', // Serve docs at the root
          sidebarPath: './sidebars.ts',
          editUrl: 'https://github.com/phgermanov/project-piper/tree/main/docs-site/',
          showLastUpdateTime: true,
          showLastUpdateAuthor: true,
        },
        blog: false, // Disable blog
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themes: [
    [
      require.resolve("@easyops-cn/docusaurus-search-local"),
      {
        hashed: true,
        language: ["en"],
        highlightSearchTermsOnTargetPage: true,
        explicitSearchResultPath: true,
      },
    ],
  ],

  themeConfig: {
    image: 'img/piper-social-card.jpg',
    colorMode: {
      defaultMode: 'light',
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'Project Piper',
      logo: {
        alt: 'Project Piper Logo',
        src: 'img/logo.svg',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'overviewSidebar',
          position: 'left',
          label: 'Overview',
        },
        {
          type: 'docSidebar',
          sidebarId: 'jenkinsLibrarySidebar',
          position: 'left',
          label: 'Jenkins Library',
        },
        {
          type: 'docSidebar',
          sidebarId: 'githubPipelineSidebar',
          position: 'left',
          label: 'GitHub Pipeline',
        },
        {
          type: 'docSidebar',
          sidebarId: 'guidesSidebar',
          position: 'left',
          label: 'Guides',
        },
        {
          type: 'search',
          position: 'right',
        },
        {
          href: 'https://github.com/phgermanov/project-piper',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Documentation',
          items: [
            {
              label: 'Overview',
              to: '/overview/architecture',
            },
            {
              label: 'Jenkins Library',
              to: '/jenkins-library/overview',
            },
            {
              label: 'GitHub Pipeline',
              to: '/github-pipeline/overview',
            },
          ],
        },
        {
          title: 'Components',
          items: [
            {
              label: 'Azure DevOps',
              to: '/azure-devops/overview',
            },
            {
              label: 'GitHub Action',
              to: '/github-action/overview',
            },
            {
              label: 'Configuration',
              to: '/configuration/overview',
            },
          ],
        },
        {
          title: 'Resources',
          items: [
            {
              label: 'Step Reference',
              to: '/resources/step-reference',
            },
            {
              label: 'FAQ',
              to: '/resources/faq',
            },
            {
              label: 'Troubleshooting',
              to: '/resources/troubleshooting',
            },
          ],
        },
        {
          title: 'More',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/phgermanov/project-piper',
            },
            {
              label: 'SAP Project Piper',
              href: 'https://www.project-piper.io/',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} Project Piper. Documentation built with Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['groovy', 'yaml', 'bash', 'typescript', 'java', 'python', 'go'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
