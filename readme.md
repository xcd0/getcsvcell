# getcsvcell

csvからちょっとデータを取りたいときに使う。  

## usage

```sh
$ ./getcsvcell.exe 
Usage: getcsvcell [--input INPUT] [--row ROW] [--col COL] [--grep GREP] [--debug] [--version] <command> [<args>]

Options:
  --input INPUT, -i INPUT
                         入力csvファイル。
  --row ROW, -r ROW      行番号指定。0始まり。無効値の時無視される。 [default: -1]
  --col COL, -c COL      列番号指定。0始まり。無効値の時無視される。 [default: -1]
  --grep GREP, -g GREP   文字列検索。grepする。-c,--colとは組み合わせられる。この指定があるとき-r,--rowの指定は無視される。
  --debug, -d            デバッグ用。ログが詳細になる。
  --version, -v          バージョン情報を出力する。
  --help, -h             display this help and exit

Commands:
  version                バージョン情報を出力する。
```

## install

```sh
go install github.com/xcd0/getcsvcell@latest
```

