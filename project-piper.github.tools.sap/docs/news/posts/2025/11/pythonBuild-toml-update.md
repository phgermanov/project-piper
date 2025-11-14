---
date: 2025-11-03
title: Support for toml file format in pythonBuild step
authors:
 - petko
categories:
 - General Purpose Pipeline
 - Python
---

The `pythonBuild` step will support the `pyproject.toml` file format, enabling the step to recognize and utilize modern Python build metadata.

<!-- more -->

## 📢 Do I need to do something?

If your project uses either one of the file formats (`setup.py` or `pyproject.toml`), no action is required. The `pythonBuild` step will automatically detect and use the appropriate build descriptor file.

However, if your project includes **both** file formats in your project files, the `pythonBuild` step will prioritize `pyproject.toml` for the build process. If you prefer to use `setup.py` instead, you will need to remove the `pyproject.toml` file from your project files.
When you update to this version, make sure your toml file is properly maintained.

## 💡 What do I need to know?

The update will be released with the next Piper version on November 10, 2025.
This enhancement allows the `pythonBuild` step to automatically detect the presence of a `pyproject.toml` file, parse its contents to determine build requirements and configuration, and execute the appropriate build commands based on the information provided. The goal is to ensure compatibility with both legacy `setup.py` files and the newer TOML-based build system, providing a seamless experience for users migrating to or adopting the `pyproject.toml` standard.

## 📖 Learn more

- [pythonBuild step documentation](https://pages.github.tools.sap/project-piper/steps/pythonBuild/)
