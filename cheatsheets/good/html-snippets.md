---
title: HTML Snippets
category: HTML, webdev
---

# Intro

Snippets of frequently used functionality.

# List

## Spinner

```css
#nprogress .spinner {
  display: block;
  position: fixed;
  z-index: 1031;
  top: 15px;
  right: 15px;
}

#nprogress .spinner-icon {
  width: 18px;
  height: 18px;
  box-sizing: border-box;

  border: solid 2px transparent;
  border-top-color: #29d;
  border-left-color: #29d;
  border-radius: 50%;

  -webkit-animation: nprogress-spinner 400ms linear infinite;
          animation: nprogress-spinner 400ms linear infinite;
}

```

```html
<div class="spinner">
  <div class="spinner-icon"></div>
</div>
```

From: https://cdn.jsdelivr.net/npm/nprogress@0.2.0/nprogress.css and https://ricostacruz.com/nprogress/

