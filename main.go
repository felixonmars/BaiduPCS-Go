package main

import (
	"fmt"
	"github.com/gobs/args"
	"github.com/iikira/BaiduPCS-Go/command"
	"github.com/iikira/BaiduPCS-Go/config"
	"github.com/iikira/BaiduPCS-Go/util"
	"github.com/iikira/BaiduPCS-Go/web"
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
	historyFile = pcsutil.ExecutablePathJoin("pcs_command_history.txt")
	reloadFn    = func(c *cli.Context) error {
		baidupcscmd.ReloadIfInConsole()
		return nil
	}
)

func init() {
	pcsconfig.Init()
	baidupcscmd.ReloadInfo()
}

func main() {
	app := cli.NewApp()
	app.Name = "BaiduPCS-Go"
	app.Version = "beta-v3.1"
	app.Author = "iikira/BaiduPCS-Go: https://github.com/iikira/BaiduPCS-Go"
	app.Usage = "百度网盘工具箱 for " + runtime.GOOS + "/" + runtime.GOARCH
	app.Description = `BaiduPCS-Go 使用 Go语言编写, 为操作百度网盘, 提供实用功能.
	具体功能, 参见 COMMANDS 列表

	特色:
		网盘内列出文件和目录, 支持通配符匹配路径;
		下载网盘内文件, 支持网盘内目录 (文件夹) 下载, 支持多个文件或目录下载, 支持断点续传和高并发高速下载.

	程序目前处于测试版, 后续会添加更多的实用功能.
	
	---------------------------------------------------
	前往 https://github.com/iikira/BaiduPCS-Go/releases 以获取程序更新信息!
	---------------------------------------------------`
	app.Action = func(c *cli.Context) {
		if c.NArg() != 0 {
			fmt.Printf("未找到命令: %s\n运行命令 %s help 获取帮助\n", c.Args().Get(0), app.Name)
			return
		}
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
	}

	app.Commands = []cli.Command{
		{
			Name:     "web",
			Usage:    "启用 web 客户端 (测试中)",
			Category: "其他",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				fmt.Println("web 客户端功能为实验性功能, 测试中, 打开 http://localhost:8080 查看效果")
				fmt.Println(pcsweb.StartServer())
				return nil
			},
		},
		{
			Name:  "login",
			Usage: "使用百度BDUSS登录百度账号",
			Description: fmt.Sprintf("\n   示例: \n\n      %s\n\n      %s\n\n   %s\n\n      %s\n\n      %s\n",
				filepath.Base(os.Args[0])+" login --bduss=123456789",
				filepath.Base(os.Args[0])+" login",
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
			Name:    "su",
			Aliases: []string{"chuser"}, // 兼容旧版本
			Usage:   "切换已登录的百度帐号",
			Description: fmt.Sprintf("%s\n   示例:\n\n      %s\n      %s\n",
				"如果运行该条命令没有提供参数, 程序将会列出所有的百度帐号, 供选择切换",
				filepath.Base(os.Args[0])+" su --uid=123456789",
				filepath.Base(os.Args[0])+" su",
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
				filepath.Base(os.Args[0])+" logout --uid=123456789",
				filepath.Base(os.Args[0])+" logout",
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
			Name:     "loglist",
			Usage:    "获取当前帐号, 和所有已登录的百度帐号",
			Category: "百度帐号操作",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				fmt.Printf("\n当前帐号 uid: %d, 用户名: %s\n", pcsconfig.ActiveBaiduUser.UID, pcsconfig.ActiveBaiduUser.Name)
				fmt.Println(pcsconfig.Config.GetAllBaiduUser())
				return nil
			},
		},
		{
			Name:     "quota",
			Usage:    "获取配额, 即获取网盘总空间, 和已使用空间",
			Category: "网盘操作",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				baidupcscmd.RunGetQuota()
				return nil
			},
		},
		{
			Name:      "cd",
			Category:  "网盘操作",
			Usage:     "切换工作目录",
			UsageText: fmt.Sprintf("%s cd <目录 绝对路径或相对路径>", filepath.Base(os.Args[0])),
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
			Usage:     "列出当前工作目录内的文件和目录 或 指定目录内的文件和目录",
			UsageText: fmt.Sprintf("%s ls <目录 绝对路径或相对路径>", filepath.Base(os.Args[0])),
			Category:  "网盘操作",
			Before:    reloadFn,
			Action: func(c *cli.Context) error {
				baidupcscmd.RunLs(c.Args().Get(0))
				return nil
			},
		},
		{
			Name:      "pwd",
			Usage:     "输出当前所在目录 (工作目录)",
			UsageText: fmt.Sprintf("%s pwd", filepath.Base(os.Args[0])),
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
			UsageText: fmt.Sprintf("%s meta <文件/目录 绝对路径或相对路径>", filepath.Base(os.Args[0])),
			Category:  "网盘操作",
			Before:    reloadFn,
			Action: func(c *cli.Context) error {
				baidupcscmd.RunGetMeta(c.Args().Get(0))
				return nil
			},
		},
		{
			Name:      "rm",
			Usage:     "删除 单个/多个 文件/目录",
			UsageText: fmt.Sprintf("%s rm <网盘文件或目录的路径1> <文件或目录2> <文件或目录3> ...", filepath.Base(os.Args[0])),
			Description: fmt.Sprintf("\n   %s\n",
				"注意: 删除多个文件和目录时, 请确保每一个文件和目录都存在, 否则删除操作会失败.",
			),
			Category: "网盘操作",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				baidupcscmd.RunRemove(getSubArgs(c)...)
				return nil
			},
		},
		{
			Name:      "mkdir",
			Usage:     "创建目录",
			UsageText: fmt.Sprintf("%s mkdir <目录 绝对路径或相对路径> ...", filepath.Base(os.Args[0])),
			Category:  "网盘操作",
			Before:    reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				baidupcscmd.RunMkdir(c.Args().Get(0))
				return nil
			},
		},
		{
			Name:  "cp",
			Usage: "拷贝(复制) 文件/目录",
			UsageText: fmt.Sprintf(
				"%s cp <文件/目录> <目标 文件/目录>\n   %s cp <文件/目录1> <文件/目录2> <文件/目录3> ... <目标目录>",
				filepath.Base(os.Args[0]),
				filepath.Base(os.Args[0]),
			),
			Category: "网盘操作",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() <= 1 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				baidupcscmd.RunCopy(getSubArgs(c)...)
				return nil
			},
		},
		{
			Name:  "mv",
			Usage: "移动/重命名 文件/目录",
			UsageText: fmt.Sprintf(
				"移动\t: %s mv <文件/目录1> <文件/目录2> <文件/目录3> ... <目标目录>\n   重命名: %s mv <文件/目录> <重命名的文件/目录>",
				filepath.Base(os.Args[0]),
				filepath.Base(os.Args[0]),
			),
			Category: "网盘操作",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() <= 1 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				baidupcscmd.RunMove(getSubArgs(c)...)
				return nil
			},
		},
		{
			Name:        "download",
			Aliases:     []string{"d"},
			Usage:       "下载文件或目录",
			UsageText:   fmt.Sprintf("%s download <网盘文件或目录的路径1> <文件或目录2> <文件或目录3> ...", filepath.Base(os.Args[0])),
			Description: "下载的文件将会保存到, 程序所在目录的 download/ 目录 (文件夹).\n   已支持目录下载.\n   已支持多个文件或目录下载.\n   自动跳过下载重名的文件! \n",
			Category:    "网盘操作",
			Before:      reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				baidupcscmd.RunDownload(getSubArgs(c)...)
				return nil
			},
		},
		{
			Name:        "upload",
			Aliases:     []string{"u"},
			Usage:       "上传文件或目录",
			UsageText:   fmt.Sprintf("%s upload <本地文件或目录的路径1> <文件或目录2> <文件或目录3> ... <网盘的目标目录>", filepath.Base(os.Args[0])),
			Description: "上传的文件将会保存到 网盘的目标目录.\n   遇到同名文件将会自动覆盖! \n",
			Category:    "网盘操作",
			Before:      reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() <= 1 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				subArgs := getSubArgs(c)

				baidupcscmd.RunUpload(subArgs[:c.NArg()-1], subArgs[c.NArg()-1])
				return nil
			},
		},
		{
			Name:      "set",
			Usage:     "设置配置",
			UsageText: fmt.Sprintf("%s set OptionName Value", filepath.Base(os.Args[0])),
			Description: `
可设置的值:

	OptionName		Value
	------------------------------------------------------
	appid	baidupcs应用ID, 没问题不要修改!

	user_agent	浏览器标识
	cache_size	下载缓存, 如果硬盘占用高, 请尝试调大此值, 建议值 ( 1024 ~ 16384 )
	max_parallel	下载最大线程 (并发量) - 建议值 ( 50 ~ 500 )
	savedir	下载文件的储存目录

例子:

	set appid 260149
	set cache_size 2048
	set max_parallel 250
	set savedir D:\\download
`,
			Category: "配置",
			Before:   reloadFn,
			After:    reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() < 2 || c.Args().Get(1) == "" {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				err := pcsconfig.Config.Set(c.Args().Get(0), c.Args().Get(1)) // 设置
				if err != nil {
					fmt.Println(err)
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

// getSubArgs 获取子命令参数
func getSubArgs(c *cli.Context) (sargs []string) {
	for i := 0; c.Args().Get(i) != ""; i++ {
		sargs = append(sargs, c.Args().Get(i))
	}
	return
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
