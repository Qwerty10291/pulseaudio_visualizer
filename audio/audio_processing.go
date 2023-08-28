package audio

import (
	"math"
	"time"

	"github.com/jfreymuth/pulse"
)

type AudioProcessor struct {
	client *pulse.Client
	samples int
	currIndex int
	stream *pulse.RecordStream
	dt time.Time
	inBuffer []float32
	inWin []float32
	
	outFFT []complex128
	outLog []float64
	outSmooth []float64
	callback func([]float64)
	
	VolumeMult float32
	SmoothStep float64
}

func NewAudioProcessor(buffSize int, callback func([]float64)) (*AudioProcessor, error) {
	cl, err := pulse.NewClient()
	if err != nil{
		return nil, err
	}
	sampleRate := buffSize / 2
	p := &AudioProcessor{
		client: cl,
		samples: sampleRate,
		dt: time.Now(),
		inBuffer: make([]float32, buffSize),
		inWin: make([]float32, buffSize),
		outFFT: make([]complex128, buffSize),
		outLog: make([]float64, buffSize),
		outSmooth: make([]float64, buffSize),
		VolumeMult: 2,
		SmoothStep: 3,
		callback: callback,
	}
	sink, err := cl.DefaultSink()
	stream, err := cl.NewRecord(pulse.Float32Writer(p.samplesHandler),  pulse.RecordMonitor(sink), pulse.RecordBufferFragmentSize(uint32(buffSize)))
	if err != nil{
		return nil, err
	}
	p.stream = stream


	return p, nil
} 

func (p *AudioProcessor) Start() {
	p.stream.Start()
}

func (p *AudioProcessor) Stop() {
	p.stream.Stop()
}

func (p *AudioProcessor) samplesHandler(data []float32) (int, error) {
	
	copy(p.inBuffer[p.currIndex:], data)
	p.currIndex += len(data)
	if p.currIndex >= p.samples {
		p.processSamples(p.inBuffer[:p.samples])
		p.currIndex = copy(p.inBuffer, p.inBuffer[p.samples:p.currIndex])
	}
	return 0, nil
}

func (p *AudioProcessor) processSamples(samples []float32) {
	for i := range samples{
		t := float64(i) / (float64(len(samples)) - 1)
		hann := 0.5 - 0.5 * math.Cos(2 * math.Pi * t)
		p.inWin[i] = samples[i] * float32(hann) * p.VolumeMult
	}
	p.FFT(p.inWin, 1, p.outFFT, len(samples))
	step := 1.1
	lowf := 1.0
	m := 0
	max_amp := 1.0
	for f := lowf; f < float64(len(samples) / 2); f = math.Ceil(f * step) {
		f1 := math.Ceil(f * step)
		a := .0
		for q := int(f); q < len(samples) / 2 && q < int(f1); q++{
			b := p.amp(p.outFFT[q])
			if b > a{
				a = b
			}
		}
		if max_amp < a{
			max_amp = a
		}

		p.outLog[m] = a
		m++
	}
	dt := float64(time.Since(p.dt).Milliseconds()) / 1000
	for i := 0; i < m; i++ {
		p.outLog[i] /= max_amp
		p.outSmooth[i] += (p.outLog[i] - p.outSmooth[i]) * p.SmoothStep * dt

	}
	p.dt = time.Now()
	p.callback(p.outSmooth[:m])

}

func (p *AudioProcessor) FFT(in []float32, stride int, out []complex128, n int) {
	if n <= 0 {
		return
	}

	if n == 1 {
		out[0] = complex(float64(in[0]), 0)
		return
	}

	p.FFT(in, stride*2, out, n/2)
	p.FFT(in[stride:], stride*2, out[n/2:], n/2)

	for k := 0; k < n/2; k++ {
		t := float64(k) / float64(n)
		f := -2 * math.Pi * t
		vImag, vReal := math.Sincos(f)
		v := complex(vReal, vImag)
		e := out[k]
		out[k] = e + v*out[k+n/2]
		out[k+n/2] = e - v*out[k+n/2]
	}
}

func (p *AudioProcessor) amp(f complex128) float64 {
	a := real(f)
	b := imag(f)
	return math.Log(a * a + b * b)
}