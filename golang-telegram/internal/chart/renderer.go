package chart

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	gochart "github.com/wcharczuk/go-chart/v2"

	"golang-springboot-monitor-bot/internal/model"
)

type Renderer struct{}

func NewRenderer() *Renderer { return &Renderer{} }

func (renderer *Renderer) Render(out io.Writer, serviceName, metricName, rangeText string, samples []model.MetricSample, thresholds []float64) error {
	if len(samples) == 0 {
		return fmt.Errorf("cannot render chart without samples")
	}

	scale := metricScaleFor(metricName, samples)
	times := make([]time.Time, 0, len(samples))
	values := make([]float64, 0, len(samples))
	for _, sample := range samples {
		times = append(times, sample.CollectedAt)
		values = append(values, sample.Value*scale.factor)
	}
	series := []gochart.Series{gochart.TimeSeries{
		Name:    metricName,
		XValues: times,
		YValues: values,
	}}
	yRange := dataYRange(values)
	for _, threshold := range thresholds {
		scaledThreshold := threshold * scale.factor
		if scaledThreshold < yRange.Min || scaledThreshold > yRange.Max {
			continue
		}
		series = append(series, gochart.TimeSeries{
			Name:    fmt.Sprintf("門檻值 %s", scale.format(scaledThreshold)),
			Style:   gochart.Style{StrokeColor: gochart.ColorRed, StrokeWidth: 1},
			XValues: []time.Time{times[0], times[len(times)-1]},
			YValues: []float64{scaledThreshold, scaledThreshold},
		})
	}

	graph := gochart.Chart{
		Width:  1000,
		Height: 560,
		Title:  fmt.Sprintf("%s | %s | %s", serviceName, metricName, rangeText),
		XAxis:  gochart.XAxis{Name: "時間", ValueFormatter: trendTimeFormatter(rangeText)},
		YAxis:  gochart.YAxis{Name: scale.unit, ValueFormatter: scale.valueFormatter(), Range: yRange},
		Series: series,
	}
	graph.Elements = []gochart.Renderable{gochart.Legend(&graph)}
	return graph.Render(gochart.PNG, out)
}

type metricScale struct {
	factor   float64
	unit     string
	decimals int
}

func metricScaleFor(metricName string, samples []model.MetricSample) metricScale {
	maxValue := 0.0
	for _, sample := range samples {
		maxValue = math.Max(maxValue, math.Abs(sample.Value))
	}
	switch {
	case metricName == "process_cpu_usage":
		return metricScale{factor: 100, unit: "百分比（%）", decimals: 2}
	case strings.HasSuffix(metricName, "_bytes"):
		const (
			kib = 1024
			mib = 1024 * kib
			gib = 1024 * mib
		)
		switch {
		case maxValue >= gib:
			return metricScale{factor: 1.0 / gib, unit: "GB", decimals: 2}
		case maxValue >= mib:
			return metricScale{factor: 1.0 / mib, unit: "MB", decimals: 2}
		case maxValue >= kib:
			return metricScale{factor: 1.0 / kib, unit: "KB", decimals: 2}
		default:
			return metricScale{factor: 1, unit: "bytes", decimals: 0}
		}
	case strings.HasSuffix(metricName, "_count"):
		return metricScale{factor: 1, unit: "次數", decimals: 0}
	case strings.Contains(metricName, "_seconds") && maxValue < 1:
		return metricScale{factor: 1000, unit: "毫秒", decimals: 2}
	case strings.Contains(metricName, "_seconds"):
		return metricScale{factor: 1, unit: "秒", decimals: 2}
	case strings.Contains(metricName, "connections_"):
		return metricScale{factor: 1, unit: "連線數", decimals: 0}
	case strings.HasSuffix(metricName, "_threads"):
		return metricScale{factor: 1, unit: "執行緒數", decimals: 0}
	default:
		return metricScale{factor: 1, unit: "數值", decimals: 2}
	}
}

func dataYRange(values []float64) *gochart.ContinuousRange {
	minValue, maxValue := values[0], values[0]
	for _, value := range values[1:] {
		minValue = math.Min(minValue, value)
		maxValue = math.Max(maxValue, value)
	}
	if minValue == maxValue {
		switch {
		case maxValue > 0:
			minValue = 0
		case minValue < 0:
			maxValue = 0
		default:
			maxValue = 1
		}
	}
	return &gochart.ContinuousRange{Min: minValue, Max: maxValue}
}

func (scale metricScale) format(value float64) string {
	return strconv.FormatFloat(value, 'f', scale.decimals, 64)
}

func (scale metricScale) valueFormatter() gochart.ValueFormatter {
	return func(value interface{}) string {
		switch typed := value.(type) {
		case float64:
			return scale.format(typed)
		case int:
			return scale.format(float64(typed))
		default:
			return fmt.Sprint(value)
		}
	}
}

func trendTimeFormatter(rangeText string) gochart.ValueFormatter {
	layout := "01-02 15:04"
	if rangeText == "1h" {
		layout = "15:04:05"
	}
	return func(value interface{}) string {
		switch typed := value.(type) {
		case time.Time:
			return typed.Format(layout)
		case float64:
			return gochart.TimeFromFloat64(typed).Format(layout)
		default:
			return fmt.Sprint(value)
		}
	}
}
