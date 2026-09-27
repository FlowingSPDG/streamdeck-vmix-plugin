# StreamDeck vMix Plugin Project Brief

## Overview
StreamDeck用のvMixコントロールプラグインで、vMixの操作をStreamDeckから行えるようにするものです。

## Core Requirements
1. vMixとの接続管理
   - TCP接続によるリアルタイム制御
   - 自動再接続機能
   - 複数vMixインスタンスのサポート

2. アクション
   - Preview: プレビュー入力の切り替え
   - Program: プログラム出力の切り替え（トランジション設定可能）
   - Function: vMixのショートカット機能の実行
   - Activator: vMixのActs状態の監視

3. フィードバック
   - Tally状態の表示
   - Acts状態の表示
   - 接続状態の表示

## Goals
- スタジオ・放送用途で使用可能な安定性の提供
- companionのvMixプラグインと同等以上の機能性
- ネイティブStreamDeckプラグインとしての利点を活かした実装

## Success Criteria
- 放送現場での実運用に耐える安定性
- 低レイテンシーでのvMix制御
- 直感的なUI/UX
- エラー時の適切なフィードバック