package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/pkg/errors"
)

var (
	Version  string = "0.0.2"
	Revision        = func() string {
		revision := ""
		modified := false

		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				if setting.Key == "vcs.revision" {
					//return setting.Value
					revision = setting.Value
					if len(setting.Value) > 7 {
						revision = setting.Value[:7] // 最初の7文字にする
					}
				}
				if setting.Key == "vcs.modified" {
					modified = setting.Value == "true"
				}
			}
		}
		if modified {
			revision = "develop+" + revision
		}
		return revision
	}()
)

type Args struct {
	Csv        string       `arg:"-i,--input"         help:"入力csvファイル。"`
	Row        int          `arg:"-r,--row"           help:"行番号指定。0以上。無効値の時無視される。"`
	Col        int          `arg:"-c,--col"           help:"列番号指定。0以上。無効値の時無視される。"`
	Grep       string       `arg:"-g,--grep"          help:"文字列検索。grepする。-c,--colとは組み合わせられる。この指定があるとき-r,--rowの指定は無視される。"`
	Debug      bool         `arg:"-d,--debug"         help:"デバッグ用。ログが詳細になる。"`
	Version    bool         `arg:"-v,--version"       help:"バージョン情報を出力する。"`
	VersionSub *ArgsVersion `arg:"subcommand:version" help:"バージョン情報を出力する。"`
}
type ArgsVersion struct {
}

func (args *Args) Print() {
	log.Printf(`
Csv  : %v
Row  : %v
Col  : %v
Grep : %v
`, args.Csv, args.Row, args.Col, args.Grep)
}

// ShowHelp() で使う
var parser *arg.Parser

func ShowHelp(post string) {
	buf := new(bytes.Buffer)
	parser.WriteHelp(buf)
	fmt.Printf("%v\n", strings.ReplaceAll(buf.String(), "display this help and exit", "ヘルプを出力する。"))
	if len(post) != 0 {
		fmt.Println(post)
	}
	os.Exit(1)
}
func ShowVersion() {
	if len(Revision) == 0 {
		// go installでビルドされた場合、gitの情報がなくなる。その場合v0.0.0.のように末尾に.がついてしまうのを避ける。
		fmt.Printf("%v version %v\n", GetFileNameWithoutExt(os.Args[0]), Version)
	} else {
		fmt.Printf("%v version %v.%v\n", GetFileNameWithoutExt(os.Args[0]), Version, Revision)
	}
	os.Exit(0)
}

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile) // ログの出力書式を設定する
	args := &Args{
		Csv:  "",
		Row:  -1,
		Col:  -1,
		Grep: "",
	}
	//parser = arg.MustParse(args)
	var err error
	parser, err = arg.NewParser(arg.Config{Program: GetFileNameWithoutExt(os.Args[0]), IgnoreEnv: false}, args)
	if err != nil {
		ShowHelp(fmt.Sprintf("%v", errors.Errorf("%v", err)))
	}
	if err := parser.Parse(os.Args[1:]); err != nil {
		//log.Printf("%#v", err.Error())
		if err.Error() == "help requested by user" {
			//ShowHelp(fmt.Sprintf("%v", errors.Errorf("%v", err)))
			ShowHelp("")
			return
		} else if err.Error() == "version requested by user" {
			ShowVersion()
			return
		} else {
			panic(errors.Errorf("%v", err))
		}
	}

	if args.Version || args.VersionSub != nil {
		ShowVersion()
		return
	}

	if len(args.Csv) == 0 {
		ShowHelp("Error: 入力csvファイルを指定してください。")
	}
	if len(args.Grep) == 0 && args.Row == -1 && args.Col == -1 {
		ShowHelp("Error: 行番号、列番号等を指定してください。")
	}
	if args.Debug {
		args.Print()
	}

	str, err := GetCsvCell(args, args.Csv, args.Row, args.Col, args.Grep)
	if err != nil {
		panic(errors.Errorf("%v", err))
	}

	fmt.Println(str)
}

