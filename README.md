GoGitlabNotifer
===============
gitlab notifer terminal tool golang

## Settings ##

`create config.json`

```
$ mkdir ~/.ggn
$ cd ~/.ggn
$ touch config.json
```

`edit config.json`

```
{
  "host":     "https://git.hogehoge.com",
  "api_path": "/api/v3",
  "token":    "aaaaaaaaaaaaaaaaaaaaaa"
}
```

### Download Gilab Icon

```
$ cd ~/.ggn
$ curl -O https://github.com/kyokomi/go-gitlab-notifer/raw/master/build/logo.png
```


### Install terminal-notifer

```
$ brew install terminal-notifier
```

## Usage ##

#### Mac ユーザーの場合

**don't supported Windows.**

[gogitlabnotifer - バイナリDL](https://github.com/kyokomi/go-gitlab-notifer/blob/master/build/go-gitlab-notifer_darwin_amd64)

```
$ go-gitlab-notifer help
```

## Refrence ##

- [go-gitlb-client](https://gowalker.org/github.com/plouc/go-gitlab-client)
- [terminal color](https://github.com/wsxiaoys/terminal)

