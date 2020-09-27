# goToMap
Google マップで宝探しするゲーム (Go言語製サーバー)

### 現実の街探検を仮想で楽しむゲームです。

![goToMap4.gif](https://github.com/yasutakatou/goToMap/blob/pic/goToMap.gif)

 - Googleマップの緯度・経度を使って各プレイヤーのスタート地点とゴールを決めます
 - ゴールまでにヒントのテキストや、画像を仕掛けたりできます
 - プレイヤー同士でメッセージをやりとりしたり、書置きを読んだりできます

### ・Windows、Linux、Rasbian対応のGolang製Websocketサーバー
<a href="https://github.com/yasutakatou/goToMap"><img src="https://github-link-card.s3.ap-northeast-1.amazonaws.com/yasutakatou/goToMap.png" width="460px"></a>

サーバーとするPCで動かしてください。

### ・Chrome拡張 (Chromiumも対応)
<a href="https://github.com/yasutakatou/goToMapExt"><img src="https://github-link-card.s3.ap-northeast-1.amazonaws.com/yasutakatou/goToMapExt.png" width="460px"></a>

ゲームする端末上のChrome(と互換ブラウザ)で動かしてください。サーバーと同居できます。

# じゅんび:

Googleマップで遊びたい街を開いてください。

![url.png](https://github.com/yasutakatou/goToMap/blob/pic/url.png)

こんなかんじでURL出ますんで**下線の緯度経度**をピックアップしてコンフィグを作ります。
コンフィグの例は以下です。

```
[PLAYER]
https://www.google.co.jp/maps/@35.6608375,139.7008749,3a,75y,43.91h,92.5t/data=!3m6!1e1!3m4!1sDrFpHa0VreQbapnpTC-QgA!2e0!7i16384!8i8192
https://www.google.co.jp/maps/@35.6613982,139.7001959,3a,75y,131.45h,104.53t/data=!3m6!1e1!3m4!1szZPFU_fUtAcf-e3DGysGlw!2e0!7i16384!8i8192

[GOAL]
35.6618983,139.7008538

[RESULT]
ゴールに着きました！

[ACTION]
35.6602069,139.7007403	マックに着きました～
35.6632348,139.701151	https://www.google.co.jp/images/branding/googlelogo/1x/googlelogo_color_272x92dp.png
```

 - [PLAYER]はスタート地点です。URLをベタ貼りしてください
 - [GOAL]は目的地、ゴールの緯度経度を書いてください
 - [RESULT]はゴール時のメッセージです。httpでURL指定で画像も指定できます
 - [ACTION]は緯度経度＋タブで、緯度経度に来た時にメッセージや画像を表示します

緯度経度を調べながらヒント作ったり、画像貼ったりでゲームの"面"を作ってください

# きどうほうほう：

　**Go言語のWebsocketサーバーを起動**します。**同フォルダにコンフィグ置いて実行**すればOK。
　次にChrome拡張インストールします。二人以上のプレイヤーの場合は**拡張のアイコン**を開いて**「接続サーバー」欄にサーバーのIP:Portを指定**してください。保存して拡張を**再起動**します。

![1.png](https://github.com/yasutakatou/goToMap/blob/pic/1.png)

**join client!と出たら接続成功**です。**スタート地点のタブ**が自動的に開きます。

![2.png](https://github.com/yasutakatou/goToMap/blob/pic/2.png)

# うごかしかた:

基本**Googleマップで目的地**に着けば良いです。その他は装飾になります。

![3.png](https://github.com/yasutakatou/goToMap/blob/pic/3.png)

コンフィグで指定した**任意の場所でメッセージ**を出せます

![4.png](https://github.com/yasutakatou/goToMap/blob/pic/4.png)

任意の場所で**画像**も出せます

![5.png](https://github.com/yasutakatou/goToMap/blob/pic/51.png)

SNSぽく、**参加者全員にメッセージ**を表示できます。

![6.png](https://github.com/yasutakatou/goToMap/blob/pic/6.png)

他プレイヤーにお知らせが出ました

![7.png](https://github.com/yasutakatou/goToMap/blob/pic/7.png)

**特定の場所に書置き**ができます。メッセージ通知を後から追加できる機能です

![8.png](https://github.com/yasutakatou/goToMap/blob/pic/8.png)

# LICENSE

BSD-2-Clause License

# Contributors

- [yasutakatou](https://github.com/yasutakatou)
