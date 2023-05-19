package esox

import (
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/hlog"
	"github.com/xremming/abborre/esox/flash"
	"github.com/xremming/abborre/esox/utils"
)

type Renderer struct {
	fs     fs.FS
	prefix string
}

func NewRenderer(fs fs.FS, prefix string) *Renderer {
	return &Renderer{fs, prefix}
}

type Page struct {
	tmpl *template.Template
	name string
}

func (re *Renderer) GetTemplate(name, baseName string) *Page {
	baseFile, err := re.fs.Open(filepath.Join(re.prefix, baseName))
	if err != nil {
		panic(err)
	}
	defer baseFile.Close()

	childFile, err := re.fs.Open(filepath.Join(re.prefix, name))
	if err != nil {
		panic(err)
	}
	defer childFile.Close()

	buf := utils.GetBytesBuffer()
	defer utils.PutBytesBuffer(buf)

	_, err = io.Copy(buf, baseFile)
	if err != nil {
		panic(err)
	}

	tmpl, err := template.New(name).Parse(buf.String())
	if err != nil {
		panic(err)
	}

	buf.Reset()
	_, err = io.Copy(buf, childFile)
	if err != nil {
		panic(err)
	}

	tmpl, err = tmpl.Parse(buf.String())
	if err != nil {
		panic(err)
	}

	return &Page{tmpl, name}
}

func checkFlashCookie(w http.ResponseWriter, r *http.Request) bool {
	flashCookieDeleted := false
	cookie, err := r.Cookie("flash")
	if err == nil {
		flashCookieDeleted = true
		flashes := flash.Decode(cookie.Value)
		*r = *r.WithContext(flash.NewContext(r.Context(), flashes))
		http.SetCookie(w, &http.Cookie{
			Name:    "flash",
			MaxAge:  -1,
			Expires: time.Now().Add(-1 * time.Hour),
		})
	}

	return flashCookieDeleted
}

func setFlashCookie(w http.ResponseWriter, r *http.Request, redirect bool, flashes []flash.Data) {
	flashCookieDeleted := checkFlashCookie(w, r)
	if !redirect || flashCookieDeleted {
		return
	}

	if len(flashes) == 0 {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "flash",
		Value:    flash.Encode(flashes),
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
	})
}

type RenderData interface {
	SetFlashes(flashes []flash.Data)
}

func (p *Page) Render(w http.ResponseWriter, r *http.Request, code int, data RenderData) {
	log := hlog.FromRequest(r).With().
		Int("code", code).
		Str("template", p.name).
		Logger()

	buf := utils.GetBytesBuffer()
	defer utils.PutBytesBuffer(buf)

	flashes := flash.FromRequest(r)
	setFlashCookie(w, r, false, flashes)
	data.SetFlashes(flashes)

	err := p.tmpl.Execute(buf, data)
	if err != nil {
		log.Error().Err(err).Msg("failed to execute template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)

	_, err = buf.WriteTo(w)
	if err != nil {
		log.Error().Err(err).Msg("failed to write template to response")
	}
}

func Redirect(w http.ResponseWriter, r *http.Request, url string, code int) {
	setFlashCookie(w, r, true, flash.FromRequest(r))
	http.Redirect(w, r, url, code)
}