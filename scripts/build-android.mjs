import { execFileSync } from 'node:child_process'
import {
  chmodSync,
  copyFileSync,
  cpSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  writeFileSync,
} from 'node:fs'
import { join, resolve } from 'node:path'

const sdkPath = process.env.ANDROID_HOME || process.env.ANDROID_SDK_ROOT
if (!sdkPath) throw new Error('Set ANDROID_HOME to the Android SDK directory.')
const sdk = resolve(sdkPath)
const ndk = process.env.ANDROID_NDK_HOME || join(sdk, 'ndk', '29.0.14206865')
const host = { win32: 'windows', linux: 'linux', darwin: 'darwin' }[
  process.platform
]
if (!host) throw new Error('Unsupported Android build host.')
const executable = process.platform === 'win32' ? '.exe' : ''
const clang = join(
  ndk,
  'toolchains',
  'llvm',
  'prebuilt',
  `${host}-x86_64`,
  'bin',
  `clang${executable}`,
)
const java = process.env.JAVA_HOME
  ? join(process.env.JAVA_HOME, 'bin', `java${executable}`)
  : 'java'

const work = resolve('.work')
mkdirSync(work, { recursive: true })
const project = mkdtempSync(join(work, 'android-project-'))
const app = join(project, 'app')
const main = join(app, 'src/main')
const overlay = join(project, 'overlay.json')
const version = readFileSync('build/config.yml', 'utf8').match(
  /^  version: "(\d+\.\d+\.\d+)"/m,
)?.[1]
if (!version) throw new Error('Could not read the FitTrack version.')
const [major, minor, patch] = version.split('.').map(Number)
const versionCode = major * 1_000_000 + minor * 1_000 + patch
if (
  minor > 999 ||
  patch > 999 ||
  versionCode < 1 ||
  versionCode > 2_100_000_000
) {
  throw new Error(
    'The FitTrack version cannot be represented as an Android version code.',
  )
}

const env = { ...process.env, ANDROID_HOME: sdk }
if (env.ANDROID_KEYSTORE_FILE)
  env.ANDROID_KEYSTORE_FILE = resolve(env.ANDROID_KEYSTORE_FILE)
if (
  process.env.GITHUB_REF_TYPE === 'tag' &&
  ![
    'ANDROID_KEYSTORE_FILE',
    'ANDROID_KEYSTORE_PASSWORD',
    'ANDROID_KEY_ALIAS',
    'ANDROID_KEY_PASSWORD',
  ].every((key) => env[key])
) {
  throw new Error(
    'Tagged Android releases require a persistent release signing key.',
  )
}

function run(command, args, options = {}) {
  return execFileSync(command, args, { stdio: 'inherit', env, ...options })
}

function edit(path, changes) {
  let source = readFileSync(path, 'utf8')
  for (const [before, after] of changes) {
    const next = source.replace(before, after)
    if (next === source)
      throw new Error(`Wails template changed in ${path}: ${before}`)
    source = next
  }
  chmodSync(path, 0o644)
  writeFileSync(path, source)
}

run('go', ['mod', 'download', 'github.com/wailsapp/wails/v3'])
const wails = run(
  'go',
  ['list', '-m', '-f', '{{.Dir}}', 'github.com/wailsapp/wails/v3'],
  {
    stdio: ['ignore', 'pipe', 'inherit'],
    encoding: 'utf8',
  },
).trim()
cpSync(join(wails, 'internal/commands/build_assets/android'), project, {
  recursive: true,
})

