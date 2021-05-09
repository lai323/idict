### 用于命令行的英语词典

#### 功能
- 英文到中文的翻译查询
- 类似多邻国的单词记忆练习

#### 演示
![translate](./img/translate.gif)
![practice 属性文本](./img/practice.gif)

#### 配置

配置文件默认位置：`~/.config/idict/idict.yaml`

- `StoragePath`: 存储位置，默认：`~/.local/share/idict`
- `GroupNum`: 一组练习的单词数量，这一组单词不断循环出现，直到拼写正确 默认：`20`
- `RestudyInterval`: 一个单词的连续正确次数，与复习时间间隔，以小时为单位

    默认为

        ```
            RestudyInterval:
              3:0
              5:12
              8:36
              12:72
              17:120
              23:240
              26:-1
        ```

    这代表:

        3  >= 连续正确输入次数    会立即重复
        5  >= 连续正确输入次数    会在 12 小时后重复
        8  >= 连续正确输入次数    会在 36 小时后重复
        12 >= 连续正确输入次数    会在 72 小时后重复
        17 >= 连续正确输入次数    会在 120 小时后重复
        23 >= 连续正确输入次数    会在 240 小时后重复
        26 <= 连续正确输入次数    不再重复

- `FfplayPath`: 设置 ffplay 可以启用单词发音
- `FfplayArgs`: ffplay 的参数
