# fincode JS ガッチャ集

## 1. `ui.mount()` が `#elementId-form` 要素を必要とする

### 症状

```
TypeError: Cannot read properties of null (reading 'setAttribute')
    at Object.mount (fincode.js:1:7089)
```

### 原因

`mount(elementId, width)` は内部で `document.getElementById(elementId + "-form")` を呼び出し、スタイルを適用する。しかし `create()` は iframe の src URL を組み立てるだけで DOM 要素を生成しないため、`#elementId-form` が存在しないと null 参照エラーになる。ドキュメントにこの要件の記載はない。

### 対処

`mount()` を呼び出す前に `#elementId-form` 要素を HTML 側で用意する。

```html
<div id="fincode-ui">
  <div id="fincode-ui-form"></div>
</div>
```

---

## 2. React StrictMode との非互換

### 症状

開発環境でのみ初期化エラーが発生する。

### 原因

React 18 の StrictMode は開発環境でエフェクトを mount → unmount → remount する。`@fincode/js` の `initFincode()` が2回呼ばれると fincode.js 内部の状態が壊れる。

### 対処

`main.tsx` で `<StrictMode>` を外す。

```tsx
// Before
createRoot(root).render(<StrictMode><App /></StrictMode>)

// After
createRoot(root).render(<App />)
```

---

## 3. fincode.js は CDN から直接ロードする

### 原因

`@fincode/js` の `initFincode()` はスクリプトタグを動的に注入するため、StrictMode との組み合わせで問題が起きやすい。

### 対処

`index.html` に CDN スクリプトを直接置き、`window.Fincode` を同期的に呼び出す。

```html
<script src="https://js.test.fincode.jp/v1/fincode.js"></script>
```

```ts
import type { FincodeInstance, FincodeUI } from '@fincode/js';
// window.Fincode の型は別途 fincode.d.ts で宣言
const fc = window.Fincode(publicKey);
```
