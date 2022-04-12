package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"

	"strings"

	"os"
	"strconv"
	"time"

	"guiSearch/rely"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/atotto/clipboard"
	"github.com/golang/freetype"
	"github.com/sirupsen/logrus"
)

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func main() {
	a := app.New()
	w := a.NewWindow("EasyPost")
	w.FixedSize()
	w.Resize(fyne.NewSize(450, 400))
	// b := btnDemo(w)
	// e := entryDemo(w, "")
	l := ListenOnClipBoard(w)
	w.SetContent(container.NewVBox(l))
	w.ShowAndRun()
}

func showImg(fileName string) *fyne.Container {
	// image := canvas.NewImageFromResource(theme.FyneLogo())
	// image := canvas.NewImageFromURI(uri)
	// image := canvas.NewImageFromImage(src)
	// image := canvas.NewImageFromReader(reader, name)
	image := canvas.NewImageFromFile(fileName)
	image.FillMode = canvas.ImageFillOriginal
	c := container.NewWithoutLayout(image)
	return c
}

func ListenOnClipBoard(w fyne.Window) *fyne.Container {
	label1 := widget.NewEntry()
	btn1 := widget.NewButton("copy", func() {
		ct, err := rely.ReadFromClipBoard()
		if err != nil {
			fmt.Println(err)
			label1.SetText(err.Error())
			return
		}
		text := saveImg(&ct, w)
		label1.SetText(text)
	})
	c := container.NewVBox(label1, btn1)
	return c
}

func saveImg(data *io.Reader, w fyne.Window) string {
	// fmt.Println(*data)
	ret, url := writeInFile(data, w)
	if ret {
		fmt.Println("save success")
	}
	return url
}

func writeInFile(data *io.Reader, c fyne.Window) (bool, string) {
	fmt.Printf("%T\n", *data)
	timestamps := time.Now().UnixNano()
	filename := strconv.FormatInt(timestamps, 10) + ".png"
	dst, err4 := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
	if false {
		judge := *data
		// judje: 0xc000562400 data: 0xc00030b460
		// fmt.Println("judje:", &judge, "data:", data)
		// judge: 0xc00038def0, data: 0xc00038def0
		// fmt.Printf("judge: %p, data: %p\n", judge, *data)
		buf := make([]byte, 100)
		n, _ := judge.Read(buf)
		fmt.Println("read:", n)
		if n == 0 {
			return false, ""
		}
	}
	if err4 != nil {
		fmt.Println(err4)
		return false, ""
	}
	io.Copy(dst, *data)
	dst.Close()

	ret := judgeFileSize(filename)
	if ret {
		err := os.Remove(filename)
		if err != nil {
			fmt.Println(err)
		}
		return true, ""
	}
	addWaterMarkHandler(filename)
	url := PostImg(filename, c)
	os.Remove(filename)
	return true, url
}

func judgeFileSize(filename string) bool {
	// 判断文件大小，如果大小==0k,则返回true,删除文件
	fileInfo, _ := os.Open(filename)

	fi, _ := fileInfo.Stat()
	fmt.Println("size", fi.Size())
	if fi.Size() == 0 {
		fileInfo.Close()
		return true
	}
	fileInfo.Close()
	return false
}

