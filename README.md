# mono
before: https://github.com/nerikeshi-k/benienma

## 概略
- /_/&lt;blob_name&gt; でGCSにある画像を配信する
- 一度配信した画像はしばらくキャッシュする
- /w=400/&lt;blob_name&gt; で最大横幅が400になるように縮小させて画像を配信する
- /.../&lt;blob_name.png&gt;.webp のように拡張子を追加するように指定するとWebP形式で画像を配信する（他、jpegとpngも）

## ビルド

### 開発版ビルド
`go build`

### リリースビルド
`go build --tags=prod`

## 実行
`./mono --conf=<config_json_path>`  
  
`GOOGLE_APPLICATION_CREDENTIALS` 環境変数にGCSキーのパスが指定されている必要があります。  
詳しくは https://cloud.google.com/docs/authentication/production?hl=ja  

## Docker

### ビルド

```sh
$ docker build . -t mono:latest
```

### 実行
`/etc/*` のexampleファイルを参考に設定ファイルとクレデンシャルを配置してから

```sh
$ docker run -it -v ./etc:/etc/mono mono
```
