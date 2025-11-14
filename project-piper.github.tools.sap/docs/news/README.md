# Piper News

This section describes how to create a new entry in the Piper news section (blog post).

## Post a new article

To write a new blog post, make a copy of `docs/news/_template.md` and place it into `docs/new/posts/<currentYear>/<currentMonth>/<title>.md`.

Each blog post needs at least a `date` in the [metadata section](https://squidfunk.github.io/mkdocs-material/plugins/blog/#metadata). Feel free to add one or more [authors](#author) in the metadata section.

### Teaser

Each blog post needs to have a `<!-- more -->` tag to be able to display on the overview page properly. The template already contains it, make sure that the paragraph before acts as a teaser.

### Author

Each blog post can have one or multiple authors. Make sure the author exists in `docs/news/.authors.yml`.

```yml
  christopher:
    name: Christopher Fenner
    description: Developer
    avatar: https://github.wdf.sap.corp/d065687.png
    url: https://github.wdf.sap.corp/d065687
```

The `avatar` should point to `https://github.wdf.sap.corp/<sap-user-id>.png` as these images are accessible without authentication.
