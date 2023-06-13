### [简体中文](#简体中文) | [English](#English)

# 简体中文
## [go-scout是一个侦察文件、目录发生变化的侦察服务](#)
![golang](https://img.shields.io/badge/golang-v1.19-blue)
![simple](https://img.shields.io/badge/simple-extend-green)
![development](https://img.shields.io/badge/development-master-yellowgreen)
![serve](https://img.shields.io/badge/serve-v0.2-red)

[作者邮箱1261930106@qq.com]()
### [核心原理:检测文件的变更时间，推导是否新建、删除、更新文件，基于此项目可以轻松的制作一些对文件的监控程序。]()
### 特性
<<<<<<< HEAD
* 简单已用
=======
* 简单易用
* 快速 基于 `go语言` 性能无需多言
>>>>>>> d5b6441e32e3de228e1ce797bdd10a30218a68c6
* 友好的二次开发
* 服务版和开发版

### 版本
* [开发版本 位于master分支](#)
* [服务版本 位于v0.2分支](#)

### 开发版用法

```go
    // 创建侦察对象
	scout,Paths,err := New(1000,"./")
	if err != nil {
		log.Fatalln(err)
	}
	// 返回一个侦察对象，所有管理的文件、目录路径和一个错误
	fmt.Println("Paths：",Paths)
    // 开启了调试模式 入参可选 为空时对历史值进行取反 声明对象是定义的是关闭所以这里就是开启 
	scout.SetDebug()
	// 开启侦察 入参是一个回调函数 如果发生变化会执行回调函数
	// ScoutChange 对象包含变化的路径和变化的类型（增删改）
	err = scout.Scout(func(changePath *[]ScoutChange) {
		for _, change := range *changePath {
			fmt.Printf("type:%v path:%v \n",change.Type,change.Path)
		}
	})

	fmt.Println(err)
```

### 服务版用法
[服务版是一个将发生变化的文件、目录详细信息通过配置的Http接口 发送Post请求到指定的接口]()

# English

[author email 1261930106@qq.com]()

### Character
* [Simple and easy to use]()
* [High performance based on go language performance need not be said]()
* [Friendly secondary development]()

### version
* [Development version in master branch](#)
* [serve version in v0.2 branch](#)

### Development version usage

```go
    // Create a scout object
	scout,Paths,err := New(1000,"./")
	if err != nil {
		log.Fatalln(err)
	}
	// Returns a scout object, all managed files, directory paths, and an error
	fmt.Println("Paths：",Paths)
    // Debug mode is turned on and the entry parameter is optional and the history value is taken when the undeclared object is defined to be off so this is on
	scout.SetDebug()
	// The open recon parameter is a callback function that executes if something changes
	// The ScoutChange object contains the change path and the change type (add, delete, change).
	err = scout.Scout(func(changePath *[]ScoutChange) {
		for _, change := range *changePath {
			fmt.Printf("type:%v path:%v \n",change.Type,change.Path)
		}
	})

	fmt.Println(err)
```

### Serve version usage
[The service version is a Post request that sends the changed file and directory details through the configured Http interface to the specified interface]()