func GetCsvCell(args *Args, filename string, n int, m int, grep string) (string, error) {
	// CSVファイルを開く
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// CSVリーダーを作成
	reader := csv.NewReader(file)

	// CSVの全データを読み込む
	records, err := reader.ReadAll()
	if err != nil {
		return "", err
	}

	if args.Debug {
		log.Printf("n:%v,m:%v,records:%v", n, m, len(records))
	}

	// 引数で動作を変える。
	if len(grep) != 0 {
		// 検索して含まれていた行のみ取り出す。
		// 指定された列が含まれる行のインデックスを保持するスライス
		var rowsContaining []int
		// CSVファイルを1行ずつ読み込む
		for n, v := range records {
			for _, cell := range v {
				if strings.Contains(cell, grep) {
					rowsContaining = append(rowsContaining, n)
				}
			}
		}
		if len(rowsContaining) == 0 {
			return "", nil
		}
		if m < 0 {
			// 含まれていた行を出力する。
			picked := ""
			for _, n := range rowsContaining {
				for _, row_n_col_m := range records[n] {
					cell := row_n_col_m
					if strings.Contains(cell, "\n") {
						cell = fmt.Sprintf("%#v", cell) // 括弧で括る。
					}
					if len(picked) == 0 {
						picked = cell
					} else {
						picked = fmt.Sprintf("%v,%v", picked, cell)
					}
				}
				picked = fmt.Sprintf("%v\n", picked)
			}
			return picked, nil
		} else {
			// 取り出された行のm列目のみ取り出し、改行区切りで返す。
			picked := ""
			for _, n := range rowsContaining {
				cell := records[n][m]
				if strings.Contains(cell, "\n") {
					cell = fmt.Sprintf("%#v", cell) // 括弧で括る。
				}
				if len(picked) == 0 {
					picked = cell
				} else {
					picked = fmt.Sprintf("%v\n%v", picked, cell)
				}
			}
			return picked, nil
		}
	} else if m < 0 && n >= 0 {
		if n < len(records) {
			// n行目のみ出力する。
			picked := ""
			for _, row_n_col_m := range records[n] {
				cell := row_n_col_m
				if strings.Contains(cell, "\n") {
					cell = fmt.Sprintf("%#v", cell) // 括弧で括る。
				}
				if len(picked) == 0 {
					picked = cell
				} else {
					picked = fmt.Sprintf("%v,%v", picked, cell)
				}
			}
			return picked, nil
		} else {
			return "", fmt.Errorf("指定された行 %v が範囲外です。", n)
		}
	} else if n < 0 && m >= 0 {
		if m < len(records[n]) {
			// m列目のみ出力する。
			picked := ""
			for _, row_n := range records {
				cell := row_n[m]
				if strings.Contains(cell, "\n") {
					cell = fmt.Sprintf("%#v", cell) // 括弧で括る。
				}
				if len(picked) == 0 {
					picked = cell
				} else {
					picked = fmt.Sprintf("%v\n%v", picked, cell)
				}
			}
			return picked, nil
		} else {
			return "", fmt.Errorf("指定された列 %v が範囲外です。", m)
		}
	} else {
		// 指定された行と列がデータの範囲内にあるか確認
		if n < 0 || n >= len(records) {
			return "", fmt.Errorf("指定された行 %v が範囲外です。", n)
		}
		if m < 0 || m >= len(records[n]) {
			return "", fmt.Errorf("指定された列 %v が範囲外です。", m)
		}
		// 指定されたデータを返す
		return records[n][m], nil
	}
}

func GetFileNameWithoutExt(path string) string {
	return filepath.Base(path[:len(path)-len(filepath.Ext(path))])
}
func GetFilePathWithoutExt(path string) string {
	return filepath.ToSlash(filepath.Join(filepath.Dir(path), GetFileNameWithoutExt(path)))
}
