/**
// @ts-check
/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */

const sidebars = {
  sidebar: [
    {
      type: "doc",
      id: "introduction",
    },

    {
      type: "category",
      label: "Getting Started",
      collapsible: true,
      collapsed: false,
      items: ["installation", "shell-completion", "man-pages", "usage"],
    },

    {
      type: "category",
      label: "Documentation",
      collapsible: true,
      collapsed: false,
      items: [
        "examples",
        "config",
        "commands",
      ]
    },

    {
      type: "doc",
      id: "changelog",
    },

    {
      type: "category",
      label: "Development",
      collapsible: true,
      collapsed: false,
      items: ["development", "contributing"],
    },
  ],
};

module.exports = sidebars;
