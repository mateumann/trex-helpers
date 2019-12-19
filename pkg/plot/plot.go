package plot

import (
	"fmt"
	"math"
	"os"
	"time"
	"trex-helpers/pkg/analytics"
	"trex-helpers/pkg/packet"

	"github.com/signintech/gopdf"
)

//const xPaperSize float64 = 842.0 * 3
//const yPaperSize float64 = 595.0

type plotter struct {
	xPaperSize    float64
	yPaperSize    float64
	xLeftMargin   float64
	xRightMargin  float64
	yTopMargin    float64
	yBottomMargin float64
	titlePrefix   string
	inputFilename string
	xMin          int64
	xMax          int64
	yMin          float64
	yMax          float64
	xScale        float64
	yScale        float64
	yZeroAt       float64
	xLineStep     int64
	yLineStep     int64
}

func (plot plotter) width() float64 {
	return plot.xPaperSize - plot.xLeftMargin - plot.xRightMargin
}

func (plot plotter) height() float64 {
	return plot.yPaperSize - plot.yTopMargin - plot.yBottomMargin
}

func (plot *plotter) fromPackets(packets []packet.Packet) {
	plot.xMin, plot.xMax, plot.yMin, plot.yMax = maxPacketsValue(packets)
	plot.xScale, plot.yScale = plot.width()/float64(plot.xMax-plot.xMin), plot.height()/(plot.yMax-plot.yMin)
	plot.yZeroAt = plot.yPaperSize - plot.yBottomMargin + plot.yMin*plot.yScale
}

type stats struct {
	averageLatency    float64
	periodicLatencies []analytics.PeriodicAvgLatency
}

func (sts *stats) fromPackets(packets []packet.Packet) {
	sts.averageLatency = analytics.CalcPositiveAverageLatency(packets)
	sts.periodicLatencies = analytics.CalcPeriodicAverageLatency(packets)
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
	//fmt.Printf("boundaries: x = %v .. %v, y = %v .. %v\n", xMin, xMax, yMin, yMax)
	return
}

func SavePDF(packets []packet.Packet, inputFilename string, filename string, verbose bool) (err error) {
	plot := plotter{
		xPaperSize:    842 * 4,
		yPaperSize:    595,
		xLeftMargin:   12,
		xRightMargin:  12,
		yTopMargin:    24,
		yBottomMargin: 12,
		titlePrefix:   "TRex Packets Chart",
		inputFilename: inputFilename,
		//outputFilename: filename,
	}
	plot.fromPackets(packets)

	stats := stats{averageLatency: 0}
	stats.fromPackets(packets)

	pdf, err := preparePdf(&plot, stats)
	if err != nil {
		return
	}

	drawPackets(&pdf, packets, &plot)

	drawAnalytics(&pdf, stats, &plot)

	drawAxis(&pdf, &plot)

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

func preparePdf(plot *plotter, sts stats) (pdf gopdf.GoPdf, err error) {
	pdf = gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: plot.xPaperSize, H: plot.yPaperSize}, Unit: gopdf.Unit_PT})
	pdf.SetInfo(gopdf.PdfInfo{
		Title:        fmt.Sprintf("%v for %v", plot.titlePrefix, plot.inputFilename),
		Subject:      plot.titlePrefix,
		Creator:      "trex-helpers",
		Producer:     "https://github.com/signintech/gopdf",
		CreationDate: time.Now(),
	})
	pdf.AddPage()
	err = pdf.AddTTFFont("FiraSans-Book", "/usr/share/fonts/TTF/FiraSans-Book.ttf")
	if err != nil {
		return
	}
	err = pdf.AddTTFFont("FiraSans-Medium", "/usr/share/fonts/TTF/FiraSans-Medium.ttf")
	if err != nil {
		return
	}

	err = makeTitle(&pdf, plot.inputFilename)
	if err != nil {
		return
	}
	err = makeFootnote(&pdf, plot)
	if err != nil {
		return
	}
	// due to some bug(?) in gopdf one cannot reliably write text on already “drawn” PDF page
	err = makeAxisAnnotations(&pdf, plot)
	if err != nil {
		return
	}
	err = makeStatsAnnotations(&pdf, sts, plot)
	if err != nil {
		return
	}

	return pdf, err
}

