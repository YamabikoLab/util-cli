# util-cli

`util-cli`は汎用的なコマンドラインユーティリティです。

## 環境構築

リリースページから最新バージョンのツールをダウンロードしてください。  
システムのアーキテクチャに応じて、いずれかを選択してください。

https://github.com/YamabikoLab/util-cli/releases

### ubuntu環境での設定
tar.gzファイルをダウンロードして解凍してください。

```bash
tar xvzf util-cli-v*.*.*-linux-arm64.tar.gz
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
config.ymlファイルがホームディレクトリの.util-cliに作成されます。

## Egrepサブコマンド

Egrepサブコマンドは指定されたキーワードと一致するプロジェクト中のコードを検索し、キーワードごとにExcelシートに出力します。

```bash
ut egrep
``` 

`~/.util-cli/config.yml` ファイルを編集して egrep サブコマンドの設定を行ってください。

- keywordsに検索したいキーワードを追加してください。
- regexに検索したい正規表現を追加してください。 {key}はキーワードに置換されます。
- キーワードごとにExcelシートに出力されます。
- 実行したコマンドはresultシートに出力されます。

```yaml
egrep:
  keywords:  # 検索対象となるキーワードのリスト。
    - something1
    - something2
  concurrencyLimit: 10  # 同時に実行可能な並行性の制限。数が大きいほど、多くの検索タスクを同時に実行します。
  options: -iran  # egrep コマンドのオプション。詳細は man egrep を参照。
  regex: '\.{key}|\.try\(.*:{key}|\.try\!\(.*:{key}'  # 行がキーワードと一致するかどうかを判断するための正規表現。
  exclusions:  # 除外したいディレクトリやファイルを指定。
    directories: # 検索から除外するディレクトリのリスト。
      - spec
      - tmp
      - .git
      - db
      - virtualbox_sidekiq_test
      - log
      - public
      - vendor
    files:  # 検索から除外するファイルのリスト。
      - webpack_admin.js
  targetDir: ./  # 検索を開始する対象ディレクトリ。
  output: # 出力フォーマットの設定。
    excel:
      enable: true  # Excelファイルへの出力を有効にするかどうか。
      filePath: 'EgrepResults.xlsx'  # 出力のExcelファイル名。
      sheet:
        nameLimit: 31  # Excelシート名の最大文字数の制限。
    text:
      enable: true  # テキストファイルへの出力を有効にするかどうか。
      dirPath: 'egrep_results'  # 出力のテキストファイルのディレクトリ。
```

# 要望・バグ報告
要望やバグ報告は、[GitHub Issues](https://github.com/YamabikoLab/util-cli/issues) にて受け付けています。
