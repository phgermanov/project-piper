import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  // Overview sidebar
  overviewSidebar: [
    {
      type: 'doc',
      id: 'README',
      label: 'Introduction',
    },
    {
      type: 'category',
      label: 'Overview',
      collapsed: false,
      items: [
        'overview/architecture',
        'overview/project-structure',
        'overview/how-it-works',
      ],
    },
  ],

  // Jenkins Library sidebar
  jenkinsLibrarySidebar: [
    {
      type: 'category',
      label: 'Jenkins Library (piper-os)',
      collapsed: false,
      items: [
        'jenkins-library/overview',
        'jenkins-library/build-tools',
        'jenkins-library/security-scanning',
        'jenkins-library/testing-frameworks',
        'jenkins-library/deployment',
        'jenkins-library/sap-integration',
        'jenkins-library/abap-development',
        'jenkins-library/container-operations',
        'jenkins-library/version-control',
        'jenkins-library/utilities',
      ],
    },
  ],

  // GitHub Pipeline sidebar
  githubPipelineSidebar: [
    {
      type: 'category',
      label: 'GitHub Pipeline (GPP)',
      collapsed: false,
      items: [
        'github-pipeline/overview',
        'github-pipeline/init-stage',
        'github-pipeline/build-stage',
        'github-pipeline/integration-stage',
        'github-pipeline/acceptance-stage',
        'github-pipeline/performance-stage',
        'github-pipeline/promote-stage',
        'github-pipeline/release-stage',
        'github-pipeline/post-stage',
        'github-pipeline/oss-ppms-stages',
      ],
    },
  ],

  // Azure DevOps sidebar
  azureDevOpsSidebar: [
    {
      type: 'category',
      label: 'Azure DevOps Integration',
      collapsed: false,
      items: [
        'azure-devops/overview',
        'azure-devops/azure-task',
        'azure-devops/pipeline-templates',
      ],
    },
  ],

  // GitHub Action sidebar
  githubActionSidebar: [
    {
      type: 'category',
      label: 'GitHub Action',
      collapsed: false,
      items: [
        'github-action/overview',
        'github-action/features',
        'github-action/usage-guide',
      ],
    },
  ],

  // Configuration sidebar
  configurationSidebar: [
    {
      type: 'category',
      label: 'Configuration',
      collapsed: false,
      items: [
        'configuration/overview',
        'configuration/configuration-hierarchy',
        'configuration/default-settings',
        'configuration/platform-deviations',
        'configuration/stage-configuration',
        'configuration/step-configuration',
        'configuration/credentials-management',
      ],
    },
  ],

  // Guides sidebar
  guidesSidebar: [
    {
      type: 'doc',
      id: 'guides/README',
      label: 'Getting Started',
    },
    {
      type: 'category',
      label: 'Setup Guides',
      collapsed: false,
      items: [
        'guides/jenkins-setup',
        'guides/github-setup',
        'guides/azure-setup',
      ],
    },
    {
      type: 'category',
      label: 'Advanced',
      collapsed: false,
      items: [
        'guides/extensibility',
        'guides/migration',
      ],
    },
  ],

  // Resources sidebar
  resourcesSidebar: [
    {
      type: 'category',
      label: 'Resources',
      collapsed: false,
      items: [
        'resources/step-reference',
        'resources/faq',
        'resources/troubleshooting',
        'resources/glossary',
      ],
    },
  ],
};

export default sidebars;
