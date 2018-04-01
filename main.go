package main

import (
	"encoding/hex"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/internal/pcscommand"
	"github.com/iikira/BaiduPCS-Go/internal/pcsconfig"
	"github.com/iikira/BaiduPCS-Go/internal/pcsweb"
	"github.com/iikira/BaiduPCS-Go/pcscache"
	"github.com/iikira/BaiduPCS-Go/pcsliner"
	"github.com/iikira/BaiduPCS-Go/pcstable"
	"github.com/iikira/BaiduPCS-Go/pcsutil"
	"github.com/iikira/BaiduPCS-Go/pcsverbose"
	"github.com/iikira/BaiduPCS-Go/requester"
	"github.com/iikira/args"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

var (
	// Version 版本号
	Version = "v3.3.2"

	historyFilePath = pcsutil.ExecutablePathJoin("pcs_command_history.txt")
	reloadFn        = func(c *cli.Context) error {
		pcscommand.ReloadIfInConsole()
		return nil
	}
)

func init() {
	pcsconfig.Init()
	pcscommand.ReloadInfo()

	// 启动缓存回收
	pcscache.DirCache.GC()
	requester.TCPAddrCache.GC()
}

func main() {
	app := cli.NewApp()
	app.Name = "BaiduPCS-Go"
	app.Version = Version
	app.Author = "iikira/BaiduPCS-Go: https://github.com/iikira/BaiduPCS-Go"
	app.Usage = "百度网盘客户端 for " + runtime.GOOS + "/" + runtime.GOARCH
	app.Description = `BaiduPCS-Go 使用 Go语言编写, 为操作百度网盘, 提供实用功能.
	具体功能, 参见 COMMANDS 列表

	特色:
		网盘内列出文件和目录, 支持通配符匹配路径;
		下载网盘内文件, 支持网盘内目录 (文件夹) 下载, 支持多个文件或目录下载, 支持断点续传和高并发高速下载.
	
	---------------------------------------------------
	前往 https://github.com/iikira/BaiduPCS-Go 以获取更多帮助信息!
	前往 https://github.com/iikira/BaiduPCS-Go/releases 以获取程序更新信息!
	---------------------------------------------------`

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "verbose",
			Usage:       "启用调试",
			EnvVar:      "BAIDUPCS_GO_VERBOSE",
			Destination: &pcsverbose.IsVerbose,
		},
	}
	app.Action = func(c *cli.Context) {
		if c.NArg() != 0 {
			fmt.Printf("未找到命令: %s\n运行命令 %s help 获取帮助\n", c.Args().Get(0), app.Name)
			return
		}
		cli.ShowAppHelp(c)
		pcsverbose.Verbosef("这是一条调试信息\n\n")

		line := pcsliner.NewLiner()

		var err error
		line.History, err = pcsliner.NewLineHistory(historyFilePath)
		if err != nil {
			fmt.Printf("警告: 读取历史命令文件错误, %s\n", err)
		}

		line.ReadHistory()
		defer func() {
			line.DoWriteHistory()
			line.Close()
		}()

		// tab 自动补全命令
		line.State.SetCompleter(func(line string) (s []string) {
			cmds := cli.CommandsByName(app.Commands)

			for k := range cmds {
				if !strings.HasPrefix(cmds[k].FullName(), line) {
					continue
				}
				s = append(s, cmds[k].FullName()+" ")
			}
			return s
		})

		for {
			var (
				prompt          string
				activeBaiduUser = pcsconfig.Config.MustGetActive()
			)

			if activeBaiduUser.Name != "" {
				// 格式: BaiduPCS-Go:<工作目录> <百度ID>$
				// 工作目录太长的话会自动缩略
				prompt = app.Name + ":" + pcsutil.ShortDisplay(path.Base(activeBaiduUser.Workdir), 20) + " " + activeBaiduUser.Name + "$ "
			} else {
				// BaiduPCS-Go >
				prompt = app.Name + " > "
			}

			commandLine, err := line.State.Prompt(prompt)
			if err != nil {
				fmt.Println(err)
				return
			}

			line.State.AppendHistory(commandLine)

			cmdArgs := args.GetArgs(commandLine)
			if len(cmdArgs) == 0 {
				continue
			}

			s := []string{os.Args[0]}
			s = append(s, cmdArgs...)

			// 恢复原始终端状态
			// 防止运行命令时程序被结束, 终端出现异常
			line.Pause()

			c.App.Run(s)

			line.Resume()
		}
	}

	app.Commands = []cli.Command{
		{
			Name:     "web",
			Usage:    "启用 web 客户端 (测试中)",
			Category: "其他",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				fmt.Printf("web 客户端功能为实验性功能, 测试中, 打开 http://localhost:%d 查看效果\n", c.Uint("port"))
				fmt.Println(pcsweb.StartServer(c.Uint("port")))
				return nil
			},
			Flags: []cli.Flag{
				cli.UintFlag{
					Name:  "port",
					Usage: "自定义端口",
					Value: 8080,
				},
			},
		},
		{
			Name:     "run",
			Usage:    "执行系统命令",
			Category: "其他",
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				cmd := exec.Command(c.Args().First(), c.Args().Tail()...)
				cmd.Stdout = os.Stdout
				cmd.Stdin = os.Stdin
				cmd.Stderr = os.Stderr

				err := cmd.Run()
				if err != nil {
					fmt.Println(err)
				}

				return nil
			},
		},
		{
			Name:      "login",
			Usage:     "登录百度账号",
			UsageText: app.Name + " login [command options]",
			Description: `
	示例:
		BaiduPCS-Go login
		BaiduPCS-Go login --username=liuhua
		BaiduPCS-Go login --bduss=123456789

	常规登录:
		按提示一步一步来即可.

	百度BDUSS获取方法:
		参考这篇 Wiki: https://github.com/iikira/BaiduPCS-Go/wiki/关于-获取百度-BDUSS
		或者百度搜索: 获取百度BDUSS`,
			Category: "百度帐号",
			Before:   reloadFn,
			After:    reloadFn,
			Action: func(c *cli.Context) error {
				var bduss, ptoken, stoken string
				if c.IsSet("bduss") {
					bduss = c.String("bduss")
				} else if c.NArg() == 0 {
					var err error
					bduss, ptoken, stoken, err = pcscommand.RunLogin(c.String("username"), c.String("password"))
					if err != nil {
						fmt.Println(err)
						return err
					}
				} else {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				username, err := pcsconfig.Config.SetBDUSS(bduss, ptoken, stoken)
				if err != nil {
					fmt.Println(err)
					return nil
				}

				fmt.Println("百度帐号登录成功:", username)
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "username",
					Usage: "登录百度帐号的用户名(手机号/邮箱/用户名)",
				},
				cli.StringFlag{
					Name:  "password",
					Usage: "登录百度帐号的用户名的密码",
				},
				cli.StringFlag{
					Name:  "bduss",
					Usage: "使用百度 BDUSS 来登录百度帐号",
				},
			},
		},
		{
			Name:    "su",
			Aliases: []string{"chuser"}, // 兼容旧版本
			Usage:   "切换已登录的百度帐号",
			Description: fmt.Sprintf("%s\n   示例:\n\n      %s\n      %s\n",
				"如果运行该条命令没有提供参数, 程序将会列出所有的百度帐号, 供选择切换",
				app.Name+" su <uid>",
				app.Name+" su",
			),
			Category: "百度帐号",
			Before:   reloadFn,
			After:    reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() >= 2 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				if len(pcsconfig.Config.BaiduUserList) == 0 {
					fmt.Printf("未设置任何百度帐号, 不能切换\n")
					return nil
				}

				var uid uint64
				if c.NArg() == 1 {
					uid, _ = strconv.ParseUint(c.Args().Get(0), 10, 64)
					if !pcsconfig.Config.CheckUIDExist(uid) {
						fmt.Printf("切换用户失败, uid 不存在\n")
						return nil
					}
				} else if c.NArg() == 0 {
					cli.HandleAction(app.Command("loglist").Action, c)

					// 提示输入 index
					var index string
					fmt.Printf("输入要切换帐号的 # 值 > ")
					_, err := fmt.Scanln(&index)
					if err != nil {
						return nil
					}

					if n, err := strconv.Atoi(index); err == nil && n >= 0 && n < len(pcsconfig.Config.BaiduUserList) {
						uid = pcsconfig.Config.BaiduUserList[n].UID
					} else {
						fmt.Printf("切换用户失败, 请检查 # 值是否正确\n")
						return nil
					}
				} else {
					cli.ShowCommandHelp(c, c.Command.Name)
				}

				pcsconfig.Config.BaiduActiveUID = uid
				if err := pcsconfig.Config.Save(); err != nil {
					fmt.Printf("%s\n", err)
					return nil
				}

				fmt.Printf("切换用户成功, %s\n", pcsconfig.Config.MustGetActive().Name)
				return nil
			},
		},
		{
			Name:     "logout",
			Usage:    "退出当前登录的百度帐号",
			Category: "百度帐号",
			Before:   reloadFn,
			After:    reloadFn,
			Action: func(c *cli.Context) error {
				if len(pcsconfig.Config.BaiduUserList) == 0 {
					fmt.Println("未设置任何百度帐号, 不能退出")
					return nil
				}

				var (
					au      = pcsconfig.Config.MustGetActive()
					confirm string
				)

				if !c.Bool("y") {
					fmt.Printf("确认退出百度帐号: %s ? (y/n) > ", au.Name)
					_, err := fmt.Scanln(&confirm)
					if err != nil || (confirm != "y" && confirm != "Y") {
						return err
					}
				}

				err := pcsconfig.Config.DeleteBaiduUserByUID(au.UID)
				if err != nil {
					fmt.Printf("退出用户 %s, 失败, 错误: %s\n", au.Name, err)
				}

				fmt.Printf("退出用户成功, %s\n", au.Name)
				return nil
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "y",
					Usage: "确认退出帐号",
				},
			},
		},
		{
			Name:     "loglist",
			Usage:    "获取当前帐号, 和所有已登录的百度帐号",
			Category: "百度帐号",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				au := pcsconfig.Config.MustGetActive()

				fmt.Printf("\n当前帐号 uid: %d, 用户名: %s\n\n", au.UID, au.Name)

				fmt.Println(pcsconfig.Config.BaiduUserList.String())

				return nil
			},
		},
		{
			Name:     "quota",
			Usage:    "获取配额, 即获取网盘的总储存空间, 和已使用的储存空间",
			Category: "百度网盘",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				pcscommand.RunGetQuota()
				return nil
			},
		},
		{
			Name:      "cd",
			Category:  "百度网盘",
			Usage:     "切换工作目录",
			UsageText: app.Name + "%s cd <目录 绝对路径或相对路径>",
			Before:    reloadFn,
			After:     reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				pcscommand.RunChangeDirectory(c.Args().Get(0), c.Bool("l"))

				return nil
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "l",
					Usage: "切换工作目录后自动列出工作目录下的文件和目录",
				},
			},
		},
		{
			Name:      "ls",
			Aliases:   []string{"l", "ll"},
			Usage:     "列出当前工作目录内的文件和目录 或 指定目录内的文件和目录",
			UsageText: fmt.Sprintf("%s ls <目录 绝对路径或相对路径>", app.Name),
			Category:  "百度网盘",
			Before:    reloadFn,
			Action: func(c *cli.Context) error {
				pcscommand.RunLs(c.Args().Get(0))
				return nil
			},
		},
		{
			Name:      "pwd",
			Usage:     "输出当前所在目录 (工作目录)",
			UsageText: fmt.Sprintf("%s pwd", app.Name),
			Category:  "百度网盘",
			Before:    reloadFn,
			Action: func(c *cli.Context) error {
				fmt.Println(pcsconfig.Config.MustGetActive().Workdir)
				return nil
			},
		},
		{
			Name:      "meta",
			Usage:     "获取单个文件/目录的元信息 (详细信息)",
			UsageText: fmt.Sprintf("%s meta <文件/目录 绝对路径或相对路径>", app.Name),
			Category:  "百度网盘",
			Before:    reloadFn,
			Action: func(c *cli.Context) error {
				pcscommand.RunGetMeta(c.Args().Get(0))
				return nil
			},
		},
		{
			Name:      "rm",
			Usage:     "删除 单个/多个 文件/目录",
			UsageText: fmt.Sprintf("%s rm <网盘文件或目录的路径1> <文件或目录2> <文件或目录3> ...", app.Name),
			Description: fmt.Sprintf("\n   %s\n   %s\n",
				"注意: 删除多个文件和目录时, 请确保每一个文件和目录都存在, 否则删除操作会失败.",
				"被删除的文件或目录可在网盘文件回收站找回.",
			),
			Category: "百度网盘",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				pcscommand.RunRemove(c.Args()...)
				return nil
			},
		},
		{
			Name:      "mkdir",
			Usage:     "创建目录",
			UsageText: fmt.Sprintf("%s mkdir <目录 绝对路径或相对路径> ...", app.Name),
			Category:  "百度网盘",
			Before:    reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				pcscommand.RunMkdir(c.Args().Get(0))
				return nil
			},
		},
		{
			Name:  "cp",
			Usage: "拷贝(复制) 文件/目录",
			UsageText: fmt.Sprintf(
				"%s cp <文件/目录> <目标 文件/目录>\n   %s cp <文件/目录1> <文件/目录2> <文件/目录3> ... <目标目录>",
				app.Name,
				app.Name,
			),
			Category: "百度网盘",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() <= 1 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				pcscommand.RunCopy(c.Args()...)
				return nil
			},
		},
		{
			Name:  "mv",
			Usage: "移动/重命名 文件/目录",
			UsageText: fmt.Sprintf(
				"移动\t: %s mv <文件/目录1> <文件/目录2> <文件/目录3> ... <目标目录>\n   重命名: %s mv <文件/目录> <重命名的文件/目录>",
				app.Name,
				app.Name,
			),
			Category: "百度网盘",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() <= 1 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				pcscommand.RunMove(c.Args()...)
				return nil
			},
		},
		{
			Name:      "download",
			Aliases:   []string{"d"},
			Usage:     "下载文件或目录",
			UsageText: fmt.Sprintf("%s download <网盘文件或目录的路径1> <文件或目录2> <文件或目录3> ...", app.Name),
			Description: `下载的文件默认保存到, 程序所在目录的 download/ 目录.
	通过 BaiduPCS-Go config set -savedir <savedir>, 自定义保存的目录.
	已支持目录下载.
	已支持多个文件或目录下载.
	自动跳过下载重名的文件!`,
			Category: "百度网盘",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				pcscommand.RunDownload(c.Bool("test"), c.Int("p"), c.Args())
				return nil
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "test",
					Usage: "测试下载, 此操作不会保存文件到本地",
				},
				cli.IntFlag{
					Name:  "p",
					Usage: "指定下载线程数",
				},
			},
		},
		{
			Name:      "upload",
			Aliases:   []string{"u"},
			Usage:     "上传文件或目录",
			UsageText: fmt.Sprintf("%s upload <本地文件或目录的路径1> <文件或目录2> <文件或目录3> ... <网盘的目标目录>", app.Name),
			Description: `上传的文件将会保存到, 网盘的目标目录.
	遇到同名文件将会自动覆盖!!
	当上传的文件名和网盘的目录名称相同时, 不会覆盖目录, 防止丢失数据.
`,
			Category: "百度网盘",
			Before:   reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() <= 1 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				subArgs := c.Args()

				pcscommand.RunUpload(subArgs[:c.NArg()-1], subArgs[c.NArg()-1])
				return nil
			},
		},
		{
			Name:        "rapidupload",
			Aliases:     []string{"ru"},
			Usage:       "手动秒传文件",
			UsageText:   fmt.Sprintf("%s rapidupload -length=<文件的大小> -md5=<文件的md5值> -slicemd5=<文件前256KB切片的md5值(可选)> -crc32=<文件的crc32值(可选)> <保存的网盘路径, 需包含文件名>", app.Name),
			Description: "上传的文件将会保存到 网盘的目标目录.\n   遇到同名文件将会自动覆盖! \n",
			Category:    "百度网盘",
			Before:      reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() <= 0 || !c.IsSet("md5") || !c.IsSet("length") {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				pcscommand.RunRapidUpload(c.Args().Get(0), c.String("md5"), c.String("slicemd5"), c.String("crc32"), c.Int64("length"))
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "md5",
					Usage: "文件的 md5 值",
				},
				cli.StringFlag{
					Name:  "slicemd5",
					Usage: "文件前 256KB 切片的 md5 值 (可选)",
				},
				cli.StringFlag{
					Name:  "crc32",
					Usage: "文件的 crc32 值 (可选)",
				},
				cli.Int64Flag{
					Name:  "length",
					Usage: "文件的大小",
				},
			},
		},
		{
			Name:        "sumfile",
			Aliases:     []string{"sf"},
			Usage:       "获取文件的秒传信息",
			UsageText:   app.Name + " sumfile <本地文件的路径1> <本地文件的路径2> ...",
			Description: "获取文件的大小, md5, 前256KB切片的md5, crc32, 可用于秒传文件.",
			Category:    "其他",
			Before:      reloadFn,
			Action: func(c *cli.Context) error {
				if c.NArg() <= 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
					return nil
				}

				var (
					fileName, strLength, strMd5, strSliceMd5, strCrc32 string
				)

				for k, filePath := range c.Args() {
					lp, err := pcscommand.GetFileSum(filePath, &pcscommand.SumOption{
						IsMD5Sum:      true,
						IsCRC32Sum:    true,
						IsSliceMD5Sum: true,
					})
					if err != nil {
						fmt.Printf("[%d] %s\n", k+1, err)
						continue
					}

					fmt.Printf("[%d] - [%s]:\n", k+1, filePath)

					strLength, strMd5, strSliceMd5, strCrc32 = strconv.FormatInt(lp.Length, 10), hex.EncodeToString(lp.MD5), hex.EncodeToString(lp.SliceMD5), strconv.FormatUint(uint64(lp.CRC32), 10)
					fileName = filepath.Base(filePath)

					tb := pcstable.NewTable(os.Stdout)
					tb.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
					tb.AppendBulk([][]string{
						[]string{"文件大小", strLength},
						[]string{"md5", strMd5},
						[]string{"前256KB切片的md5", strSliceMd5},
						[]string{"crc32", strCrc32},
						[]string{"秒传命令 (完整)", app.Name + " rapidupload -length=" + strLength + " -md5=" + strMd5 + " -slicemd5=" + strSliceMd5 + " -crc32=" + strCrc32 + " " + fileName},
						[]string{"秒传命令 (精简)", app.Name + " ru -length=" + strLength + " -md5=" + strMd5 + " " + fileName},
					})
					tb.Render()
					fmt.Printf("\n")
				}

				return nil
			},
		},
		{
			Name:        "offlinedl",
			Aliases:     []string{"clouddl", "od"},
			Usage:       "离线下载",
			Description: `支持http/https/ftp/电驴/磁力链协议`,
			Category:    "百度网盘",
			Before:      reloadFn,
			Action: func(c *cli.Context) error {
				if c.NumFlags() <= 0 || c.NArg() <= 0 {
					cli.ShowCommandHelp(c, c.Command.Name)
				}
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:      "add",
					Aliases:   []string{"a"},
					Usage:     "添加离线下载任务",
					UsageText: app.Name + " offlinedl add -path=<离线下载文件保存的路径> 资源地址1 地址2 ...",
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						pcscommand.RunCloudDlAddTask(c.Args(), c.String("path"))
						return nil
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "path",
							Usage: "离线下载文件保存的路径, 默认为工作目录",
						},
					},
				},
				{
					Name:      "query",
					Aliases:   []string{"q"},
					Usage:     "精确查询离线下载任务",
					UsageText: app.Name + " offlinedl query 任务ID1 任务ID2 ...",
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						taskIDs := pcsutil.SliceStringToInt64(c.Args())

						if len(taskIDs) == 0 {
							fmt.Printf("未找到合法的任务ID, task_id\n")
							return nil
						}

						pcscommand.RunCloudDlQueryTask(taskIDs)
						return nil
					},
				},
				{
					Name:      "list",
					Aliases:   []string{"ls", "l"},
					Usage:     "查询离线下载任务列表",
					UsageText: app.Name + " offlinedl list",
					Action: func(c *cli.Context) error {
						pcscommand.RunCloudDlListTask()
						return nil
					},
				},
				{
					Name:      "cancel",
					Aliases:   []string{"c"},
					Usage:     "取消离线下载任务",
					UsageText: app.Name + " offlinedl cancel 任务ID1 任务ID2 ...",
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						taskIDs := pcsutil.SliceStringToInt64(c.Args())

						if len(taskIDs) == 0 {
							fmt.Printf("未找到合法的任务ID, task_id\n")
							return nil
						}

						pcscommand.RunCloudDlCancelTask(taskIDs)
						return nil
					},
				},
				{
					Name:      "delete",
					Aliases:   []string{"del", "d"},
					Usage:     "删除离线下载任务",
					UsageText: app.Name + " offlinedl delete 任务ID1 任务ID2 ...",
					Action: func(c *cli.Context) error {
						if c.NArg() < 1 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						taskIDs := pcsutil.SliceStringToInt64(c.Args())

						if len(taskIDs) == 0 {
							fmt.Printf("未找到合法的任务ID, task_id\n")
							return nil
						}

						pcscommand.RunCloudDlDeleteTask(taskIDs)
						return nil
					},
				},
			},
		},
		{
			// 兼容旧版本
			Name:     "set",
			Usage:    "修改程序配置项",
			Category: "配置",
			Before:   reloadFn,
			After:    reloadFn,
			Action: func(c *cli.Context) error {
				fmt.Printf("请使用 BaiduPCS-Go config set 修改程序配置项\n")
				return nil
			},
			Hidden:   true,
			HideHelp: true,
		},
		{
			Name:        "config",
			Usage:       "显示和修改程序配置项",
			Description: "显示和修改程序配置项",
			Category:    "配置",
			Before:      reloadFn,
			After:       reloadFn,
			Action: func(c *cli.Context) error {
				fmt.Printf("----\n运行 %s config set 可进行设置配置\n\n当前配置:\n", app.Name)
				tb := pcstable.NewTable(os.Stdout)
				tb.SetHeader([]string{"名称", "值", "建议值", "描述"})
				tb.SetColumnAlignment([]int{tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_LEFT})
				tb.AppendBulk([][]string{
					[]string{"appid", fmt.Sprint(pcsconfig.Config.AppID), "", "百度 PCS 应用ID"},
					[]string{"user_agent", pcsconfig.Config.UserAgent, "", "浏览器标识"},
					[]string{"cache_size", strconv.Itoa(pcsconfig.Config.CacheSize), "1024 ~ 262144", "下载缓存, 如果硬盘占用高或下载速度慢, 请尝试调大此值"},
					[]string{"max_parallel", strconv.Itoa(pcsconfig.Config.MaxParallel), "50 ~ 500", "下载最大并发量"},
					[]string{"savedir", pcsconfig.Config.SaveDir, "", "下载文件的储存目录"},
				})
				tb.Render()
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:      "set",
					Usage:     "修改程序配置项",
					UsageText: app.Name + " config set [arguments...]",
					Description: `
	例子:
		BaiduPCS-Go config set -appid=260149
		BaiduPCS-Go config set -user_agent="chrome"
		BaiduPCS-Go config set -cache_size 16384 -max_parallel 200 -savedir D:/download`,
					Action: func(c *cli.Context) error {
						if c.NumFlags() <= 0 || c.NArg() > 0 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						err := pcsconfig.Config.Save()
						if err != nil {
							fmt.Println(err)
							return err
						}

						fmt.Printf("保存配置成功\n")

						return nil
					},
					Flags: []cli.Flag{
						cli.IntFlag{
							Name:        "appid",
							Usage:       "百度 PCS 应用ID",
							Value:       pcsconfig.Config.AppID,
							Destination: &pcsconfig.Config.AppID,
						},
						cli.StringFlag{
							Name:        "user_agent",
							Usage:       "浏览器标识",
							Value:       pcsconfig.Config.UserAgent,
							Destination: &pcsconfig.Config.UserAgent,
						},
						cli.IntFlag{
							Name:        "cache_size",
							Usage:       "下载缓存",
							Value:       pcsconfig.Config.CacheSize,
							Destination: &pcsconfig.Config.CacheSize,
						},
						cli.IntFlag{
							Name:        "max_parallel",
							Usage:       "下载最大并发量",
							Value:       pcsconfig.Config.MaxParallel,
							Destination: &pcsconfig.Config.MaxParallel,
						},
						cli.StringFlag{
							Name:        "savedir",
							Usage:       "下载文件的储存目录",
							Value:       pcsconfig.Config.SaveDir,
							Destination: &pcsconfig.Config.SaveDir,
						},
					},
				},
			},
		},
		{
			Name:  "tool",
			Usage: "工具箱",
			Action: func(c *cli.Context) error {
				cli.ShowCommandHelp(c, c.Command.Name)
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:  "showtime",
					Usage: "显示当前时间(北京时间)",
					Action: func(c *cli.Context) error {
						fmt.Printf(pcsutil.BeijingTimeOption("printLog"))
						return nil
					},
				},
				{
					Name:        "enc",
					Usage:       "加密文件",
					UsageText:   app.Name + " enc -method=<method> -key=<key> [files...]",
					Description: cryptoDescription,
					Action: func(c *cli.Context) error {
						if c.NArg() <= 0 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						for _, filePath := range c.Args() {
							encryptedFilePath, err := pcsutil.EncryptFile(c.String("method"), []byte(c.String("key")), filePath, !c.Bool("disable-gzip"))
							if err != nil {
								fmt.Printf("%s\n", err)
								continue
							}

							fmt.Printf("加密成功, %s -> %s\n", filePath, encryptedFilePath)
						}

						return nil
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "method",
							Usage: "加密方法",
							Value: "aes-128-ctr",
						},
						cli.StringFlag{
							Name:  "key",
							Usage: "加密密钥",
							Value: app.Name,
						},
						cli.BoolFlag{
							Name:  "disable-gzip",
							Usage: "不启用GZIP",
						},
					},
				},
				{
					Name:        "dec",
					Usage:       "解密文件",
					UsageText:   app.Name + " dec -method=<method> -key=<key> [files...]",
					Description: cryptoDescription,
					Action: func(c *cli.Context) error {
						if c.NArg() <= 0 {
							cli.ShowCommandHelp(c, c.Command.Name)
							return nil
						}

						for _, filePath := range c.Args() {
							decryptedFilePath, err := pcsutil.DecryptFile(c.String("method"), []byte(c.String("key")), filePath, !c.Bool("disable-gzip"))
							if err != nil {
								fmt.Printf("%s\n", err)
								continue
							}

							fmt.Printf("解密成功, %s -> %s\n", filePath, decryptedFilePath)
						}

						return nil
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "method",
							Usage: "加密方法",
							Value: "aes-128-ctr",
						},
						cli.StringFlag{
							Name:  "key",
							Usage: "加密密钥",
							Value: app.Name,
						},
						cli.BoolFlag{
							Name:  "disable-gzip",
							Usage: "不启用GZIP",
						},
					},
				},
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

// �
