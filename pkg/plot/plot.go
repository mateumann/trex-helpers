package plot

import (
	"fmt"
	"os"
	"time"
	"trex-helpers/pkg/packet"

	"github.com/signintech/gopdf"
)

func SavePDF(packets []packet.Packet, inputFilename string, filename string, verbose bool) error {
	//xScale, yScale := 0.0000002, 1.0
	xMin, xMax, yMin, yMax := maxPacketsValue(packets)
	width, height := xMax-xMin, yMax-yMin
	//xOffset := box.Min.X
	//yOffset := box.Min.Y
	if verbose {
		fmt.Printf("PDF for data in range %v × %v\n", width, height)
	}
	pdf, err := preparePdf(inputFilename)
	if err != nil {
		return err
	}

	//img := image.NewRGBA(image.Rect(0, 0, width+100, height+100))
	//
	//for _, pkt := range packets {
	//	x := int(float64(pkt.ReceivedAt().UnixNano()) * xScale)
	//	if pkt.Value() >= 0 {
	//		for y := 0; y < int(pkt.Value()*yScale); y++ {
	//			img.Set(x-xOffset+50, height-(y-yOffset+50), pktTypeToColor(pkt.Type()))
	//		}
	//	} else {
	//		for y := 0; y > int(pkt.Value()*yScale); y-- {
	//			img.Set(x-xOffset+50, height-(y-yOffset+50), pktTypeToColor(pkt.Type()))
	//		}
	//	}
	//
	//	img.Set(x-xOffset+50, height-(yOffset+50), color.RGBA{0, 0, 0, 0xff})
	//}
	//
	//addLabel(img, 20, height-(yOffset+50), "ala i kot")

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	err = pdf.Write(f)
	if err != nil {
		return err
	}

	return nil
}

func maxPacketsValue(packets []packet.Packet) (xMin int64, xMax int64, yMin float64, yMax float64) {
	xMin, xMax = int64(999999999999999999), int64(-999999999999999999)
	yMin, yMax = 999999999999999999.0, -999999999999999999.0
	for _, pkt := range packets {
		x := pkt.ReceivedAt().UnixNano()
		y := pkt.Value()
		if x < xMin {
			xMin = x
		}
		if x > xMax {
			xMax = x
		}
		if y < yMin {
			yMin = y
		}
		if y > yMax {
			yMax = y
		}
	}
	return
}

//
//func pktTypeToColor(t packet.Type) color.RGBA {
//	switch t {
//	case packet.TypeLatency:
//		return color.RGBA{
//			R: 0x33,
//			G: 0xff,
//			B: 0,
//			A: 0xff,
//		}
//	case packet.TypePTP:
//		return color.RGBA{
//			R: 0xff,
//			G: 0x33,
//			B: 0,
//			A: 0xff,
//		}
//	case packet.TypeOther:
//		return color.RGBA{
//			R: 0x80,
//			G: 0x80,
//			B: 0x80,
//			A: 0x40,
//		}
//	default:
//		return color.RGBA{
//			R: 0,
//			G: 0,
//			B: 0,
//			A: 0,
//		}
//	}
//}
//
//func addLabel(aImg *image.RGBA, x, y int, label string) {
//	fmt.Printf("Will add label %v at (%v,%v)\n", label, x, y)
//	point := fixed.Point26_6{X: fixed.Int26_6(x), Y: fixed.Int26_6(y)}
//
//	d := &font.Drawer{
//		Dst:  aImg,
//		Src:  image.NewUniform(color.RGBA{R: 0xff, G: 0x66, B: 0x99, A: 0xff}),
//		Face: inconsolata.Regular8x16,
//		Dot:  point,
//	}
//	d.DrawString(label)
//}

func preparePdf(inputFilename string) (pdf gopdf.GoPdf, err error) {
	pdf = gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: 842, H: 595}, Unit: gopdf.Unit_PT})
	pdf.AddPage()
	err = pdf.AddTTFFont("DejaVuSans-Regular", "ttf/DejaVuSans.ttf")
	if err != nil {
		return
	}
	err = pdf.AddTTFFont("DejaVuSans-Bold", "ttf/DejaVuSans-Bold.ttf")
	if err != nil {
		return
	}

	err = titlePdf(&pdf, inputFilename)
	if err != nil {
		return
	}
	err = footnotePdf(&pdf)
	if err != nil {
		return
	}
	return
}

func titlePdf(pdf *gopdf.GoPdf, inputFilename string) (err error) {
	pdf.SetTextColor(0x00, 0x11, 0x66)
	err = pdf.SetFont("DejaVuSans-Regular", "", 18)
	if err != nil {
		return
	}
	pdf.SetX(4)
	pdf.SetY(22)
	err = pdf.Text("TRex Packets Chart for ")
	if err != nil {
		return err
	}
	err = pdf.SetFont("DejaVuSans-Bold", "", 18)
	if err != nil {
		return
	}
	err = pdf.Text(inputFilename)
	if err != nil {
		return err
	}
	return nil
}

func footnotePdf(pdf *gopdf.GoPdf) (err error) {
	err = pdf.SetFont("DejaVuSans-Regular", "", 8)
	if err != nil {
		return
	}
	pdf.SetTextColor(0x00, 0x11, 0x66)

	pdf.SetX(4)
	pdf.SetY(592)
	err = pdf.Text(fmt.Sprintf("%v", time.Now()))
	if err != nil {
		return err
	}
	pdf.SetX(730)
	err = pdf.Text("generated with trex-helper")
	if err != nil {
		return err
	}
	pdf.AddExternalLink("https://github.com/mateumann/trex-helpers", 729.5, 584.5, 111, 10)
	return nil
}
