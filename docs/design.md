# Mobile BFF Gateway Design

## 目的

このサンプルは、モバイルアプリ向けのAPI GatewayをGoで実装します。
バックエンドの詳細実装はモックHTTPサービスに留め、Gatewayが担う認証、レート制限、SSL終端、集約、キャッシュ、タイムアウト制御、JSON最適化を理解しやすくします。

## 対象トラフィック

この設計はStripe級の大規模決済基盤を対象にしません。
Monzoが公開している通常時1,500 RPS、イベント時4,300 RPSを参考にし、通常500から1,500 RPS、短時間バースト3,000から4,300 RPSを扱える入口設計を目標にします。

## アーキテクチャ

本番想定では、Route 53、CloudFront、AWS WAF、ALB、ECS Fargateを使います。
ECSサービスは2つ以上のAvailability Zoneで起動し、最小タスク数を2以上にします。
Go GatewayはALBの背後で稼働し、ユーザー、決済注文、口座残高の各モックサービスを並列に呼び出します。
SSL終端はCloudFrontとALBで行い、ALBにはACM証明書を関連付けます。

## Gatewayの責務

GatewayはJWT形式のBearerトークンを検証します。
GatewayはクライアントIP単位でトークンバケット型のレート制限を行います。
Gatewayは下流サービスごとにタイムアウトを設定します。
Gatewayは一時的な下流障害に対して限定的なリトライを行います。
Gatewayは連続障害を検知した場合にサーキットブレーカーを開きます。
Gatewayは複数サービスのレスポンスをモバイル画面用の小さいJSONに集約します。
Gatewayは読み取りレスポンスを短いTTLでキャッシュします。

## ローカル構成

ローカルではdocker composeでGatewayと3つのモックサービスを起動します。
モックサービスはHTTPで固定データを返します。
ユーザーサービスはKYC状態を含む会員情報を返します。
決済注文サービスは認可済みと精算済みの決済注文を返します。
口座サービスは利用可能残高と通貨を返します。
Gatewayだけが実装の中心であり、モックサービスは構成理解のために最小限のコードにします。

## 品質ゲート

品質ゲートはgofmt、go vet、go test、staticcheck、golangci-lint、go buildを必須にします。
pre-commitは品質ゲートとgitleaksを実行します。
GitHub Actionsでも同じ品質ゲートとgitleaksを実行します。
