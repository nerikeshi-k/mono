# mono
before: https://github.com/nerikeshi-k/benienma

## 概略
- /_/&lt;blob_name&gt; でGCSにある画像を配信できます
- 一度配信した画像はしばらくキャッシュされます
- /w=400/&lt;blob_name&gt; で最大横幅が400になるように縮小させて画像を配信できます
- 対応はGoogle Cloud Storageのみで、AWSやAzureでは使えません

## ビルド

### 開発版ビルド
`go build`

### リリースビルド
`go build --tags=prod`

## 実行
`./mono --conf=<config_json_path>`  
  
`GOOGLE_APPLICATION_CREDENTIALS` 環境変数にGCSキーのパスが指定されている必要があります。  
詳しくは https://cloud.google.com/docs/authentication/production?hl=ja  
