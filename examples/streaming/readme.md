基本用法
在代码中使用 SentenceBoundary 功能，您需要在创建 Communicate 实例时传递额外的参数。以下是一个简单的例子：

package main

import (
    "context"
    "fmt"
    "os"

    "github.com/difyz9/edge-tts-go/pkg/communicate"
)

func main() {
    // 创建上下文
    ctx := context.Background()

    // 要转换为语音的文本
    text := "你好，世界！这是一个使用 SentenceBoundary 功能的示例。"

    // 使用的语音
    voice := "zh-CN-XiaoxiaoNeural"

    // 创建新的 Communicate 实例，注意最后一个参数 "SentenceBoundary"
    comm, err := communicate.NewCommunicate(
        text,
        voice,
        "+0%",  // 语速
        "+0%",  // 音量
        "+0Hz", // 音调
        "",     // 代理
        10,     // 连接超时
        60,     // 接收超时
        "SentenceBoundary", // 边界类型
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "创建 Communicate 实例时出错: %v\n", err)
        os.Exit(1)
    }

    // 将音频保存到文件
    err = comm.Save(ctx, "output.mp3", "")
    if err != nil {
        fmt.Fprintf(os.Stderr, "保存音频时出错: %v\n", err)
        os.Exit(1)
    }

    fmt.Println("音频已保存到 output.mp3")
}
带字幕的流式处理
如果您想使用 SentenceBoundary 功能并生成字幕，可以参考以下示例：

package main

import (
    "context"
    "fmt"
    "os"

    "github.com/difyz9/edge-tts-go/pkg/communicate"
    "github.com/difyz9/edge-tts-go/pkg/submaker"
)

func main() {
    // 创建上下文
    ctx := context.Background()

    // 要转换为语音的文本
    text := "你好，世界！这是一个使用 SentenceBoundary 功能和字幕的示例。"

    // 使用的语音
    voice := "zh-CN-XiaoxiaoNeural"

    // 创建新的 Communicate 实例，使用 SentenceBoundary
    comm, err := communicate.NewCommunicate(
        text,
        voice,
        "+0%",  // 语速
        "+0%",  // 音量
        "+0Hz", // 音调
        "",     // 代理
        10,     // 连接超时
        60,     // 接收超时
        "SentenceBoundary", // 边界类型
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "创建 Communicate 实例时出错: %v\n", err)
        os.Exit(1)
    }

    // 创建 SubMaker 实例
    sm := submaker.NewSubMaker()

    // 打开输出文件
    audioFile, err := os.Create("output.mp3")
    if err != nil {
        fmt.Fprintf(os.Stderr, "创建音频文件时出错: %v\n", err)
        os.Exit(1)
    }
    defer audioFile.Close()

    subFile, err := os.Create("output.srt")
    if err != nil {
        fmt.Fprintf(os.Stderr, "创建字幕文件时出错: %v\n", err)
        os.Exit(1)
    }
    defer subFile.Close()

    // 流式处理音频和元数据
    chunkChan, errChan := comm.Stream(ctx)

    // 处理数据块
    for chunk := range chunkChan {
        if chunk.Type == "audio" {
            _, err := audioFile.Write(chunk.Data)
            if err != nil {
                fmt.Fprintf(os.Stderr, "写入音频数据时出错: %v\n", err)
                os.Exit(1)
            }
        } else if chunk.Type == "WordBoundary" || chunk.Type == "SentenceBoundary" {
            // 注意这里处理了两种边界类型
            err := sm.Feed(chunk)
            if err != nil {
                fmt.Fprintf(os.Stderr, "处理 %s 时出错: %v\n", chunk.Type, err)
                os.Exit(1)
            }
        }
    }

    // 检查错误
    if err := <-errChan; err != nil {
        fmt.Fprintf(os.Stderr, "流式处理时出错: %v\n", err)
        os.Exit(1)
    }

    // 合并字幕提示以减少数量
    err = sm.MergeCues(10) // 每个提示 10 个单词
    if err != nil {
        fmt.Fprintf(os.Stderr, "合并字幕提示时出错: %v\n", err)
        os.Exit(1)
    }

    // 将字幕写入文件
    _, err = fmt.Fprint(subFile, sm.GetSRT())
    if err != nil {
        fmt.Fprintf(os.Stderr, "写入字幕时出错: %v\n", err)
        os.Exit(1)
    }

    fmt.Println("音频已保存到 output.mp3")
    fmt.Println("字幕已保存到 output.srt")
}
SentenceBoundary 与 WordBoundary 的区别
WordBoundary：为每个单词返回一个边界，适合大多数西方语言。
SentenceBoundary：为每个句子返回一个边界，特别适合中文等亚洲语言，可以提供更自然的句子分隔。
注意事项
在处理数据块时，需要同时检查 chunk.Type == "WordBoundary" || chunk.Type == "SentenceBoundary"，以确保能够处理所有边界事件。
对于中文文本，建议使用 SentenceBoundary 以获得更自然的句子分隔。
如果您只关心音频输出而不需要字幕，可以使用简单的例子，只需添加 SentenceBoundary 参数即可。
