package main

import (
	"fmt"
	"math"
	"strings"
	"unsafe"

	"github.com/ikemen-minimalist/gl/v3.3-core/gl"
)

const vertexShaderSrc = `#version 330 core
layout(location = 0) in vec2 aPos;
layout(location = 1) in vec2 aUV;
out vec2 vUV;
void main(){ vUV = aUV; gl_Position = vec4(aPos, 0.0, 1.0); }
` + "\x00"

const fragmentShaderSrc = `#version 330 core
in vec2 vUV;
uniform sampler2D uSprite;
uniform sampler2D uPalette;
uniform bool uIndexed;
uniform vec4 uTint;
out vec4 fragColor;
void main(){
	vec4 c;
	if(uIndexed){ float rawIndex = texture(uSprite, vUV).r; int index = int(rawIndex * 255.0 + 0.5); c = texelFetch(uPalette, ivec2(index, 0), 0); }
	else { c = texture(uSprite, vUV); }
	fragColor = c * uTint;
}
` + "\x00"

type GLRenderer struct {
	winW, winH   int32
	program      uint32
	vao, vbo     uint32
	uSprite      int32
	uPalette     int32
	uIndexed     int32
	uTint        int32
	spriteCache  map[*Sprite]uint32
	paletteCache map[*PaletteEntry]uint32
}

func NewGLRenderer(winW, winH int32) (*GLRenderer, error) {
	if err := gl.Init(); err != nil {
		return nil, fmt.Errorf("GL init failed: %v (update graphics drivers)", err)
	}

	gl.Viewport(0, 0, winW, winH)
	gl.Enable(gl.BLEND)
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)

	program, err := newProgram(vertexShaderSrc, fragmentShaderSrc)
	if err != nil {
		return nil, err
	}

	r := &GLRenderer{
		winW:         winW,
		winH:         winH,
		program:      program,
		spriteCache:  map[*Sprite]uint32{},
		paletteCache: map[*PaletteEntry]uint32{},
	}
	r.uSprite = gl.GetUniformLocation(program, gl.Str("uSprite\x00"))
	r.uPalette = gl.GetUniformLocation(program, gl.Str("uPalette\x00"))
	r.uIndexed = gl.GetUniformLocation(program, gl.Str("uIndexed\x00"))
	r.uTint = gl.GetUniformLocation(program, gl.Str("uTint\x00"))

	gl.GenVertexArrays(1, &r.vao)
	gl.GenBuffers(1, &r.vbo)
	gl.BindVertexArray(r.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 6*4*4, nil, gl.DYNAMIC_DRAW)

	stride := int32(16)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, stride, unsafe.Pointer(uintptr(0)))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, stride, unsafe.Pointer(uintptr(8)))
	gl.BindVertexArray(0)

	return r, nil
}

func (r *GLRenderer) BeginFrame() {
	gl.ClearColor(0.125, 0.125, 0.125, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)
}

func (r *GLRenderer) Destroy() {
	for _, t := range r.spriteCache {
		gl.DeleteTextures(1, &t)
	}
	for _, t := range r.paletteCache {
		gl.DeleteTextures(1, &t)
	}
	if r.vbo != 0 {
		gl.DeleteBuffers(1, &r.vbo)
	}
	if r.vao != 0 {
		gl.DeleteVertexArrays(1, &r.vao)
	}
	if r.program != 0 {
		gl.DeleteProgram(r.program)
	}
}

func setNearestClamp() {
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
}

func (r *GLRenderer) ensureSprite(sp *Sprite) uint32 {
	if t := r.spriteCache[sp]; t != 0 {
		return t
	}
	var tex uint32
	gl.GenTextures(1, &tex)
	gl.BindTexture(gl.TEXTURE_2D, tex)
	setNearestClamp()
	if sp.IsIndexed {
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.R8, int32(sp.W), int32(sp.H), 0, gl.RED, gl.UNSIGNED_BYTE, gl.Ptr(sp.Indexed))
	} else {
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, int32(sp.W), int32(sp.H), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(sp.RGBA))
	}
	r.spriteCache[sp] = tex
	return tex
}

func (r *GLRenderer) ensurePalette(p *PaletteEntry) uint32 {
	if p == nil {
		return 0
	}
	if t := r.paletteCache[p]; t != 0 {
		return t
	}
	var tex uint32
	gl.GenTextures(1, &tex)
	gl.BindTexture(gl.TEXTURE_2D, tex)
	setNearestClamp()
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, 256, 1, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(p.RGBA))
	r.paletteCache[p] = tex
	return tex
}

