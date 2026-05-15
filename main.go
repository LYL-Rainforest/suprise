package main

import (
	"embed"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/speaker"
)

//go:embed bgm.flac
var res embed.FS

const (
	clear  = "\033[2J\033[H"
	hide   = "\033[?25l"
	width  = 100
	height = 40
)

func main() {
	// 1. 初始化音频并等待缓冲区准备就绪
	playReady := make(chan bool)
	go playMusic(playReady)
	<-playReady // 阻塞直到音乐真正开始播放

	fmt.Print(hide)
	startTime := time.Now()

	for {
		elapsed := time.Since(startTime).Seconds()
		var sb strings.Builder
		sb.WriteString(clear)

		switch {
		case elapsed < 3.0:
			renderRainbowCountdown(&sb, elapsed)
		case elapsed < 7.5:
			renderRainbowFirework(&sb, elapsed-3.0)
		case elapsed < 22.5: // 超长覆盖：15秒心跳狂欢
			renderDrumHeartRainbow(&sb, elapsed-7.5)
		default:
			// 终极炫彩对冲
			renderContrastRainbowLove(&sb, elapsed-22.5)
		}

		fmt.Print(sb.String())
		time.Sleep(20 * time.Millisecond)
	}
}

func playMusic(ready chan bool) {
	f, err := res.Open("bgm.flac")
	if err != nil {
		return
	}
	defer f.Close()

	streamer, format, err := flac.Decode(f.(io.ReadCloser))
	if err != nil {
		return
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	// 播放并通知主线程开始计时
	speaker.Play(beep.Seq(streamer))
	ready <- true
}

// 虹彩生成函数：基于时间生成 RGB
func getRainbow(t float64, offset float64) (int, int, int) {
	r := int(127 + 127*math.Sin(t+offset))
	g := int(127 + 127*math.Sin(t+offset+2.094)) // +120度
	b := int(127 + 127*math.Sin(t+offset+4.188)) // +240度
	return r, g, b
}

// 1. 虹彩倒计时
func renderRainbowCountdown(sb *strings.Builder, t float64) {
	fonts := map[int][]string{
		3: {"33333", "    3", " 3333", "    3", "33333"},
		2: {"22222", "    2", " 2222", "2    ", "22222"},
		1: {"  1  ", " 11  ", "  1  ", "  1  ", " 1111"},
	}
	screen := createScreen()
	num := 3 - int(t)
	if num < 1 {
		num = 1
	}
	offX, offY := 45, 15
	r, g, b := getRainbow(t*5, 0)
	for row, line := range fonts[num] {
		for col, char := range line {
			if char != ' ' {
				screen[offY+row][offX+col] = fmt.Sprintf("\033[38;2;%d;%d;%dm%c\033[0m", r, g, b, char)
			}
		}
	}
	drawScreen(sb, screen)
}

// 2. 虹彩烟花
func renderRainbowFirework(sb *strings.Builder, t float64) {
	screen := createScreen()
	cycleTime := 0.9
	localT := math.Mod(t, cycleTime)
	offsets := []int{25, 45, 50, 55, 75}
	cX := offsets[int(t/cycleTime)%5]
	cY := 15
	if localT < 0.45 {
		py := int(38 - (localT/0.45)*25)
		if py > 0 && py < height {
			r, g, b := getRainbow(t*10, 0)
			screen[py][cX] = fmt.Sprintf("\033[38;2;%d;%d;%dm*\033[0m", r, g, b)
		}
	} else {
		dt := localT - 0.45
		scale := dt * 45.0
		for i := 0.0; i < math.Pi*2; i += 0.1 {
			tx := 16 * math.Pow(math.Sin(i), 3)
			ty := -(13*math.Cos(i) - 5*math.Cos(2*i) - 2*math.Cos(3*i) - math.Cos(4*i))
			px, py := int(float64(cX)+tx*scale*0.08), int(float64(cY)+ty*scale*0.08+dt*dt*15)
			if px > 0 && px < width && py > 0 && py < height {
				r, g, b := getRainbow(t*5, i)
				screen[py][px] = fmt.Sprintf("\033[38;2;%d;%d;%dm♥\033[0m", r, g, b)
			}
		}
	}
	drawScreen(sb, screen)
}

// 3. 虹彩打击感心跳 (BPM 128)
func renderDrumHeartRainbow(sb *strings.Builder, t float64) {
	screen := createScreen()
	beatPeriod := 60.0 / 128.0
	phase := math.Mod(t, beatPeriod) / beatPeriod
	impact := math.Pow(math.Sin(phase*math.Pi), 8)
	scale := 1.0 + 0.45*impact

	for y := 1.5; y > -1.5; y -= 0.1 {
		for x := -1.5; x < 1.5; x += 0.04 {
			xx, yy := x/scale, y/scale
			if math.Pow(xx*xx+yy*yy-1, 3)-xx*xx*yy*yy*yy <= 0 {
				px, py := int(50+x*25), int(18-y*12)
				if px > 0 && px < width && py > 0 && py < height {
					// 颜色随心跳深度旋转
					r, g, b := getRainbow(t*8, math.Atan2(y, x))
					screen[py][px] = fmt.Sprintf("\033[38;2;%d;%d;%dm♥\033[0m", r, g, b)
				}
			}
		}
	}
	drawScreen(sb, screen)
}

// 4. 终极对冲虹彩 LOVE
func renderContrastRainbowLove(sb *strings.Builder, t float64) {
	screen := createScreen()
	asciiLove := []string{
		"L      OOO  V   V EEEEE",
		"L     O   O V   V E    ",
		"L     O   O V   V EEEEE",
		"L     O   O  V V  E    ",
		"LLLLL  OOO    V   EEEEE",
	}
	offX, offY := 35, 15
	for row, line := range asciiLove {
		for col, char := range line {
			if char != ' ' {
				px, py := offX+col, offY+row
				if px < width && py < height {
					var r, g, b int
					// 左右对冲颜色：左边使用正常色相，右边偏移 180 度(math.Pi)
					if col < 12 {
						r, g, b = getRainbow(t*15, float64(col)*0.1)
					} else {
						r, g, b = getRainbow(t*15, float64(col)*0.1+math.Pi)
					}
					// 增加全局闪烁感
					if int(t*30)%4 == 0 {
						r, g, b = 255, 255, 255
					}
					screen[py][px] = fmt.Sprintf("\033[38;2;%d;%d;%dm%c\033[0m", r, g, b, char)
				}
			}
		}
	}
	drawScreen(sb, screen)
}

func createScreen() [][]string {
	s := make([][]string, height)
	for i := range s {
		s[i] = make([]string, width)
		for j := range s[i] {
			s[i][j] = " "
		}
	}
	return s
}

func drawScreen(sb *strings.Builder, screen [][]string) {
	for _, row := range screen {
		sb.WriteString(strings.Join(row, "") + "\n")
	}
}
