---
title: Alpine.js
category: JavaScript library
---

# Intro

[Alpine.js](https://alpinejs.dev/) is a minimalist, reactive JavaScript framework.

To include Alpine.js in your HTML:
```html
<script src="//unpkg.com/alpinejs" defer></script>

<div x-data="{ open: false }">
    <button @click="open = true">Expand</button>
 
    <span x-show="open">
      Content...
    </span>
</div>
```

# Directives

## x-data

`x-data` defines a chunk of HTML as an Alpine component and provides the reactive data for that component to reference.

```html
<div x-data="{ open: false }">
	<button @click="open = ! open">Toggle Content</button>

	<div x-show="open">
		Content...
	</div>
</div>
```

### Scope

Properties defined in an `x-data` directive are available to all element children. Even ones inside other, nested `x-data` components.

For example:
```html
<div x-data="{ foo: 'bar' }">
	<span x-text="foo"><!-- Will output: "bar" --></span>

	<div x-data="{ bar: 'baz' }">
		<span x-text="foo"><!-- Will output: "bar" --></span>

		<div x-data="{ foo: 'bob' }">
			<span x-text="foo"><!-- Will output: "bob" --></span>
		</div>
	</div>
</div>
```

### Methods
Because `x-data` is evaluated as a normal JavaScript object, in addition to state, you can store methods and even getters.

For example, let's extract the "Toggle Content" behavior into a method on `x-data`.

```html
<div x-data="{ open: false, toggle() { this.open = ! this.open } }">
	<button @click="toggle()">Toggle Content</button>

	<div x-show="open">
		Content...
	</div>
</div>
```

Notice the added `toggle() { this.open = ! this.open }` method on `x-data`. This method can now be called from anywhere inside the component.

You'll also notice the usage of `this.` to access state on the object itself. This is because Alpine evaluates this data object like any standard JavaScript object with a `this` context.

If you prefer, you can leave the calling parenthesis off of the `toggle` method completely. For example:

```html
<!-- Before -->
<button @click="toggle()">...</button>

<!-- After -->
<button @click="toggle">...</button>
```


### Getters

JavaScript [getters](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Functions/get) are handy when the sole purpose of a method is to return data based on other state.

Think of them like "computed properties" (although, they are not cached like Vue's computed properties).

Let's refactor our component to use a getter called `isOpen` instead of accessing `open` directly.

```html
<div x-data="{
	open: false,
	get isOpen() { return this.open },
	toggle() { this.open = ! this.open },
}">
	<button @click="toggle()">Toggle Content</button>

	<div x-show="isOpen">
		Content...
	</div>
</div>
```

Notice the "Content" now depends on the `isOpen` getter instead of the `open` property directly.

In this case there is no tangible benefit. But in some cases, getters are helpful for providing a more expressive syntax in your components.

### Data-less components

Occasionally, you want to create an Alpine component, but you don't need any data.

In these cases, you can always pass in an empty object.

```html
<div x-data="{}"...
```

However, if you wish, you can also eliminate the attribute value entirely if it looks better to you.

```html
<div x-data...
```

### Single-element components
Sometimes you may only have a single element inside your Alpine component, like the following:

```html
<div x-data="{ open: true }">
	<button @click="open = false" x-show="open">Hide Me</button>
</div>
```

In these cases, you can declare `x-data` directly on that single element:

```html
<button x-data="{ open: true }" @click="open = false" x-show="open">
	Hide Me
</button>
```

### Re-usable Data

If you find yourself duplicating the contents of `x-data`, or you find the inline syntax verbose, you can extract the `x-data` object out to a dedicated component using `Alpine.data`.

Here's a quick example:

```html
<div x-data="dropdown">
	<button @click="toggle">Toggle Content</button>

	<div x-show="open">
		Content...
	</div>
</div>

<script>
	document.addEventListener('alpine:init', () => {
		Alpine.data('dropdown', () => ({
			open: false,
			toggle() {
				this.open = ! this.open
			},
		}))
	})
</script>
```

## x-bind

Dynamically set HTML attributes on an element

```html
<div x-bind:class="! open ? 'hidden' : ''">
  ...
</div>
```

## x-on

```html
<button x-on:click="open = ! open">
  Toggle
</button>
```

## x-text

Set the text content of an element

```html

<div>
  Copyright ©
 
  <span x-text="new Date().getFullYear()"></span>
</div>
```

## x-html

Set the inner HTML of an element

```html
<div x-html="(await axios.get('/some/html/partial')).data">
  ...
</div>
```

## x-model

Synchronize a piece of data with an input element

```html
<div x-data="{ search: '' }">
  <input type="text" x-model="search">
 
  Searching for: <span x-text="search"></span>
</div>```

## x-show

Toggle the visibility of an element.

```html
<div x-show="open">
  ...
</div>
```


## x-transition

Transition an element in and out using CSS transitions

```html
<div x-show="open" x-transition>
  ...
</div>
```

## x-for

Repeat a block of HTML based on a data set

```html
<template x-for="post in posts">
  <h2 x-text="post.title"></h2>
</template>
```

## x-if

Conditionally add/remove a block of HTML from the page entirely.

Only use on `<template>`, use x-show for HTML elements.

```html
<template x-if="open">
  <div>...</div>
</template>
```

## x-init

Run code when an element is initialized by Alpine

```html
<div x-init="date = new Date()"></div>
```

## x-effect

Execute a script each time one if its dependancies change

```html
<div x-effect="console.log('Count is '+count)"></div>
```

## x-ref

Reference elements directly by their specified keys using the $refs magic property

```html
<input type="text" x-ref="content">
 
<button x-on:click="navigator.clipboard.writeText($refs.content.value)">
  Copy
</button>
```

## x-cloak

Hide a block of HTML until after Alpine is finished initializing its contents

```html
<div x-cloak>
  ...
</div>
```

## x-ignore

Prevent a block of HTML from being initialized by Alpine

```html
<div x-ignore>
  ...
</div>
```

# Properties

## $store

Access a global store registered using Alpine.store(...)

```html
<h1 x-text="$store.site.title"></h1>
```

## $el

Reference the current DOM element

```html
<div x-init="new Pikaday($el)"></div>
```

## $dispatch

Dispatch a custom browser event from the current element

```html
<div x-on:notify="...">
  <button x-on:click="$dispatch('notify')">...</button>
</div>
```

## $watch

Watch a piece of data and run the provided callback anytime it changes

```html
<div x-init="$watch('count', value => {
  console.log('count is ' + value))"
}">...</div>
```

## $refs

Reference an element by key (specified using x-ref)

```html
<div x-init="$refs.button.remove()">
  <button x-ref="button">Remove Me</button>
</div>
```

## $nextTick

Wait until the next "tick" (browser paint) to run a bit of code

```html
<div
  x-text="count"
  x-text="$nextTick(() => {"
    console.log('count is ' + $el.textContent)
  })
>...</div>
```


# Methods

## Alpine.data

Reuse a data object and reference it using x-data

```html
<div x-data="dropdown">
  ...
</div>
```

```js
 
Alpine.data('dropdown', () => ({
  open: false,
 
  toggle() { 
    this.open = ! this.open
  }
}))
```

## Alpine.store

Declare a piece of global, reactive, data that can be accessed from anywhere using $store

```html
<button @click="$store.notifications.notify('...')">
  Notify
</button>
```

```js
Alpine.store('notifications', {
  items: [],
 
  notify(message) { 
    this.items.push(message)
  }
})
```
