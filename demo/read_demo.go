package main

import (
	"fmt"
	"io"
	"os"
)

const (
	sourceFile string = "/Users/Bottle/Downloads/java_pid19653.hprof"
	destFile   string = "/Users/Bottle/Downloads/java_pid19653-copy.hprof"
)

func readGoRoutine(source string) uint64 {
	if stat, err := os.Stat(source); nil != err {
		if os.IsNotExist(err) {
			fmt.Printf("文件[%v]不存在, %v\n", source, err)
			return 0
		} else {
			fmt.Printf("文件[%v]不存在, %v\n", source, err)
			return 0
		}
	} else {
		if stat.Size() > (100 * 1024 * 1024) {
			return 8
		} else if stat.Size() > (50 * 1024 * 1024) {
			return 4
		} else if stat.Size() > (20 * 1024 * 1024) {
			return 2
		} else {
			return 1
		}
	}
}

func main() {
	file, err := os.OpenFile(sourceFile, os.O_RDONLY, 0755)
	if nil != err {
		fmt.Printf("打开文件[%s]失败: %v", sourceFile, err)
		return
	}
	defer file.Close()
	dest, err := os.OpenFile(destFile, os.O_WRONLY|os.O_EXCL|os.O_TRUNC, 0755)
	if nil != err {
		if os.IsNotExist(err) {
			dest, err = os.Create(destFile)
			if nil != err {
				fmt.Printf("创建文件[%s]失败, %v\n", destFile, err)
				return
			}
		} else {
			fmt.Printf("打开文件[%s]失败: %v\n", destFile, err)
			return
		}
	}
	defer dest.Close()

	buf := make([]byte, 4096)
	var offset int64 = 0
	var brk bool = false
	for {
		read, err := file.ReadAt(buf, offset)
		if nil != err {
			if err == io.EOF { // 读完退出
				fmt.Printf("操作完毕, 退出, 最后一次读到: %d字节", read)
				brk = true
			} else {
				fmt.Printf("读取文件出错: %v\n", err)
				return
			}
		}
		_, err = dest.WriteAt(buf[:read], offset)
		if nil != err {
			fmt.Printf("写文件出错: %v\n", err)
			return
		}
		offset = offset + int64(read)
		if brk {
			break
		}
	}
}
