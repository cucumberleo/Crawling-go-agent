### Problem1：go mod tidy (Openai库) error
### solution：添加中国代理环境路径
####  `$env:GOPROXY = "https://goproxy.cn,direct"`
--------------
### Problem2: GPT回复过慢 or timeout
### solution: 
    1. 首先检查OPENAI_API_KEY是否通过终端设置好了？ $env:OPENAI_API_KEY="your_apikey"
    2. Stream initialization error说明是梯子没挂？ 检查方法:ping api.openai.com
    3. 如果429 too many requests 说明GPT API没有额度需要充值
