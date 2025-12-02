package main

import "jt808/pkg"

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {

	//如果有多个服务，用GO异步处理
	go pkg.StartJT808()
	//启动HTTP服务，TODO传JT808Server的SessionManager接口给HTTPServer实现下行数据发送

	//阻塞
	select {}
}