func makeTitle(pdf *gopdf.GoPdf, inputFilename string) (err error) {
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

func makeFootnote(pdf *gopdf.GoPdf, plot *plotter) (err error) {
	err = pdf.SetFont("FiraSans-Book", "", 8)
	if err != nil {
		return
	}
	pdf.SetX(4)
	pdf.SetY(plot.yPaperSize - 3)
	err = pdf.Text(fmt.Sprintf("%v", time.Now()))
	if err != nil {
		return err
	}
	pdf.SetX(plot.xPaperSize - 106)
	err = pdf.Text("generated with trex-helpers")
	if err != nil {
		return err
	}
	pdf.AddExternalLink("https://github.com/mateumann/trex-helpers",
		plot.xPaperSize-106.5, plot.yPaperSize-10.5, 105, 10)
	return nil
}

func makeAxisAnnotations(pdf *gopdf.GoPdf, plot *plotter) (err error) {
	pdf.SetTextColor(0, 0, 0)
	err = pdf.SetFont("FiraSans-Book", "", 12)
	if err != nil {
		return
	}

	for _, y := range verticalSteps(plot) {
		yOnPaper := plot.yPaperSize - plot.yBottomMargin - (y-plot.yMin)*plot.yScale
		err = makeAnnotation(pdf, plot.xLeftMargin, yOnPaper-4, 0, 0, 0, fmt.Sprintf("%v µs", y))
		if err != nil {
			return
		}
	}

	for _, x := range horizontalSteps(true, plot) {
		xOnPaper := plot.xLeftMargin + x*plot.xScale
		err = makeAnnotation(pdf, xOnPaper-6, plot.yZeroAt+16, 0, 0, 0, fmt.Sprintf("%v s", x/1000/1000/1000))
		if err != nil {
			return
		}
	}
	return nil
}

func verticalSteps(plot *plotter) (steps []float64) {
	plot.yLineStep = int64(math.Pow10(int(math.Ceil(math.Log10((plot.yMax-plot.yMin)/2))) - 1))
	lo := plot.yLineStep * (int64(plot.yMin) / plot.yLineStep)
	hi := plot.yLineStep * (int64(plot.yMax) / plot.yLineStep)
	for y := lo; y <= hi; y += plot.yLineStep {
		steps = append(steps, float64(y))
	}
	return
}

func horizontalSteps(forAnnotations bool, plot *plotter) (steps []float64) {
	if forAnnotations {
		plot.xLineStep = int64(math.Pow10(int(math.Ceil(math.Log10(float64(plot.xMax-plot.xMin)))) - 1))
	} else {
		plot.xLineStep = int64(math.Pow10(int(math.Ceil(math.Log10(float64(plot.xMax-plot.xMin)))) - 2))
	}
	//fmt.Printf("forAnnotations = %5v, xLineStep = %v\n", forAnnotations, plot.xLineStep)
	hi := plot.xLineStep * (plot.xMax / plot.xLineStep)
	for x := plot.xMin + plot.xLineStep; x <= hi+plot.xLineStep; x += plot.xLineStep {
		steps = append(steps, float64(x-plot.xMin))
	}
	return
}

func makeAnnotation(pdf *gopdf.GoPdf, x, y float64, r, g, b uint8, text string) (err error) {
	pdf.SetTextColor(r, g, b)
	pdf.SetX(x)
	pdf.SetY(y)
	err = pdf.Text(text)
	if err != nil {
		return
	}
	return nil
}
func makeStatsAnnotations(pdf *gopdf.GoPdf, sts stats, plot *plotter) (err error) {
	yOnPaper := plot.yPaperSize - plot.yBottomMargin - (sts.averageLatency-plot.yMin)*plot.yScale
	err = pdf.SetFont("FiraSans-Book", "", 12)
	if err != nil {
		return err
	}
	err = makeAnnotation(pdf, plot.xLeftMargin+20, yOnPaper-5, 0xd3, 0x86, 0x9b,
		fmt.Sprintf("avg. lat. %.2f µs", sts.averageLatency))
	if err != nil {
		return err
	}

	pdf.SetStrokeColor(0xfb, 0x49, 0x34)
	for _, periodicData := range sts.periodicLatencies {
		x0 := float64(periodicData.StartTimestamp.UnixNano() - plot.xMin)
		x0OnPaper := plot.xLeftMargin + x0*plot.xScale
		yOnPaper := plot.yPaperSize - plot.yBottomMargin - (periodicData.Value-plot.yMin)*plot.yScale
		if periodicData.Value < 0 {
			yOnPaper += 20
		}
		err = makeAnnotation(pdf, x0OnPaper+5, yOnPaper-5, 0xfb, 0x49, 0x34,
			fmt.Sprintf("%.2f µs", periodicData.Value))
		if err != nil {
			return err
		}
	}

	return nil
}

func drawAnalytics(pdf *gopdf.GoPdf, sts stats, plot *plotter) {
	yOnPaper := plot.yPaperSize - plot.yBottomMargin - (sts.averageLatency-plot.yMin)*plot.yScale
	pdf.SetStrokeColor(0xd3, 0x86, 0x9b)
	pdf.SetLineWidth(1)
	pdf.SetLineType("dashed")
	pdf.Line(plot.xLeftMargin, yOnPaper, plot.xPaperSize-plot.xRightMargin, yOnPaper)

	//pdf.SetStrokeColor(0xd3, 0x86, 0x9b)
	pdf.SetStrokeColor(0xfb, 0x49, 0x34)
	for _, periodicData := range sts.periodicLatencies {
		x0 := float64(periodicData.StartTimestamp.UnixNano() - plot.xMin)
		x1 := float64(periodicData.EndTimestamp.UnixNano() - plot.xMin)
		x0OnPaper := plot.xLeftMargin + x0*plot.xScale
		x1OnPaper := plot.xLeftMargin + x1*plot.xScale
		yOnPaper := plot.yPaperSize - plot.yBottomMargin - (periodicData.Value-plot.yMin)*plot.yScale
		pdf.Line(x0OnPaper, yOnPaper, x1OnPaper, yOnPaper)
	}
}

func drawPackets(pdf *gopdf.GoPdf, packets []packet.Packet, plot *plotter) {
	// first draw "other" packets, then the rest
	pdf.SetLineWidth(plot.xScale)
	pdf.SetLineType("solid")
	for _, pkt := range packets {
		if pkt.Type() != packet.TypeOther {
			continue
		}
		makeLine(pdf, pkt, plot)
	}
	for _, pkt := range packets {
		if pkt.Type() == packet.TypeOther {
			continue
		}
		makeLine(pdf, pkt, plot)
	}
}

func makeLine(pdf *gopdf.GoPdf, pkt packet.Packet, plot *plotter) {
	x := float64(pkt.ReceivedAt().UnixNano() - plot.xMin)
	y := pkt.Value()
	pdf.SetStrokeColor(pktColor(pkt))
	xOnPaper := plot.xLeftMargin + x*plot.xScale
	yOnPaper := plot.yPaperSize - plot.yBottomMargin - (y-plot.yMin)*plot.yScale
	pdf.Line(xOnPaper, plot.yZeroAt, xOnPaper, yOnPaper)
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

func drawAxis(pdf *gopdf.GoPdf, plot *plotter) {
	// main X axis
	pdf.SetStrokeColor(0, 0, 0)
	pdf.SetLineWidth(1)
	pdf.SetLineType("solid")
	pdf.Line(plot.xLeftMargin, plot.yZeroAt, plot.xPaperSize-plot.xRightMargin, plot.yZeroAt)

	// X axis marks
	for _, x := range horizontalSteps(false, plot) {
		xOnPaper := plot.xLeftMargin + x*plot.xScale
		pdf.Line(xOnPaper, plot.yZeroAt-4, xOnPaper, plot.yZeroAt+4)
	}

	// helper X lines
	pdf.SetStrokeColor(0x66, 0x66, 0x66)
	pdf.SetLineWidth(0.01)
	pdf.SetLineType("dotted")
	for _, y := range verticalSteps(plot) {
		yOnPaper := plot.yPaperSize - plot.yBottomMargin - (y-plot.yMin)*plot.yScale
		pdf.Line(plot.xLeftMargin, yOnPaper, plot.xPaperSize-plot.xRightMargin, yOnPaper)
	}
}
