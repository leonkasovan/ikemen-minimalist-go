package main

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/ikemen-minimalist/go-sdl2/sdl"
)

const (
	windowW int32 = 640
	windowH int32 = 480
)

func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	_ = sdl.ShowSimpleMessageBox(sdl.MESSAGEBOX_ERROR, "ikemen-minimalist", fmt.Sprintf(format, a...), nil)
	os.Exit(1)
}

func setAnimation(list []*Animation, index int) (*Animation, int) {
	if len(list) == 0 {
		fatalf("no drawable animations")
	}
	index = (index%len(list) + len(list)) % len(list)
	a := list[index]
	a.Reset()
	return a, index
}

func blendName(t TransType) string {
	switch t {
	case TransAdd:
		return "add"
	case TransSub:
		return "sub"
	default:
		return "alpha"
	}
}

func updateTitle(win *sdl.Window, anim *Animation, inst *AnimationInstance, idx, total int) {
	title := fmt.Sprintf("ikemen-minimalist | action %d (%d/%d) | blend %s | pal %s", anim.No, idx+1, total, blendName(inst.BlendMode), inst.PaletteName())
	win.SetTitle(title)
}

func main() {
	runtime.LockOSThread()
	os.Stdout.Sync()

	// Parse optional -log flag before positional args.
	// This lets GUI-subsystem builds capture stdout to a file.
	args := os.Args[1:]
	logPath := ""
	posArgs := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if args[i] == "-log" && i+1 < len(args) {
			logPath = args[i+1]
			i++ // skip the value
		} else {
			posArgs = append(posArgs, args[i])
		}
	}

	if logPath != "" {
		f, err := os.Create(logPath)
		if err != nil {
			fatalf("log file: %v", err)
		}
		os.Stdout = f
		defer f.Close()
	}

	if len(posArgs) != 2 {
		fatalf("usage: %s [-log <file>] <char.sff> <char.air>", os.Args[0])
	}

	sffPath, airPath := posArgs[0], posArgs[1]

	rand.Seed(time.Now().UnixNano())

	fmt.Printf("Loading SFF: %s...\n", sffPath)
	sff, err := LoadSFF(sffPath)
	if err != nil {
		fatalf("SFF: %v", err)
	}
	fmt.Printf("  sprites: %d, palettes: %d\n", len(sff.Sprites), len(sff.Palettes))

	fmt.Printf("Loading AIR: %s...\n", airPath)
	anims, err := LoadAIR(airPath)
	if err != nil {
		fatalf("AIR: %v", err)
	}
	fmt.Printf("  animations: %d\n", len(anims))

	diagnostics := BuildRenderDiagnostics(sff, anims)
	diagnostics.Print(os.Stdout)

	drawable := DrawableAnimations(anims, sff)
	fmt.Printf("  drawable: %d\n", len(drawable))
	if len(drawable) == 0 {
		fatalf("no AIR animations reference sprites found in SFF")
	}

	anim, animIndex := setAnimation(drawable, rand.Intn(len(drawable)))
	inst := NewAnimationInstance(sff, anim)

	fmt.Println("Starting SDL...")

	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_EVENTS | sdl.INIT_TIMER); err != nil {
		fatalf("SDL: %v", err)
	}
	defer sdl.Quit()

	sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 3)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 3)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE)
	sdl.GLSetAttribute(sdl.GL_DOUBLEBUFFER, 1)
	sdl.SetHint(sdl.HINT_WINDOWS_NO_CLOSE_ON_ALT_F4, "0")

	win, err := sdl.CreateWindow("ikemen-minimalist", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, windowW, windowH, sdl.WINDOW_OPENGL|sdl.WINDOW_SHOWN)
	if err != nil {
		fatalf("window: %v", err)
	}
	defer win.Destroy()
	win.Raise()

	ctx, err := win.GLCreateContext()
	if err != nil {
		fatalf("GL context: %v", err)
	}
	defer sdl.GLDeleteContext(ctx)

	if err := win.GLMakeCurrent(ctx); err != nil {
		fatalf("GL make current: %v", err)
	}
	sdl.GLSetSwapInterval(1)

	renderer, err := NewGLRenderer(windowW, windowH)
	if err != nil {
		fatalf("renderer: %v", err)
	}
	defer renderer.Destroy()

	updateTitle(win, anim, inst, animIndex, len(drawable))

	dl := &DrawList{}
	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()

	running := true
	for running {
		for e := sdl.PollEvent(); e != nil; e = sdl.PollEvent() {
			switch ev := e.(type) {
			case sdl.QuitEvent:
				running = false
			case sdl.WindowEvent:
				if ev.Event == sdl.WINDOWEVENT_CLOSE {
					running = false
				}
			case sdl.KeyboardEvent:
				if ev.Type != sdl.KEYDOWN || ev.Repeat != 0 {
					continue
				}
				changed := true
				switch ev.Keysym.Scancode {
				case sdl.SCANCODE_ESCAPE:
					running = false
				case sdl.SCANCODE_RIGHT, sdl.SCANCODE_SPACE, sdl.SCANCODE_N:
					anim, animIndex = setAnimation(drawable, animIndex+1)
					inst.SetAnim(anim)
				case sdl.SCANCODE_LEFT, sdl.SCANCODE_P:
					anim, animIndex = setAnimation(drawable, animIndex-1)
					inst.SetAnim(anim)
				case sdl.SCANCODE_R:
					anim, animIndex = setAnimation(drawable, rand.Intn(len(drawable)))
					inst.SetAnim(anim)
				case sdl.SCANCODE_1:
					inst.BlendMode = TransAlpha
				case sdl.SCANCODE_2:
					inst.BlendMode = TransAdd
				case sdl.SCANCODE_3:
					inst.BlendMode = TransSub
				case sdl.SCANCODE_4:
					inst.ClearPalette()
				case sdl.SCANCODE_LEFTBRACKET:
					inst.CyclePaletteOverride(-1)
				case sdl.SCANCODE_RIGHTBRACKET:
					inst.CyclePaletteOverride(1)
				case sdl.SCANCODE_5:
					inst.RemapCurrentDefault(1)
				case sdl.SCANCODE_6:
					inst.RemapCurrentDefault(-1)
				default:
					changed = false
				}
				if changed {
					updateTitle(win, anim, inst, animIndex, len(drawable))
				}
			}
		}
		if !running {
			break
		}
		<-ticker.C
		inst.Update()
		dl.Clear()
		inst.Draw(dl)
		renderer.BeginFrame()
		if err := dl.DrawLayer(0, renderer); err != nil {
			fatalf("draw: %v", err)
		}
		win.GLSwap()
	}
}
