package main

import (
	"fmt"
	"github.com/gobs/args"
	"github.com/iikira/BaiduPCS-Go/command"
	"github.com/iikira/BaiduPCS-Go/config"
	"github.com/kardianos/osext"
	"github.com/peterh/liner"
	"github.com/urfave/cli"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
)

var (
	historyFile = "pcs_command_history.txt"

	reloadFn = func(c *cli.Context) error {
		baidupcscmd.ReloadIfInConsole()
		return nil
	}
)

func init() {
	// change work directory
	folderPath, err := osext.ExecutableFolder()
	if err != nil {
		folderPath, err = filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			folderPath = filepath.Dir(os.Args[0])
		}
	}
	os.Chdir(folderPath)
}

func main() {
	app := cli.NewApp()
	app.Name = "baidupcs_go"
	app.Author = "iikira/BaiduPCS-Go: https://github.com/iikira/BaiduPCS-Go"
	app.Usage = fmt.Sprintf("百度网盘工具箱 %s/%s GoVersion %s", runtime.GOOS, runtime.GOARCH, runtime.Version())
	app.Description = `baidupcs_go 使用 Go语言编写, 为操作百度网盘, 提供实用功能.
	具体功能, 参见 COMMANDS 列表

	特色:
		网盘内列出文件和目录, 支持通配符匹配路径;
		下载网盘内文件, 支持高并发下载和断点续传.

	程序目前处于测试版, 后续会添加更多的实用功能.`
	app.Version = "beta-v1"
	app.Action = func(c *cli.Context) {
		if c.NArg() == 0 {
			cli.ShowAppHelp(c)

			line := newLiner()
			defer closeLiner(line)

			for {
				if commandLine, err := line.Prompt("BaiduPCS-Go > "); err == nil {
					line.AppendHistory(commandLine)

					cmdArgs := args.GetArgs(commandLine)
					if len(cmdArgs) == 0 {
						continue
					}

					s := []string{os.Args[0]}
					s = append(s, cmdArgs...)

					closeLiner(line)

					c.App.Run(s)

					line = newLiner()

				} else if err == liner.ErrPromptAborted || err == io.EOF {
					break
				} else {
					log.Print("Error reading line: ", err)
					continue
				}
			}
		} else {
			fmt.Printf("未找到命令: %s\n运行命令 %s help 获取帮助\n", c.Args().Get(0), app.Name)
		}
	}

	app.Commands = []cli.Command{
		{
			Name:  "login",
			Usage: "使用百度BDUSS登录百度账号",
			Description: fmt.Sprintf("\n   示例: \n\n      %s\n\n      %s\n\n   %s\n\n      %s\n\n      %s\n",
				app.Name+" login --bduss=123456789",
				app.Name+" login",
				"百度BDUSS获取方法: ",
				"参考这篇 Wiki: https://github.com/iikira/BaiduPCS-Go/wiki/关于-获取百度-BDUSS",
				"或者百度搜索: 获取百度BDUSS",
			),
			Category: "百度帐号操作",
			Before:   reloadFn,
			After:    reloadFn,
			Action: func(c *cli.Context) error {
				bduss := ""
				if c.IsSet("bduss") {
					bduss = c.String("bduss")
				} else if c.NArg() == 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
					line := liner.NewLiner()
					line.SetCtrlCAborts(true)
					defer line.Close()
					bduss, _ = line.Prompt("请输入百度BDUSS值, 回车键提交 > ")
				} else {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				username, err := pcsconfig.Config.SetBDUSS(bduss)
				if err != nil {
					fmt.Println(err)
					return nil
				}

				fmt.Println("百度帐号登录成功:", username)
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "bduss",
					Usage: "百度BDUSS",
				},
			},
		},
		{
			Name:  "chuser",
			Usage: "切换已登录的百度帐号",
			Description: fmt.Sprintf("%s\n   示例:\n\n      %s\n      %s\n",
				"如果运行该条命令没有提供参数, 程序将会列出所有的百度帐号, 供选择切换",
				app.Name+" chuser --uid=123456789",
				app.Name+" chuser",
			),
			Category: "百度帐号操作",
			Before:   reloadFn,
			After:    reloadFn,
			Action: func(c *cli.Context) error {
				if len(pcsconfig.Config.BaiduUserList) == 0 {
					fmt.Println("未设置任何百度帐号, 不能切换")
					return nil
				}

				var uid uint64
				if c.IsSet("uid") {
					if pcsconfig.Config.CheckUIDExist(c.Uint64("uid")) {
						uid = c.Uint64("uid")
					} else {
						fmt.Println("切换用户失败, uid 不存在")
					}
				} else if c.NArg() == 0 {
					cli.HandleAction(app.Command("loglist").Action, c)

					line := liner.NewLiner()
					line.SetCtrlCAborts(true)
					defer line.Close()
					nLine, _ := line.Prompt("请输入要切换帐号的 index 值 > ")

					if n, err := strconv.Atoi(nLine); err == nil && n >= 0 && n < len(pcsconfig.Config.BaiduUserList) {
						uid = pcsconfig.Config.BaiduUserList[n].UID
					} else {
						fmt.Println("切换用户失败, 请检查 index 值是否正确")
					}
				} else {
					cli.ShowCommandHelp(c, c.Command.Name)
				}

				if uid == 0 {
					return nil
				}

				pcsconfig.Config.BaiduActiveUID = uid
				if err := pcsconfig.Config.Save(); err != nil {
					fmt.Println(err)
					return nil
				}

				fmt.Printf("切换用户成功, %v\n", pcsconfig.ActiveBaiduUser.Name)
				return nil

			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "uid",
					Usage: "百度帐号 uid 值",
				},
			},
		},
		{
			Name:  "logout",
			Usage: "退出已登录的百度帐号",
			Description: fmt.Sprintf("%s\n   示例:\n\n      %s\n      %s\n",
				"如果运行该条命令没有提供参数, 程序将会列出所有的百度帐号, 供选择退出",
				app.Name+" logout --uid=123456789",
				app.Name+" logout",
			),
			Category: "百度帐号操作",
			Before:   reloadFn,
			After:    reloadFn,
			Action: func(c *cli.Context) error {
				if len(pcsconfig.Config.BaiduUserList) == 0 {
					fmt.Println("未设置任何百度帐号, 不能退出")
					return nil
				}

				var uid uint64
				if c.IsSet("uid") {
					if pcsconfig.Config.CheckUIDExist(c.Uint64("uid")) {
						uid = c.Uint64("uid")
					} else {
						fmt.Println("退出用户失败, uid 不存在")
					}
				} else if c.NArg() == 0 {
					cli.HandleAction(app.Command("loglist").Action, c)

					line := liner.NewLiner()
					line.SetCtrlCAborts(true)
					defer line.Close()
					nLine, _ := line.Prompt("请输入要退出帐号的 index 值 > ")

					if n, err := strconv.Atoi(nLine); err == nil && n >= 0 && n < len(pcsconfig.Config.BaiduUserList) {
						uid = pcsconfig.Config.BaiduUserList[n].UID
					} else {
						fmt.Println("退出用户失败, 请检查 index 值是否正确")
					}
				} else {
					cli.ShowCommandHelp(c, c.Command.Name)
				}

				if uid == 0 {
					return nil
				}

				// 删除之前先获取被删除的数据, 用于下文输出日志
				baidu, err := pcsconfig.Config.GetBaiduUserByUID(uid)
				if err != nil {
					fmt.Println(err)
					return nil
				}

				if !pcsconfig.Config.DeleteBaiduUserByUID(uid) {
					fmt.Printf("退出用户失败, %s\n", baidu.Name)
				}

				fmt.Printf("退出用户成功, %v\n", baidu.Name)
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "uid",
					Usage: "百度帐号 uid 值",
				},
			},
		},
		{
			Name:      "loglist",
			Usage:     "获取当前帐号, 和所有已登录的百度帐号",
			UsageText: fmt.Sprintf("%s loglist", app.Name),
			Category:  "百度帐号操作",
			Before:    reloadFn,
			Action: func(c *cli.Context) error {
				fmt.Printf("\n当前帐号 uid: %d, 用户名: %s\n", pcsconfig.ActiveBaiduUser.UID, pcsconfig.ActiveBaiduUser.Name)
				fmt.Println(pcsconfig.Config.GetAllBaiduUser())
				return nil
			},
		},
		{
			Name:      "quota",
			Usage:     "获取配额, 即获取网盘总空间, 和已使用空间",
			UsageText: fmt.Sprintf("%s quota", app.Name),
			Category:  "网盘操作",
			Before:    reloadFn,
			Action: func(c *cli.Context) error {
				baidupcscmd.RunGetQuota()
				return nil
			},
		},
		{
			Name:      "cd",
			Usage:     "切换工作目录",
			UsageText: fmt.Sprintf("%s cd <目录>", app.Name),
			Category:  "网盘操作",
			Before:    reloadFn,
			After:     reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}
				baidupcscmd.RunChangeDirectory(c.Args().Get(0))
				return nil
			},
		},
		{
			Name:      "ls",
			Aliases:   []string{"l", "ll"},
			Usage:     "列出当前工作目录的文件和目录或指定目录",
			UsageText: fmt.Sprintf("%s ls <目录>", app.Name),
			Category:  "网盘操作",
			Before:    reloadFn,
			Action: func(c *cli.Context) error {
				baidupcscmd.RunLs(c.Args().Get(0))
				return nil
			},
		},
		{
			Name:      "pwd",
			Usage:     "输出当前所在目录",
			UsageText: fmt.Sprintf("%s pwd", app.Name),
			Category:  "网盘操作",
			Before:    reloadFn,
			Action: func(c *cli.Context) error {
				fmt.Println(pcsconfig.ActiveBaiduUser.Workdir)
				return nil
			},
		},
		{
			Name:      "meta",
			Usage:     "获取单个文件/目录的元信息 (详细信息)",
			UsageText: fmt.Sprintf("%s meta <文件/目录 路径>", app.Name),
			Category:  "网盘操作",
			Before:    reloadFn,
			Action: func(c *cli.Context) error {
				baidupcscmd.RunGetMeta(c.Args().Get(0))
				return nil
			},
		},
		{
			Name:        "download",
			Aliases:     []string{"d"},
			Usage:       "下载文件, 网盘文件绝对路径或相对路径",
			UsageText:   fmt.Sprintf("%s download <网盘文件的路径>", app.Name),
			Description: "下载的文件将会保存到, 程序所在目录的 download/ 目录 (文件夹), 暂不支持指定目录, \n   重名的文件将会被覆盖! \n",
			Category:    "网盘操作",
			Before:      reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}
				baidupcscmd.RunDownload(c.Args().Get(0))
				return nil
			},
		},
		{
			Name:      "set",
			Usage:     "设置配置",
			UsageText: fmt.Sprintf("%s set OptionName Value", app.Name),
			Description: `
可设置的值:

	OptionName		Value
	------------------
	max_parallel	下载最大线程 (并发量) - 建议值 ( 100 ~ 500 )

例子:

	set max_parallel 250
`,
			Category: "配置",
			Before:   reloadFn,
			After:    reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() < 2 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				switch c.Args().Get(0) {
				case "max_parallel":
					parallel, err := strconv.Atoi(c.Args().Get(1))
					if err != nil {
						return cli.NewExitError(fmt.Errorf("max_parallel 设置值不合法, 错误: %s", err), 1)
					}

					pcsconfig.Config.MaxParallel = parallel
					err = pcsconfig.Config.Save()
					if err != nil {
						fmt.Println("设置失败, 错误:", err)
						return nil
					}
					fmt.Printf("设置成功, %s -> %v\n", c.Args().Get(0), c.Args().Get(1))

				default:
					fmt.Printf("未知设定值: %s\n\n", c.Args().Get(0))
					cli.ShowCommandHelp(c, "set")
				}
				return nil
			},
		},
		{
			Name:    "quit",
			Aliases: []string{"exit"},
			Usage:   "退出程序",
			Action: func(c *cli.Context) error {
				return cli.NewExitError("", 0)
			},
			Hidden:   true,
			HideHelp: true,
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Run(os.Args)
}

func newLiner() *liner.State {
	line := liner.NewLiner()

	line.SetCtrlCAborts(true)

	if f, err := os.Open(historyFile); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	return line
}

func closeLiner(line *liner.State) {
	if f, err := os.Create(historyFile); err != nil {
		log.Print("Error writing history file: ", err)
	} else {
		line.WriteHistory(f)
		f.Close()
	}
	line.Close()
}
