package image

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/draw"
	"sync"

	"regexp"

	"github.com/skirrund/gcloud/imaging"
	"github.com/skirrund/gcloud/logger"
)

type circle struct { // 这里需要自己实现一个圆形遮罩，实现接口里的三个方法
	p image.Point // 圆心位置
	r int
}

func (c *circle) ColorModel() color.Model {
	return color.AlphaModel
}

func (c *circle) Bounds() image.Rectangle {
	return image.Rect(c.p.X-c.r, c.p.Y-c.r, c.p.X+c.r, c.p.Y+c.r)
}

// 对每个像素点进行色值设置，在半径以内的图案设成完全不透明
func (c *circle) At(x, y int) color.Color {
	xx, yy, rr := float64(x-c.p.X)+0.5, float64(y-c.p.Y)+0.5, float64(c.r)
	if xx*xx+yy*yy < rr*rr {
		return color.Alpha{A: 255}
	}
	return color.Alpha{}
}

type radius struct {
	p image.Point // 矩形右下角位置
	r int
}

func (c *radius) ColorModel() color.Model {
	return color.AlphaModel
}

func (c *radius) Bounds() image.Rectangle {
	return image.Rect(0, 0, c.p.X, c.p.Y)
}

// 对每个像素点进行色值设置，分别处理矩形的四个角，在四个角的内切圆的外侧，色值设置为全透明，其他区域不透明
func (c *radius) At(x, y int) color.Color {
	var xx, yy, rr float64
	var inArea bool
	// left up
	if x <= c.r && y <= c.r {
		xx, yy, rr = float64(c.r-x)+0.5, float64(y-c.r)+0.5, float64(c.r)
		inArea = true
	}
	// right up
	if x >= (c.p.X-c.r) && y <= c.r {
		xx, yy, rr = float64(x-(c.p.X-c.r))+0.5, float64(y-c.r)+0.5, float64(c.r)
		inArea = true
	}
	// left bottom
	if x <= c.r && y >= (c.p.Y-c.r) {
		xx, yy, rr = float64(c.r-x)+0.5, float64(y-(c.p.Y-c.r))+0.5, float64(c.r)
		inArea = true
	}
	// right bottom
	if x >= (c.p.X-c.r) && y >= (c.p.Y-c.r) {
		xx, yy, rr = float64(x-(c.p.X-c.r))+0.5, float64(y-(c.p.Y-c.r))+0.5, float64(c.r)
		inArea = true
	}

	if inArea && xx*xx+yy*yy >= rr*rr {
		return color.Alpha{}
	}
	return color.Alpha{A: 255}
}

var bufferPool = &sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
}

func getByteBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

func releaseByteBuffer(buffer *bytes.Buffer) {
	buffer.Reset()
	bufferPool.Put(buffer)
}

func Radius(img image.Image, r int) image.Image {
	x := img.Bounds().Dx()
	y := img.Bounds().Dy()
	c := radius{p: image.Point{X: x, Y: y}, r: r}
	radiusImg := image.NewRGBA(image.Rect(0, 0, x, y))
	draw.DrawMask(radiusImg, radiusImg.Bounds(), img, image.Point{}, &c, image.Point{}, draw.Over)
	return radiusImg
}

func Circle(img image.Image) image.Image {
	x := img.Bounds().Dx()
	y := img.Bounds().Dy()
	x1 := x / 2
	if y < x {
		x1 = y / 2
	}
	c := circle{p: image.Point{X: x1, Y: x1}, r: x1}
	circleImg := image.NewRGBA(image.Rect(0, 0, x, x))
	draw.DrawMask(circleImg, circleImg.Bounds(), img, image.Point{}, &c, image.Point{}, draw.Over)
	return circleImg
}

// base64压缩 limit:KB
// 最长边长度 max
func CommpressBase64Pic(base64Str string, limit int, max int) string {
	bs := CommpressBase64PicToByte(base64Str, limit, max)
	return base64.RawStdEncoding.EncodeToString(bs)
}

