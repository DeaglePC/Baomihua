package cmd

import (
	"fmt"
	"os"
	"strings"

	"baomihua/config"
	"baomihua/executor"
	"baomihua/llm"
	"baomihua/ui"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	modelFlag   string
	listFlag    bool
	switchFlag  string
	installFlag bool
	initFlag    string
)

var Version = "dev"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "bmh [prompt...]",
	Short:   "BaoMiHua (🐆) - 终端 AI 专属指令助手",
	Long:    `豹米花 (BaoMiHua) 能够感知当前操作系统与 Shell 环境，将自然语言转化为精准的 Shell 命令，并在终端中提供无缝的交互执行体验。`,
	Version: Version,
	Args:    cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		llm.InitRegistry()
		forceRefresh, _ := cmd.Flags().GetBool("refresh")

		if installFlag {
			InstallWrapper()
			return
		}

		if initFlag != "" {
			InitWrapper(initFlag)
			return
		}

		if switchFlag != "" {
			err := config.UpdateDefaultModel(switchFlag)
			if err != nil {
				fmt.Printf("❌ 切换默认模型失败: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✅ 默认模型已永久切换为: %s\n", switchFlag)
			return
		}

		if listFlag {
			fmt.Println("⏳ Fetching models from configured vendors...")
			if err := llm.GlobalRegistry.LoadModels(forceRefresh); err != nil {
				fmt.Printf("⚠️ Warning: Could not fetch some models: %v\n", err)
			}

			modelsByVendor := llm.GlobalRegistry.GetModelsList()
			if len(modelsByVendor) == 0 {
				fmt.Println("\n❌ No models found. Please configure API keys for at least one vendor.")
				fmt.Println("Example: export OPENAI_API_KEY=\"sk-...\"")
				return
			}

			fmt.Println("\n📦 Supported models (Configured):")
			for vendor, models := range modelsByVendor {
				fmt.Printf("\n🏢 Vendor: %s\n", strings.ToUpper(vendor))
				for _, m := range models {
					fullModelName := fmt.Sprintf("%s/%s", vendor, m)
					if fullModelName == config.GetModel() || m == config.GetModel() {
						fmt.Printf("  - %s (currently selected)\n", fullModelName)
					} else {
						fmt.Printf("  - %s\n", fullModelName)
					}
				}
			}
			return
		}

		prompt := strings.Join(args, " ")
		if prompt == "" {
			cmd.Help()
			return
		}

		if strings.ToLower(strings.TrimSpace(prompt)) == "fuck" {
			lastCmd := strings.TrimSpace(os.Getenv("BAOMIHUA_LAST_CMD"))
			lastErr := strings.TrimSpace(os.Getenv("BAOMIHUA_LAST_ERROR"))

			if lastErr != "" {
				prompt = fmt.Sprintf("我刚刚运行的命令 `%s` 报错了，错误信息是: `%s`。请判断原因，并给我一个修复后的正确命令。", lastCmd, lastErr)
			} else if lastCmd != "" {
				prompt = fmt.Sprintf("我刚刚运行的命令 `%s` 似乎出错了（或不符合预期）。请帮我检查原因，并提供一个修复后的正确命令。", lastCmd)
			} else {
				prompt = "我刚才执行的命令出错了（或者遇到问题了），请帮我检查并提供正确的解决命令。"
			}
		}

		// Pre-flight check: ensure at least one vendor is configured
		if len(config.GetAllVendors()) == 0 {
			fmt.Println("❌ Error: No API keys configured. Please configure at least one vendor's API key.")
			fmt.Println("Example: export OPENAI_API_KEY=\"sk-...\" or set it in ~/.baomihua/config.yaml")
			os.Exit(1)
		}

		res, action, exitStr, err := ui.RunUI(prompt)
		if err != nil {
			if ui.IsChinese(prompt) {
				fmt.Printf("❌ 发生致命错误: %v\n", err)
			} else {
				fmt.Printf("❌ Fatal error: %v\n", err)
			}
			os.Exit(1)
		}

		if exitStr != "" {
			fmt.Print(exitStr)
		}

		if action == ui.ActionExecute && res != nil {
			fmt.Println()
			err := executor.ExecuteCommand(res.Command, llm.GetEnvContext())
			if err != nil {
				if ui.IsChinese(prompt) {
					fmt.Printf("\n❌ 执行异常: %v\n", err)
				} else {
					fmt.Printf("\n❌ Execution exception: %v\n", err)
				}
				os.Exit(1)
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(config.InitConfig)

	// Define command line flags
	rootCmd.PersistentFlags().StringVarP(&modelFlag, "model", "m", "", "Override the default or configured model")
	rootCmd.Flags().BoolVarP(&listFlag, "list", "l", false, "List supported models")
	rootCmd.Flags().StringVarP(&switchFlag, "switch", "s", "", "Set the default model persistently (format: vendor/model)")
	rootCmd.Flags().BoolP("refresh", "r", false, "Force refresh the models cache from configured vendors")
	rootCmd.Flags().BoolVar(&installFlag, "install", false, "Install the bmh shell wrapper into your terminal profile")
	rootCmd.Flags().StringVar(&initFlag, "init", "", "Generate shell wrapper script for terminal injection (e.g. --init zsh)")

	// Bind flag to viper
	viper.BindPFlag("model", rootCmd.PersistentFlags().Lookup("model"))
}
