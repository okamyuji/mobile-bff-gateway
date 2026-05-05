# Go製モバイルBFFゲートウェイ

## 目的

このリポジトリは、Fintech系モバイルアプリを想定したBFFゲートウェイのサンプルです。
主目的は、複数のバックエンドサービスを1つのモバイル向けJSONに集約し、通信回数とレスポンスサイズを減らすことです。

## 構成

ローカル環境では、ゲートウェイと3つのモックHTTPサービスを起動します。

```text
client
  -> gateway
      -> user-service
      -> payment-service
      -> account-service
```

本番想定では、CloudFront、AWS WAF、ALB、ECS Fargateを使います。
SSL終端はCloudFrontとALBで行います。
ECSサービスは2つ以上のAvailability Zoneで動かし、単一障害点を置かない構成にします。

## Gatewayの責務

GatewayはBearer JWTを検証します。
GatewayはクライアントIP単位でレート制限を行います。
Gatewayは下流サービスを並列に呼び出します。
Gatewayは下流サービスごとにタイムアウト、リトライ、サーキットブレーカーを適用します。
Gatewayはモバイル画面に必要な項目だけを返します。
Gatewayは短いTTLでレスポンスをキャッシュします。

## モックサービス

モックサービスは現実的なFintechの境界にしています。

| サービス | 役割 |
| --- | --- |
| user-service | ユーザーID、表示名、KYC状態を返します。 |
| payment-service | 決済注文の状態と金額を返します。 |
| account-service | 利用可能残高と通貨を返します。 |

モックサービスは固定JSONを返します。
バックエンドの詳細実装を省くことで、Gatewayの構成と責務を理解しやすくしています。

## ローカル起動

次のコマンドで起動します。

```sh
docker compose up --build
```

テスト用JWTは`internal/auth`のテストヘルパーで生成できます。
手動確認をする場合は、同じ形式のHS256 JWTを`Authorization`ヘッダーに指定します。

```sh
curl -H "Authorization: Bearer <token>" http://localhost:8080/mobile/home
```

## 品質ゲート

品質ゲートは次の検証を行います。

```sh
scripts/quality-gate.sh
```

このスクリプトは、`gofmt`、`go vet`、`go test -count=1 -shuffle=on -cover ./...`、`staticcheck`、`golangci-lint`、`go build`を実行します。
pre-commitでは品質ゲートとgitleaksを実行します。
CIでも同じ品質ゲートとgitleaksを実行します。

## Terraform

AWSのIaCはTerraformで書いています。
CloudFormationは使いません。
Terraformは、AWS CLIプロファイル`fintech-apigw`で実行する前提です。
このプロファイルは、IAMユーザー`fintech-apigw`にAccessAdministrator権限が付与されている前提にしています。

Terraformはモジュールに分けています。

| モジュール | 役割 |
| --- | --- |
| `modules/alb` | ACM証明書を使うHTTPS ALBを作成します。 |
| `modules/ecs-service` | ECS FargateのGatewayサービスを作成します。 |
| `modules/edge` | CloudFrontとAWS WAFを作成します。 |

実際に適用する前に、既存VPC、2つ以上のパブリックサブネット、2つ以上のプライベートサブネット、ACM証明書ARN、Gatewayコンテナイメージを指定します。
