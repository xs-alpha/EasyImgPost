package rely

var (
	WaterMark = []string{"吾爱破解", "lovepj", "52pj"}

	ImgUrl = "https://xxxxxxxxxxxxxxxxxxx/api/upload/img"
	// 设置代理，方便抓包分析
	ProxyUrl = ""

	WaterMarkTextMode        = 1
	WaterMarkTextRight       = 16  // 右下角水印大小
	TextDistanceToRight      = 160 // 右下角水印距离右边距离
	TextTitleDistanceToRight = 160 // 右下角水印title距离右边距离
	WaterMarkTextMiddleSize  = 30  // WaterMarkTextMiddleSize==0 代表自动设置图片中央防盗水印文字大小

	// 设置自己图床cookie
	CKNameKey    = "user"
	CKTokenKey   = "token"
	CKNameValue  = ""
	CKTokenValue = ""
)
