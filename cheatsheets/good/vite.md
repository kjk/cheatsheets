---
Title: Vite
Category: JavaScript
---

# Main

[Vite](https://vitejs.dev/) is a web frontend tooling. It starts with
`index.html`, follows `.css` and `.js` references and builds
a bundled, optmized version in `dist` directory, using `rollup` as a bundler.

Main commands:
* `vite` : runs vite in dev mode, starts a http server to preview
* `vite build` : builds optimized output in `dist` directory
* `vite preview` : after `vite build` you can preview the output from `dist` directory
* `vite optimize`

```text
Usage:
  $ vite [root]

Commands:
  [root]
  build [root]
  optimize [root]
  preview [root]

For more info, run any command with the `--help` flag:
  $ vite --help
  $ vite build --help
  $ vite optimize --help
  $ vite preview --help

Options:
  --host [host]           [string] specify hostname
  --port <port>           [number] specify port
  --https                 [boolean] use TLS + HTTP/2
  --open [path]           [boolean | string] open browser on startup
  --cors                  [boolean] enable CORS
  --strictPort            [boolean] exit if specified port is already in use
  --force                 [boolean] force the optimizer to ignore the cache and re-bundle
  -c, --config <file>     [string] use specified config file
  -r, --root <path>       [string] use specified root directory
  --base <path>           [string] public base path (default: /)
  -l, --logLevel <level>  [string] info | warn | error | silent
  --clearScreen           [boolean] allow/disable clear screen when logging
  -d, --debug [feat]      [string | boolean] show debug logs
  -f, --filter <filter>   [string] filter debug logs
  -m, --mode <mode>       [string] set env mode
  -h, --help              Display this message
  -v, --version           Display version number
  ```


# build

```text
Usage:
  $ vite build [root]

Options:
  --target <target>             [string] transpile target (default: 'modules')
  --outDir <dir>                [string] output directory (default: dist)
  --assetsDir <dir>             [string] directory under outDir to place assets in (default: _assets)
  --assetsInlineLimit <number>  [number] static asset base64 inline threshold in bytes (default: 4096)
  --ssr [entry]                 [string] build specified entry for server-side rendering
  --sourcemap                   [boolean] output source maps for build (default: false)
  --minify [minifier]           [boolean | "terser" | "esbuild"] enable/disable minification, or specify minifier to use (default: esbuild)
  --manifest                    [boolean] emit build manifest json
  --ssrManifest                 [boolean] emit ssr manifest json
  --emptyOutDir                 [boolean] force empty outDir when it's outside of root
  -w, --watch                   [boolean] rebuilds when modules have changed on disk
  -c, --config <file>           [string] use specified config file
  -r, --root <path>             [string] use specified root directory
  --base <path>                 [string] public base path (default: /)
  -l, --logLevel <level>        [string] info | warn | error | silent
  --clearScreen                 [boolean] allow/disable clear screen when logging
  -d, --debug [feat]            [string | boolean] show debug logs
  -f, --filter <filter>         [string] filter debug logs
  -m, --mode <mode>             [string] set env mode
  -h, --help                    Display this message
  -v, --version                 Display version number
```

# optimize

```text
Usage:
  $ vite optimize [root]

Options:
  --force                 [boolean] force the optimizer to ignore the cache and re-bundle
  -c, --config <file>     [string] use specified config file
  -r, --root <path>       [string] use specified root directory
  --base <path>           [string] public base path (default: /)
  -l, --logLevel <level>  [string] info | warn | error | silent
  --clearScreen           [boolean] allow/disable clear screen when logging
  -d, --debug [feat]      [string | boolean] show debug logs
  -f, --filter <filter>   [string] filter debug logs
  -m, --mode <mode>       [string] set env mode
  -h, --help              Display this message
  -v, --version           Display version number
```

# preview

```text
Usage:
  $ vite preview [root]

Options:
  --host [host]           [string] specify hostname
  --port <port>           [number] specify port
  --https                 [boolean] use TLS + HTTP/2
  --open [path]           [boolean | string] open browser on startup
  --strictPort            [boolean] exit if specified port is already in use
  -c, --config <file>     [string] use specified config file
  -r, --root <path>       [string] use specified root directory
  --base <path>           [string] public base path (default: /)
  -l, --logLevel <level>  [string] info | warn | error | silent
  --clearScreen           [boolean] allow/disable clear screen when logging
  -d, --debug [feat]      [string | boolean] show debug logs
  -f, --filter <filter>   [string] filter debug logs
  -m, --mode <mode>       [string] set env mode
  -h, --help              Display this message
  -v, --version           Display version number
```
