# util-cli

`util-cli`は汎用的なコマンドラインユーティリティです。

## コマンドの使用

リリースページから最新バージョンのツールをダウンロードしてください。  
その後、コマンドラインで実行できます。Linux環境の例を以下に示します。  
システムのアーキテクチャに応じて、いずれかを選択してください。

```bash
https://github.com/YamabikoLab/util-cli/releases
```

### ubuntu環境での設定
tar.gzファイルをダウンロードして解凍してください。

```bash
tar -xvzf util-cli-v*.*.*-linux-arm64.tar.gz
```

bashrcに以下の設定を追加してください。

```bash
export PATH=$PATH:/path/to/ut
```
sourceコマンドを実行して設定を反映してください。

```bash 
source ~/.bashrc
```

initコマンドを実行して設定ファイルを作成してください。

```bash
ut init
``` 


## Egrepサブコマンド

Egrepサブコマンドは指定されたキーワードと一致するプロジェクト中のコードを検索し、キーワードごとにExcelシートに出力します。

```yaml
egrep:
  keywords:
    - something
  options: -iran
  regex: '\.{key}|\.try\(.*:{key}|\.try\!\(.*:{key}'
  exclusions:
    directories:
      - spec
      - tmp
      - .git
      - db
      - virtualbox_sidekiq_test
      - log
      - public
      - vendor
    files:
      - webpack_admin.js
  targetDir: ./
  output:
    excel:
      filename: 'EgrepResults.xlsx'
      sheet:
        nameLimit: 31
```

`~/.util-cli/config.yml` ファイルを編集して egrep サブコマンドの設定を行ってください。