const packageDir = join(main, 'java/com/wails/app')
copyFileSync(
  'build/android/FitTrackFiles.java',
  join(packageDir, 'FitTrackFiles.java'),
)
copyFileSync(
  'build/android/data_extraction_rules.xml',
  join(main, 'res/xml/data_extraction_rules.xml'),
)
const iconDir = join(main, 'res/drawable-nodpi')
mkdirSync(iconDir, { recursive: true })
copyFileSync('build/appicon.png', join(iconDir, 'fittrack.png'))
edit(join(main, 'res/values/strings.xml'), [['Wails App', 'FitTrack']])
edit(join(main, 'AndroidManifest.xml'), [
  [
    /<!-- Internet permission for WebView -->[\s\S]*?(?=    <application)/,
    '<uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />\n' +
      '    <uses-permission android:name="android.permission.USE_BIOMETRIC" tools:node="remove" />\n' +
      '    <uses-permission android:name="android.permission.USE_FINGERPRINT" tools:node="remove" />\n\n',
  ],
  [
    'android:allowBackup="true"',
    'android:allowBackup="false"\n        android:fullBackupContent="false"\n        android:dataExtractionRules="@xml/data_extraction_rules"',
  ],
  ['@mipmap/ic_launcher"', '@drawable/fittrack"'],
  ['@mipmap/ic_launcher_round"', '@drawable/fittrack"'],
])
edit(join(packageDir, 'MainActivity.java'), [
  [
    'private WebView webView;',
    'private WebView webView;\n    private FitTrackFiles fitTrackFiles;',
  ],
  [
    'setContentView(R.layout.activity_main);',
    'setContentView(R.layout.activity_main);\n' +
      '        if (Build.VERSION.SDK_INT >= 35) {\n' +
      '            findViewById(R.id.main_container).setOnApplyWindowInsetsListener((view, insets) -> {\n' +
      '                android.graphics.Insets safe = insets.getInsets(android.view.WindowInsets.Type.systemBars()\n' +
      '                    | android.view.WindowInsets.Type.displayCutout() | android.view.WindowInsets.Type.ime());\n' +
      '                view.setPadding(safe.left, safe.top, safe.right, safe.bottom);\n' +
      '                return android.view.WindowInsets.CONSUMED;\n' +
      '            });\n' +
      '        }',
  ],
  [
    'bridge.setWebView(webView);',
    'bridge.setWebView(webView);\n        fitTrackFiles = new FitTrackFiles(this, webView);\n        webView.addJavascriptInterface(fitTrackFiles, "fitTrackFiles");',
  ],
  [
    'public void onPageFinished(WebView view, String url) {',
    'public boolean shouldOverrideUrlLoading(WebView view, WebResourceRequest request) {\n' +
      '                Uri uri = request.getUrl();\n' +
      '                return !WAILS_SCHEME.equals(uri.getScheme()) || !WAILS_HOST.equals(uri.getAuthority());\n' +
      '            }\n\n            @Override\n            public void onPageFinished(WebView view, String url) {',
  ],
  [
    'super.onActivityResult(requestCode, resultCode, data);',
    'super.onActivityResult(requestCode, resultCode, data);\n        if (fitTrackFiles.onActivityResult(requestCode, resultCode, data)) return;',
  ],
])
edit(join(app, 'build.gradle'), [
  ['applicationId = "com.wails.app"', 'applicationId = "dev.vanam.fittrack"'],
  ['minSdk = 21', 'minSdk = 26'],
  ['versionCode = 1', `versionCode = ${versionCode}`],
  ['versionName = "1.0"', `versionName = "${version}"`],
])

mkdirSync('frontend/dist', { recursive: true })
// The embed directive needs a matching file before bindings can be generated.
writeFileSync('frontend/dist/.wails-build', '')
// The service API is shared. Generate it without loading a native toolchain.
run(
  'wails3',
  [
    'generate',
    'bindings',
    '-clean=true',
    '-ts',
    '-d',
    'frontend/bindings',
  ],
  {
    env: { ...env, GOOS: 'windows', GOARCH: 'amd64', CGO_ENABLED: '0' },
  },
)
run(process.execPath, ['node_modules/vite/bin/vite.js', 'build'], {
  cwd: 'frontend',
})
run('wails3', [
  'android',
  'overlay:gen',
  '-out',
  overlay,
  '-config',
  'build/config.yml',
])

for (const [arch, abi, target] of [
  ['arm64', 'arm64-v8a', 'aarch64-linux-android26'],
  ['amd64', 'x86_64', 'x86_64-linux-android26'],
]) {
  const library = join(main, 'jniLibs', abi, 'libwails.so')
  run(
    'go',
    [
      'build',
      '-buildmode=c-shared',
      '-overlay',
      overlay,
      '-tags',
      'production,android,sqlite_omit_load_extension',
      '-trimpath',
      '-buildvcs=false',
      '-ldflags=-w -s',
      '-o',
      library,
      '.',
    ],
    {
      env: {
        ...env,
        GOOS: 'android',
        GOARCH: arch,
        CGO_ENABLED: '1',
        CC: `"${clang}" --target=${target}`,
      },
    },
  )
}

run(
  java,
  [
    '-classpath',
    'gradle/wrapper/gradle-wrapper.jar',
    'org.gradle.wrapper.GradleWrapperMain',
    '--no-daemon',
    'assembleRelease',
  ],
  { cwd: project },
)

const apk = join(app, 'build/outputs/apk/release/app-release.apk')
const buildTools = join(sdk, 'build-tools', '36.0.0')
run(join(buildTools, `zipalign${executable}`), ['-c', '-P', '16', '4', apk])
run(java, [
  '-jar',
  join(buildTools, 'lib/apksigner.jar'),
  'verify',
  '--print-certs',
  apk,
])
const manifest = run(
  join(buildTools, `aapt${executable}`),
  ['dump', 'badging', apk],
  {
    stdio: ['ignore', 'pipe', 'inherit'],
    encoding: 'utf8',
  },
)
if (
  manifest.includes('android.permission.INTERNET') ||
  !manifest.includes("package: name='dev.vanam.fittrack'") ||
  !manifest.includes("native-code: 'arm64-v8a' 'x86_64'")
) {
  throw new Error(
    'The Android package failed its permission, application ID, or ABI check.',
  )
}
mkdirSync('bin', { recursive: true })
const output = `bin/FitTrack-${version}-Android.apk`
copyFileSync(apk, output)
console.log(output)
