# <··标题··>

## 题目描述
给定一个长度为 $n$ 的数组 $\{a_1,a_2,\dots,a_n\}$，初始这个数组只含 $\{0,1,2\}$ 这三个数字，并且初始有且仅有一个元素为数字 $2$。Alice 与 Bob 决定在这个数组上玩个游戏，双方轮流操作数组：

$\hspace{23pt}\bullet$当 Alice 操作时，她必须选择一个区间 $[l,r]$ 满足 $1\leqq l<r\leqq |a|$，记录 $x=a_l+a_{l+1}+\dots+a_r$，将 $a_l$ 的值修改为 $x$ 并删除 $a_{l+1},a_{l+2},\dots,a_r$，剩余元素按照原顺序拼接。

$\hspace{23pt}\bullet$当 Bob 操作时，他必须选择一个区间 $[l,r]$ 满足 $1\leqq l<r\leqq |a|$，记录 $x=a_l\times a_{l+1}\times\dots\times a_r$，将 $a_l$ 的值修改为 $x$ 并删除 $a_{l+1},a_{l+2},\dots,a_r$，剩余元素按照原顺序拼接。

$\hspace{15pt}$当数组长度只剩 $1$ 时游戏结束，Alice 的目标是使这个元素尽可能大，Bob 的目标是尽可能小。

$\hspace{15pt}$请输出最终这个元素的值。

## 输入描述
$\hspace{15pt}$每个测试文件均包含多组测试数据。第一行输入一个整数 $T\left(1\leqq T\leqq 10^4\right)$ 代表数据组数，每组测试数据描述如下：

$\hspace{23pt}\bullet$第一行输入一个整数 $n(1\leqq n\leqq 2 \times 10^5)$ 表示数组长度。

$\hspace{23pt}\bullet$第二行输入 $n$ 个整数 $a_1,a_2,\dots,a_n(0\leqq a_i\leqq 2)$ 表示数组 $a$，同时数据保证数组 $a$ 中数字 $2$ 出现且仅出现一次。

$\hspace{15pt}$除此之外，保证单个测试文件的 $n$ 之和不超过 $2 \times 10^5$。

## 输出描述
$\hspace{15pt}$对于每一组测试数据，新起一行，输出一个整数表示结果。

## 样例

```text input:#1
<··样例输入··>
```
```text output:#1
<··样例输出··>
```