// base64压缩 limit:KB
// 最长边长度 max
func CommpressBase64PicToByte(base64Str string, limit int, max int) []byte {
	reg, _ := regexp.Compile("[\\s*\t\n\r]")
	base64Str = reg.ReplaceAllString(base64Str, "")
	bs, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		logger.Error("base64 commpress error:", err.Error())
		return bs
	}
	if max > 0 {
		bs, err = doResize(bs, max)
		if err != nil {
			return bs
		}
	}
	return doCompressQuality(bs, limit)
}

func doResizeWH(b []byte, scale float64) ([]byte, error) {
	img, format, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		logger.Error("doResizeWH error:", err.Error())
		return b, err
	}

	h := float64(img.Bounds().Dy()) * scale
	w := float64(img.Bounds().Dx()) * scale

	img = imaging.Fit(img, int(w), int(h), imaging.Lanczos)
	buf := getByteBuffer()
	defer releaseByteBuffer(buf)
	err = imaging.Encode(buf, img, imaging.JPEG)
	//err = jpeg.Encode(buf, img, nil)
	if err != nil {
		logger.Error("doResizeWH error:", err.Error(), ",format:"+format)
		return b, err
	}
	return buf.Bytes(), nil
}

func doResize(b []byte, max int) ([]byte, error) {
	img, format, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		logger.Error("doResize error:", err.Error())
		return b, err
	}

	h := img.Bounds().Dy()
	w := img.Bounds().Dx()
	if h <= max && w <= max {
		return b, nil
	}
	logger.Info("doResize orgin,w:", w, ",h:", h)
	maxSide := h
	if h < w {
		maxSide = w
	}
	scale := 1.0
	if maxSide > max {
		scale = float64(max) / float64(maxSide)
		h = int(float64(h) * scale)
		w = int(float64(w) * scale)

	}
	img = imaging.Fit(img, w, h, imaging.Lanczos)
	buf := getByteBuffer()
	defer releaseByteBuffer(buf)
	err = imaging.Encode(buf, img, imaging.JPEG)
	// if strings.EqualFold("png", format) {
	// 	img = convertToJpeg(img)
	// }
	// err = jpeg.Encode(buf, img, nil)
	if err != nil {
		logger.Error("doResize error:", err.Error(), ",format:"+format)
		return b, err
	}
	return buf.Bytes(), nil
}

func convertToJpeg(pngImg image.Image) image.Image {
	newImg := image.NewRGBA(pngImg.Bounds())
	draw.Draw(newImg, newImg.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	draw.Draw(newImg, newImg.Bounds(), pngImg, pngImg.Bounds().Min, draw.Over)
	return newImg
}

func doCompressQuality(b []byte, limit int) []byte {
	size := len(b) / 1024
	if size > limit {
		logger.Info("commpressImage current:", size, ",limit:", limit)
		bs, err := doResizeWH(b, 0.7)
		if err != nil {
			return bs
		}
		return doCompressQuality(bs, limit)

		// img, format, err := image.Decode(bytes.NewReader(b))
		// if err != nil {
		// 	logger.Logger.Error("commpressImage error:", err.Error())
		// 	return b
		// }
		// logger.Logger.Info("commpressImage current:", size, ",limit:", limit, ",format:", format, ",accuracy:", accuracy)
		// buf := &bytes.Buffer{}
		// err = jpeg.Encode(buf, img, &jpeg.Options{Quality: accuracy})
		// if err != nil {
		// 	logger.Logger.Error("commpressImage error:", err.Error())
		// 	return b
		// }
		// size = buf.Len() / 1024
		// logger.Logger.Info("commpressImage after size :", size, ",limit:", limit)
		// if size <= limit {
		// 	return buf.Bytes()
		// } else {
		// 	if accuracy == 0 {
		// 		return doCompressQuality(doResizeWH(buf.Bytes(), 0.8), limit, nil)
		// 	}
		// 	return doCompressQuality(buf.Bytes(), limit, int(float64(accuracy)*0.44))
		// }
	} else {
		return b
	}
}