func addWaterMarkHandler(filename string) {
	// 选择图片水印还是文字水印
	if rely.WaterMarkTextMode == 1 {
		addworkMarkText(filename)
	} else {
		addWaterMarkPng(filename)
	}
}
func addWaterMarkPng(filename string) {
	// 图片水印
	img_file, err := os.Open(filename)
	defer img_file.Close()
	if handleErr(err) {
		fmt.Println("err1", err.Error())
		return
	}
	img, _, err2 := image.Decode(img_file)
	if handleErr(err2) {
		fmt.Println("err1", err.Error())
		return
	}

	wmb_file, err3 := os.Open("new.png")
	defer wmb_file.Close()
	if handleErr(err3) {
		fmt.Println("err1", err.Error())
		return
	}

	wmb_img, _, err4 := image.Decode(wmb_file)
	if handleErr(err4) {
		fmt.Println("err1", err.Error())
		return
	}
	// fmt.Println("img:", 333)
	//把水印写在右下角，并向0坐标偏移10个像素
	offset := image.Pt(img.Bounds().Dx()-wmb_img.Bounds().Dx()-10, img.Bounds().Dy()-wmb_img.Bounds().Dy()-10)
	b := img.Bounds()
	//根据b画布的大小新建一个新图像
	m := image.NewRGBA(b)
	draw.Draw(m, b, img, image.ZP, draw.Src)
	draw.Draw(m, wmb_img.Bounds().Add(offset), wmb_img, image.ZP, draw.Over)

	//生成新图片new.jpg,
	filename2 := "new_" + filename
	imgw, err := os.Create(filename2)
	// fmt.Println("img:", 444)
	png.Encode(imgw, m)
	defer imgw.Close()
}
func handleErr(err error) bool {
	if err != nil {
		fmt.Println(err)
		return true
	}
	return false
}
func addworkMarkText(filename string) {
	//需要加水印的图片
	imgfile, _ := os.Open(filename)
	defer imgfile.Close()

	jpgimg, _, _ := image.Decode(imgfile)

	img := image.NewNRGBA(jpgimg.Bounds())

	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			img.Set(x, y, jpgimg.At(x, y))
		}
	}
	//拷贝一个字体文件到运行目录
	fontBytes, err := ioutil.ReadFile("./res/simsun2.ttf")
	if err != nil {
		log.Println(err)
	}

	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Println(err)
	}
	for i := 1; i < 3; i++ {
		f := freetype.NewContext()
		f.SetDPI(72)
		f.SetFont(font)
		f.SetFontSize(float64(rely.WaterMarkTextRight))
		f.SetClip(jpgimg.Bounds())
		f.SetDst(img)
		if i == 1 {
			f.SetSrc(image.NewUniform(color.RGBA{R: 240, G: 240, B: 240, A: 200}))
			pt := freetype.Pt(img.Bounds().Dx()-rely.TextTitleDistanceToRight, img.Bounds().Dy()-26)
			_, err = f.DrawString(rely.WaterMark[i-1], pt)
			pt2 := freetype.Pt(img.Bounds().Dx()-rely.TextDistanceToRight, img.Bounds().Dy()-3)
			_, err = f.DrawString(rely.WaterMark[i], pt2)
		} else {
			fontSize := img.Bounds().Dx() / 4
			if rely.WaterMarkTextMiddleSize != 0 {
				fontSize = rely.WaterMarkTextMiddleSize
			}
			f.SetFontSize(float64(fontSize))
			f.SetSrc(image.NewUniform(color.RGBA{R: 232, G: 232, B: 232, A: 200}))
			pt := freetype.Pt(img.Bounds().Dx()/2, img.Bounds().Dy()/2)
			_, err = f.DrawString(rely.WaterMark[i], pt)
		}
		if err != nil {
			log.Println(err)
		}
	}
	//draw.Draw(img,jpgimg.Bounds(),jpgimg,image.ZP,draw.Over)
	//保存到新文件中
	filename2 := "./out/new_" + filename
	newfile, _ := os.Create(filename2)
	defer newfile.Close()

	err = png.Encode(newfile, img)
	if err != nil {
		fmt.Println(err)
	}

}

func PostImg(filename string, c fyne.Window) string {
	filePath := "./out/new_" + filename
	client := &http.Client{}
	// 设置代理
	// proxy_raw := "127.0.0.1:8889"
	proxy_raw := rely.ProxyUrl
	if proxy_raw != "" {
		proxy_str := fmt.Sprintf("http://%s", proxy_raw)
		fmt.Println(proxy_str)
		proxy, err := url.Parse(proxy_str)
		if err != nil {
			fmt.Println(err)
		}
		client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxy)}}
	}

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	//添加表单属性
	// 参数1 （File参数）
	f, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	// fw1, err := bodyWriter.CreateFormFile("file", filepath.Base(filePath))
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			escapeQuotes("file"), escapeQuotes("new_"+filename)))
	h.Set("Content-Type", "image/png")
	fw1, _ := bodyWriter.CreatePart(h)
	if err != nil {
		return ""
	}
	//把文件流写入到缓冲区里去
	_, err1 := io.Copy(fw1, f)
	if err1 != nil {
		return ""
	}
	err = bodyWriter.Close()
	if err != nil {
		logrus.Errorf("bodyWriter close error: %v", err.Error())
		return ""
	}
	//http://logrus.Info("请求报文:", requst.JSONString())
	//发送post请求
	uri := rely.ImgUrl
	req, err := http.NewRequest("POST", uri, bodyBuf)
	if err != nil {
		return ""
	}
	req.AddCookie(&http.Cookie{Name: rely.CKNameKey, Value: rely.CKNameValue})
	req.AddCookie(&http.Cookie{Name: rely.CKTokenKey, Value: rely.CKTokenValue})
	//添加头文件
	// req.Header.Add("Content-Type", "multipart/form-data;boundary=----WebKitFormBoundaryQniDiwoA1m7fjmVC")

	req.Header.Add("User-Agent", "")

	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
	req.Header.Add("Accept", "application/json, text/plain, */*, q=0.01")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9,en")

	req.Header.Add("sec-fetch-site", "same-origin")
	req.Header.Add("sec-fetch-mode", "cors")
	req.Header.Add("sec-fetch-dest", "empty")
	req.Header.Add("X-requested-with", "XMLHttpRequest")
	req.Header.Add("DNT", "1")
	// req.Header.Add("cookie", "")

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	var reader io.ReadCloser
	if resp.Header.Get("Content-Encoding") == "gzip" { //注意此处一定要判断，否则byte数组转换成字符串时会乱码，
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			fmt.Println(err)
			return ""
		}
	}

	data, _ := ioutil.ReadAll(reader)
	url := unmarshalJson(&data, c)
	fmt.Println(string(data))
	// filename2 := "./out/new_" + filename

	// c.SetContent(showImg(filename2))
	return url
}
func unmarshalJson(data *[]byte, c fyne.Window) string {
	var result rely.RetJson
	err := json.Unmarshal(*data, &result)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	go toClipBoard(result.URL)
	fmt.Println("---", result.URL)
	return result.URL
}

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func toClipBoard(s string) {
	str := fmt.Sprintf("![](%s)", s)
	err := clipboard.WriteAll(str)
	if err != nil {
		log.Println(err)
	}
}
