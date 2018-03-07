package main

import (
	"github.com/urfave/cli"
)

var (
	cryptoDescription = `
	可用的方法 <method>:
		aes-128-ctr, aes-192-ctr, aes-256-ctr,
		aes-128-cfb, aes-192-cfb, aes-256-cfb,
		aes-128-ofb, aes-192-ofb, aes-256-ofb.

	密钥 <key>:
		aes-128 对应key长度为16, aes-192 对应key长度为24, aes-256 对应key长度为32,
		如果key长度不符合, 则自动修剪key, 舍弃超出长度的部分, 长度不足的部分用'\0'填充.

	GZIP <disable-gzip>:
		是否启用GZIP压缩/解压缩文件, 默认启用,
		如果不启用, 则无法检测文件是否解密成功, 解密文件时会保留源文件, 避免解密失败造成文件数据丢失.`
)

func init() {
	cli.AppHelpTemplate = `----
	{{.Name}}{{if .Usage}} - {{.Usage}}{{end}}

USAGE:
	{{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}{{if .Version}}{{if not .HideVersion}}

VERSION:
	{{.Version}}{{end}}{{end}}{{if .Description}}

DESCRIPTION:
	{{.Description}}{{end}}{{if len .Authors}}

AUTHOR{{with $length := len .Authors}}{{if ne 1 $length}}S{{end}}{{end}}:
	{{range $index, $author := .Authors}}{{if $index}}
	{{end}}{{$author}}{{end}}{{end}}{{if .VisibleCommands}}

COMMANDS:{{range .VisibleCategories}}{{if .Name}}
	{{.Name}}:{{end}}{{range .VisibleCommands}}
		{{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{end}}{{if .VisibleFlags}}

GLOBAL OPTIONS:
	{{range $index, $option := .VisibleFlags}}{{if $index}}
	{{end}}{{$option}}{{end}}{{end}}{{if .Copyright}}

COPYRIGHT:
	{{.Copyright}}{{end}}
`

	cli.CommandHelpTemplate = `----
	{{.HelpName}} - {{.Usage}}

USAGE:
	{{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}}{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}{{if .Category}}

CATEGORY:
	{{.Category}}{{end}}{{if .Description}}

DESCRIPTION:
	{{.Description}}{{end}}{{if .VisibleFlags}}

OPTIONS:
	{{range .VisibleFlags}}{{.}}
	{{end}}{{end}}
`

	cli.SubcommandHelpTemplate = `----
	{{.HelpName}} - {{if .Description}}{{.Description}}{{else}}{{.Usage}}{{end}}

USAGE:
	{{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} command{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}

COMMANDS:{{range .VisibleCategories}}{{if .Name}}
	{{.Name}}:{{end}}{{range .VisibleCommands}}
		{{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}
{{end}}{{if .VisibleFlags}}
OPTIONS:
	{{range .VisibleFlags}}{{.}}
	{{end}}{{end}}
`
}
