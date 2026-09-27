# Technical Context

## Core Technologies

### Backend
- Go
  - StreamDeck SDK: github.com/FlowingSPDG/streamdeck
  - vMix SDK: github.com/FlowingSPDG/vmix-go
  - Thread-safe Map: github.com/puzpuzpuz/xsync/v3

### Frontend
- React
- TypeScript
- Tailwind CSS

## Architecture

### Connection Management
- TCP接続を使用したvMixとの通信
- 自動再接続機能
- コンテキストごとの接続管理
- 非同期処理のための各種チャネル

### Action System
- BaseActionによる共通機能の提供
- ハンドラーベースの設計
  - VMixConnector: vMix接続管理
  - PropertyInspectorHandler: UI通信
  - TallyHandler: Tally状態管理
  - ShortcutHandler: ショートカット実行
  - SettingsHandler: 設定管理

### Settings Management
- 型安全な設定管理
- ジェネリクスを活用した共通インターフェース
- コンテキストごとの設定保持

### Frontend Components
- コンポーネントごとの設定UI
- StreamDeck SDKとの通信
- リアルタイムフィードバック

## Development Environment
- Go 1.23
- Node.js
- StreamDeck SDK
- vMix API

## Key Technical Decisions
1. TCP接続の採用
   - 理由: 低レイテンシー、即時反映、TALLY/ACTS機能のサポート
   - トレードオフ: 接続管理の複雑性

2. ハンドラーベースの設計
   - 理由: 機能の組み合わせによる柔軟性、保守性の向上
   - トレードオフ: 初期実装の複雑性

3. React + TypeScriptの採用
   - 理由: 型安全性、コンポーネントの再利用性
   - トレードオフ: ビルド設定の複雑性

## Future Considerations
1. パフォーマンス最適化
   - チャネルバッファサイズの調整
   - 接続管理の効率化

2. エラーハンドリングの改善
   - より詳細なエラー情報
   - ユーザーフレンドリーなエラー表示

3. テスト coverage の向上
   - ユニットテスト
   - 統合テスト
   - E2Eテスト