func (r *GLRenderer) RenderSprite(rp RenderParams) error {
	if rp.Sprite == nil {
		return nil
	}

	spTex := r.ensureSprite(rp.Sprite)
	palTex := uint32(0)
	if rp.Sprite.IsIndexed {
		palTex = r.ensurePalette(rp.Palette)
		if palTex == 0 {
			return fmt.Errorf("indexed sprite has no palette")
		}
	}

	verts := r.makeVertices(rp)

	switch rp.BlendMode {
	case TransAdd:
		gl.BlendEquation(gl.FUNC_ADD)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
	case TransSub:
		gl.BlendEquation(gl.FUNC_REVERSE_SUBTRACT)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
	default:
		gl.BlendEquation(gl.FUNC_ADD)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	}

	gl.UseProgram(r.program)
	gl.Uniform4f(r.uTint, rp.Tint[0], rp.Tint[1], rp.Tint[2], rp.Tint[3])
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, spTex)
	gl.Uniform1i(r.uSprite, 0)

	if rp.Sprite.IsIndexed {
		gl.ActiveTexture(gl.TEXTURE1)
		gl.BindTexture(gl.TEXTURE_2D, palTex)
		gl.Uniform1i(r.uPalette, 1)
		gl.Uniform1i(r.uIndexed, 1)
	} else {
		gl.Uniform1i(r.uIndexed, 0)
	}

	gl.BindVertexArray(r.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(verts)*4, gl.Ptr(verts))
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	gl.BindVertexArray(0)

	return nil
}

func (r *GLRenderer) makeVertices(rp RenderParams) []float32 {
	sp := rp.Sprite
	sx, sy := rp.ScaleX, rp.ScaleY
	if sx == 0 {
		sx = 1
	}
	if sy == 0 {
		sy = 1
	}

	w := math.Abs(float64(sp.W) * sx)
	h := math.Abs(float64(sp.H) * sy)
	cx := rp.X + w/2
	cy := rp.Y + h/2

	rad := rp.Angle * math.Pi / 180
	ca := math.Cos(rad)
	sa := math.Sin(rad)

	u0, u1 := float32(0), float32(1)
	v0, v1 := float32(0), float32(1)
	if rp.FlipX || sx < 0 {
		u0, u1 = u1, u0
	}
	if rp.FlipY || sy < 0 {
		v0, v1 = v1, v0
	}

	type pt struct {
		x, y float64
		u, v float32
	}
	corners := []pt{
		{-w / 2, -h / 2, u0, v0},
		{w / 2, -h / 2, u1, v0},
		{w / 2, h / 2, u1, v1},
		{-w / 2, h / 2, u0, v1},
	}
	idx := []int{0, 1, 2, 0, 2, 3}

	verts := make([]float32, 0, 24)
	for _, id := range idx {
		c := corners[id]
		rx := c.x*ca - c.y*sa
		ry := c.x*sa + c.y*ca
		screenX := cx + rx
		screenY := cy + ry
		ndcX := float32(screenX/(float64(r.winW)/2) - 1)
		ndcY := float32(1 - screenY/(float64(r.winH)/2))
		verts = append(verts, ndcX, ndcY, c.u, c.v)
	}
	return verts
}

func newProgram(vsrc, fsrc string) (uint32, error) {
	vs, err := compileShader(vsrc, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}
	defer gl.DeleteShader(vs)

	fs, err := compileShader(fsrc, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}
	defer gl.DeleteShader(fs)

	p := gl.CreateProgram()
	gl.AttachShader(p, vs)
	gl.AttachShader(p, fs)
	gl.LinkProgram(p)

	var status int32
	gl.GetProgramiv(p, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var l int32
		gl.GetProgramiv(p, gl.INFO_LOG_LENGTH, &l)
		log := strings.Repeat("\x00", int(l+1))
		gl.GetProgramInfoLog(p, l, nil, gl.Str(log))
		return 0, fmt.Errorf("program link failed: %s", log)
	}
	return p, nil
}

func compileShader(src string, typ uint32) (uint32, error) {
	s := gl.CreateShader(typ)
	csrc, free := gl.Strs(src)
	defer free()
	gl.ShaderSource(s, 1, csrc, nil)
	gl.CompileShader(s)

	var status int32
	gl.GetShaderiv(s, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var l int32
		gl.GetShaderiv(s, gl.INFO_LOG_LENGTH, &l)
		log := strings.Repeat("\x00", int(l+1))
		gl.GetShaderInfoLog(s, l, nil, gl.Str(log))
		return 0, fmt.Errorf("shader compile failed: %s", log)
	}
	return s, nil
}
