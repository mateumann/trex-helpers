package plot

import (
	"fmt"
	"os"
	"time"
	"trex-helpers/pkg/packet"

	"github.com/signintech/gopdf"
)

const xPaperSize float64 = 842.0
const yPaperSize float64 = 595.0

func SavePDF(packets []packet.Packet, inputFilename string, filename string, verbose bool) (err error) {
	pdf, err := preparePdf(inputFilename)
	if err != nil {
		return
	}
	drawPackets(&pdf, packets)

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

func preparePdf(inputFilename string) (pdf gopdf.GoPdf, err error) {
	pdf = gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: xPaperSize, H: yPaperSize}, Unit: gopdf.Unit_PT})
	pdf.AddPage()
	err = pdf.AddTTFFont("FiraSans-Book", "/usr/share/fonts/TTF/FiraSans-Book.ttf")
	if err != nil {
		return
	}
	err = pdf.AddTTFFont("FiraSans-Medium", "/usr/share/fonts/TTF/FiraSans-Medium.ttf")
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
	pdf.SetTextColor(0x00, 0x00, 0x00)
	err = pdf.SetFont("FiraSans-Book", "", 18)
	if err != nil {
		return
	}
	pdf.SetX(4)
	pdf.SetY(22)
	err = pdf.Text("TRex Packets Chart for ")
	if err != nil {
		return
	}
	err = pdf.SetFont("FiraSans-Medium", "", 18)
	if err != nil {
		return
	}
	err = pdf.Text(inputFilename)
	if err != nil {
		return
	}
	return nil
}

func footnotePdf(pdf *gopdf.GoPdf) (err error) {
	err = pdf.SetFont("FiraSans-Book", "", 8)
	if err != nil {
		return
	}
	pdf.SetTextColor(0x00, 0x33, 0x99)

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

func drawPackets(pdf *gopdf.GoPdf, packets []packet.Packet) {
	xMin, xMax, yMin, yMax := maxPacketsValue(packets)
	xScale, yScale := (xPaperSize-20.0)/float64(xMax-xMin), (yPaperSize-50.0)/(yMax-yMin)
	xPaperOffset, yPaperOffset := 10.0, 30.0
	yPaperBottom := yPaperSize - 50.0 + yPaperOffset
	yPaperZero := yPaperBottom - yMin*yScale
	fmt.Printf("y range: %v .. %v\n", yMin, yMax)

	// draw packets (first other, then the rest)
	pdf.SetLineWidth(xScale)
	for _, pkt := range packets {
		if pkt.Type() != packet.TypeOther {
			continue
		}
		x := float64(pkt.ReceivedAt().UnixNano() - xMin)
		y := pkt.Value()
		pdf.SetStrokeColor(pktColor(pkt))
		xOnPaper := xPaperOffset + x*xScale
		yOnPaper := yPaperBottom - (y-yMin)*yScale
		pdf.Line(xOnPaper, yPaperZero, xOnPaper, yOnPaper)
	}
	for _, pkt := range packets {
		if pkt.Type() == packet.TypeOther {
			continue
		}
		x := float64(pkt.ReceivedAt().UnixNano() - xMin)
		y := pkt.Value()
		pdf.SetStrokeColor(pktColor(pkt))
		xOnPaper := xPaperOffset + x*xScale
		yOnPaper := yPaperBottom - (y+yMin)*yScale
		pdf.Line(xOnPaper, yPaperZero, xOnPaper, yOnPaper)
	}

	// draw axis
	pdf.SetStrokeColor(0, 0, 0)
	pdf.Line(xPaperOffset, yPaperZero, xPaperSize-20.0, yPaperZero)
	//step := math.Pow10(int(math.Ceil(math.Log10(yMax-yMin))) - 1)
}

func maxPacketsValue(packets []packet.Packet) (xMin int64, xMax int64, yMin float64, yMax float64) {
	xMin, xMax = int64(1<<63-1), -int64(1<<63-1)
	yMin, yMax = float64(xMin), float64(xMax)
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

func pktColor(pkt packet.Packet) (r uint8, g uint8, b uint8) {
	switch pkt.Type() {
	case packet.TypeLatency:
		return 0xb8, 0xbb, 0x26
	case packet.TypePTP:
		return 0xcc, 0x24, 0x1d
	case packet.TypeOther:
		return 0xeb, 0xdb, 0xb2
	}
	return 0xff, 0x00, 0x0
}
