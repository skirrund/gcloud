package image

import (
	"image"
	"image/draw"
	"os"
	"testing"

	"github.com/skirrund/gcloud/imaging"
)

const (
	circleHeight       = 194
	circleWidth        = 191
	headerY            = 144
	headerX            = 600
	hY                 = headerY + circleHeight
	hX                 = headerX + circleWidth
	headerResizeWidth  = circleWidth - 40
	headerResizeHeight = circleHeight - 40
)

func TestCommpressBase64Pic(t *testing.T) {
	af, _ := os.OpenFile("/Users/jerry.shi/Desktop/头像_1.jpeg", os.O_RDONLY, os.ModePerm)
	aImage, _, _ := image.Decode(af)
	radiusImg := Radius(aImage, aImage.Bounds().Dy()/2)
	touxiang := imaging.Resize(radiusImg, headerResizeWidth, headerResizeHeight, imaging.Lanczos)

	f1, _ := os.OpenFile("/Users/jerry.shi/Desktop/ann_2_1.jpeg", os.O_RDONLY, os.ModePerm)
	defer f1.Close()
	bg, _, _ := image.Decode(f1)
	f3, _ := os.OpenFile("/Users/jerry.shi/Desktop/多少天海报_03.png", os.O_RDONLY, os.ModePerm)
	defer f3.Close()
	//img3, _, _ := image.Decode(f3)
	base := image.NewNRGBA(bg.Bounds())
	draw.Draw(base, bg.Bounds(), bg, image.Pt(0, 0), draw.Over)
	draw.Draw(base, image.Rect(headerX, headerY, hX, hY), touxiang, image.Pt(-20, -20), draw.Over)
	//draw.Draw(base, image.Rect(headerX, headerY, hX, hY), img3, image.Pt(0, 0), draw.Over)
	imaging.Save(base, "/Users/jerry.shi/Desktop/企微头像_1.jpeg")
}
