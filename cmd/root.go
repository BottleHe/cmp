/*
Copyright © 2022 BottleHe

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	RENAME_FILE    byte = 0x0
	OVERWRITE_FILE byte = 0x1
	IGNORE_FILE    byte = 0x2
)

var (
	SourcePath       string
	DestinationPath  string
	DuplicateControl byte = RENAME_FILE
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cpm [flags] [source] [target]",
	Short: "Copy file ",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.New("参数不正确")
		}
		// 检查是否是正常的文件
		var (
			last                 int32
			err                  error
			sourceStat, destStat os.FileInfo
		)
		SourcePath, err = filepath.Abs(args[0])
		if nil != err {
			return errors.New(fmt.Sprintf("Source: %v, %v", err, SourcePath))
		}
		if sourceStat, err = os.Stat(SourcePath); err != nil {
			if os.IsNotExist(err) {
				return errors.New(fmt.Sprintf("Source: %v, %v", err, SourcePath))
			} else {
				return errors.New(fmt.Sprintf("Source: %v, %v", err, SourcePath))
			}
		}
		// 判断目标地址是否是 文件分割符结尾的
		last = int32(args[1][len(args[1])-1])
		if filepath.Separator == last {
			// 表示是以 路径分割符结尾, 表示应该是一个目录
			_, sourceFileName := filepath.Split(SourcePath)
			DestinationPath = args[1] + sourceFileName
		} else {
			DestinationPath = args[1]
		}
		DestinationPath, err = filepath.Abs(DestinationPath)
		if nil != err {
			// fmt.Errorf("Destination file path error, %v", err)
			return errors.New(fmt.Sprintf("Destination: %v, %v", err, DestinationPath))
		}
		// 此时, 目标只有三种可能
		// 1. 目标是目录
		//   * 需要判断源文件是否是目录
		//     * 如果是目录, 那么就需要判断目录是否为空, 如果不为空, 需要询问用户(如果遇到同名的情况如何处理: 1, 忽略; 2, 替换; 3, 重命名后写入)
		//     * 如果是文件, 则需要报错 (将一个文件拷贝成一个目录, 类型不一致), Linux cp的实现是直接拷贝到这个目录中, 个人觉得这种实现不严谨(因为其不是以 "/" 结尾, 就不应该放到它里边去)
		// 2. 目标是文件
		//   * 需要判断源路径是文件还是目录
		//     * 如果是文件, 那么判断文件是否存在. 存在的情况下, 需要询问用户如何处理(1, 忽略; 2, 覆盖; 3, 重命名后写入)
		//     * 如果是目录, 则需要报错(源路径是目录, 而目标是一个已存在的文件)
		// 3. 目标不存在, 根据源路径类型创建目标类型, 这里不考虑目标不存在的情况, 这种情况需要在复制逻辑中去实现
		if destStat, err = os.Stat(DestinationPath); nil != err {
			if os.IsNotExist(err) { // [3]
				return nil // 这里不处理这种错误, 这其实不算是错误, 只是文件不存在, 在复制业务场景中, 需要业务去创建这个文件或目录
			} else { // 其它错误
				return errors.New(fmt.Sprintf("Destination: %v, %v", err, DestinationPath))
			}
		}
		if destStat.IsDir() {
			if sourceStat.IsDir() {
				// 检查其是否为一个空目录
				if dir, _ := ioutil.ReadDir(DestinationPath); len(dir) > 0 {
					for {
						fmt.Printf("The target directory [%s] already exists, and there may be a file with the same name. What is the processing method? [R: Rename, O: Overwrite, I: Ignore]? [R(r)/(O)o/(I)i]", DestinationPath)
						var _type byte
						fmt.Scanf("%c", &_type)
						if 'R' == _type || 'r' == _type {
							DuplicateControl = RENAME_FILE
							break
						} else if 'O' == _type || 'o' == _type {
							DuplicateControl = OVERWRITE_FILE
							break
						} else if 'I' == _type || 'i' == _type {
							DuplicateControl = IGNORE_FILE
							break
						}
						// 循环让用户去输入
					}
				}
				// 如果目录是空的. 就不做任何处理
			} else {
				//_, fileName := filepath.Split(SourcePath)
				//DestinationPath = fmt.Sprintf("%s%d%s", DestinationPath, filepath.Separator, fileName)
				return errors.New(fmt.Sprintf("The directory [%s] with the same name already exists, if you want "+
					"to copy to this directory, please end with \"%d\", reference: %s%d",
					DestinationPath, filepath.Separator, DestinationPath, filepath.Separator))
			}
		} else {
			if sourceStat.IsDir() {
				return errors.New(fmt.Sprintf("Destination: %v already exists, Cannot copy directory to file", DestinationPath))
			} else {
				// 程序运行到这里, 说明目标文件一定存在. 不存在的情况在前边已经处理过了. 会直接return
				for {
					fmt.Printf("The target file [%s] already exists, What is the processing method? [R: Rename, O: Overwrite, I: Ignore]? [R(r)/(O)o/(I)i]", DestinationPath)
					var _type byte
					fmt.Scanf("%c", &_type)
					if 'R' == _type || 'r' == _type {
						DuplicateControl = RENAME_FILE
						break
					} else if 'O' == _type || 'o' == _type {
						DuplicateControl = OVERWRITE_FILE
						break
					} else if 'I' == _type || 'i' == _type {
						DuplicateControl = IGNORE_FILE
						break
					}
					// 循环让用户去输入
				}
			}
		}
		fmt.Printf("Copy:  %v -> %v", SourcePath, DestinationPath)
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cpm.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().StringP("source", "s", ".", "The source for pro")
	rootCmd.Flags().BoolP("concurrency", "c", true, "Enable concurrent copying")
}

func copyFile(source, destination string) {
	var (
		err  error
		stat os.FileInfo
	)
	if stat, err = os.Stat(source); nil != err {
		fmt.Errorf("Source file error, %v", err)
		return
	}
	// 判断 源文件是目录还是文件
	if stat.IsDir() { // 是目录

	} else { // 是文件
		// 判断目标是文件还是目录
		if stat, err = os.Stat(destination); nil != err {
			if os.IsNotExist(err) { // 如果表示不存在

			}
		}
	}
}

func doCopy(source, destination string, offset, length uint64) {

}
