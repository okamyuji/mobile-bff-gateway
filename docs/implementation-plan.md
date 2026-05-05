# Mobile BFF Gateway Implementation Plan

> For agentic workers: REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox syntax for tracking.

Goal: モバイルアプリ向けに小さいJSONを返すGo製BFF Gatewayを実装します。

Architecture: Gatewayは標準ライブラリのnet/httpで実装します。下流のユーザー、決済注文、口座残高サービスはモックHTTPサービスとしてdocker composeで起動します。AWS構成はCloudFront、WAF、ALB、ECS FargateをTerraformモジュールで表現します。

Tech Stack: Go標準ライブラリ、Docker Compose、Terraform、pre-commit、gitleaks、GitHub Actionsを使います。

---

## ファイル構成

- `cmd/gateway/main.go`はGatewayサーバーを起動します。
- `cmd/mockservice/main.go`はユーザー、決済注文、口座残高のモックHTTPサービスを起動します。
- `internal/auth`はBearer JWTの軽量検証を担当します。
- `internal/ratelimit`はIP単位のトークンバケットを担当します。
- `internal/breaker`は下流サービスごとのサーキットブレーカーを担当します。
- `internal/gateway`はHTTPハンドラー、下流呼び出し、集約、キャッシュを担当します。
- `internal/mobile`はモバイル向けレスポンス型を定義します。
- `terraform/modules`はCloudFront、WAF、ALB、ECSのモジュールを定義します。
- `.pre-commit-config.yaml`と`scripts/quality-gate.sh`はローカル品質ゲートを定義します。

## タスク

- [ ] テストを先に作成し、未実装による失敗を確認します。
- [ ] 認証、レート制限、サーキットブレーカーを実装します。
- [ ] Gatewayの`/mobile/home`集約処理とキャッシュを実装します。
- [ ] モックサービスとdocker composeを実装します。
- [ ] Terraformモジュール、pre-commit、CI、READMEを追加します。
- [ ] gofmt、go vet、go test、staticcheck、golangci-lint、go build、gitleaksを実行します。
