# atts.agi
AGI app for Asterisk TTS , Text to Speech..

## build
$ go build --ldflags "-X X:=main.version=$(PKG_VERSION) main.baiduSpeechAPIKey=xxxx main.baiduSpeechSecretKey=xxxx"
$ cp atts.agi /usr/share/asterisk/agi-bin/
## usage
```
exten => 114,1,agi(atts.agi, "我们都在期待一个灿烂的夏天")
```
