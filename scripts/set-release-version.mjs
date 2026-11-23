import { readFile, writeFile } from 'node:fs/promises'

const tag = process.argv[2] ?? ''
if (!/^v[1-9]\d*\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/.test(tag)) {
  throw new Error('Use a stable release tag such as v1.0.0, v1.1.0, or v1.1.1.')
}
const version = tag.slice(1)

const packageFields = [
  ['build/config.yml', /^(  version: ")\d+\.\d+\.\d+(")/gm, 1],
  ['build/linux/nfpm/nfpm.yaml', /^(version: ")\d+\.\d+\.\d+(")/gm, 1],
  [
    'build/windows/info.json',
    /("(?:file_version|product_version|FileVersion|ProductVersion)": ")\d+\.\d+\.\d+(")/g,
    4,
  ],
  [
    'build/windows/nsis/wails_tools.nsh',
    /(!define INFO_PRODUCTVERSION ")\d+\.\d+\.\d+(")/g,
    1,
  ],
  [
    'build/darwin/Info.plist',
    /(<key>CFBundle(?:ShortVersionString|Version)<\/key>\s*<string>)\d+\.\d+\.\d+(<\/string>)/g,
    2,
  ],
]

const updates = await Promise.all(
  packageFields.map(async ([path, field, expected]) => {
    const source = await readFile(path, 'utf8')
    if ([...source.matchAll(field)].length !== expected) {
      throw new Error(`Could not locate release version fields in ${path}.`)
    }
    return [
      path,
      source.replace(field, (_, prefix, suffix) => prefix + version + suffix),
    ]
  }),
)
await Promise.all(updates.map(([path, contents]) => writeFile(path, contents)))
