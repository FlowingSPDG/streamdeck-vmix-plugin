# Active Context

## Current Focus
- Function/Activatorアクションの実装
- ハンドラーベースのアーキテクチャへの移行
- コードの整理と抽象化
- UI/UXの改善

## Recent Changes
1. ハンドラーベースの設計導入
   - VMixConnector
   - PropertyInspectorHandler
   - TallyHandler
   - ShortcutHandler
   - SettingsHandler

2. Function/Activatorアクションの追加
   - FunctionComponent実装
   - バックエンド処理の実装
   - DIの設定

3. 型の整理
   - 共通インターフェースの定義
   - 設定型の整理
   - ハンドラーの型定義

4. UI/UX改善
   - 接続状態表示機能の実装
   - DestinationStatus型の導入
   - PropertyInspectorコンポーネントの改善
   - 接続状態に応じたUI表示の最適化

## Active Decisions
1. ハンドラーベースの設計採用
   - 理由: 機能の組み合わせによる柔軟性
   - 影響: コードの保守性向上、初期実装の複雑化

2. 設定型の配置
   - 理由: 型の重複を避ける
   - 影響: コードの整理、依存関係の明確化

3. エラーハンドリング
   - 方針: ユーザーフレンドリーなエラー表示
   - 実装: ログ出力とUI表示の組み合わせ

4. UI/UX設計方針
   - 接続状態の視覚的フィードバック
   - 一貫性のあるUI表示
   - ユーザーフレンドリーな操作性

## Next Steps
1. テストの実装
   - ユニットテスト
   - 統合テスト
   - E2Eテスト
   - UI/UXテスト

2. ドキュメントの更新
   - API仕様
   - 設定項目
   - エラーケース
   - UI/UXガイドライン

3. パフォーマンス最適化
   - チャネル設定
   - 接続管理
   - メモリ使用
   - UI更新の効率化

## Open Questions
1. エラーハンドリングの改善
   - より詳細なエラー情報の提供
   - エラーリカバリーの方法

2. テスト戦略
   - モックの使用方法
   - テストカバレッジの目標
   - UI/UXテストの方法

3. パフォーマンス目標
   - レイテンシーの要件
   - メモリ使用の制限
   - UI応答性の基準