# util-cli

`util-cli`はプログラム可能で強力なコマンドラインユーティリティです。

## コマンドの使用

リリースページから最新バージョンのツールをダウンロードしてください。その後、コマンドラインで実行できます。Linux環境の例を以下に示します。

## Egrepサブコマンド

Egrepサブコマンドは指定されたキーワードと一致するプロジェクト中のコードを検索し、キーワードごとにExcelシートに出力します。

```yaml
egrep:
  keywords:
    - something
  options: -iran
  regex: '\.{key}'
  exclusions:
    directories:
      - hoge
      - fuga
    files:
      - piyo.js
  targetDir: ./
  output: 'output.xlsx'
```

`~/.util-cli/config.yml` ファイルを使用して egrep サブコマンドの設定を行います。