package fynekit

import (
	"reflect"
	"unsafe"

	"fyne.io/fyne/v2"
	"golang.org/x/text/language"
)

//go:linkname SetMainGoroutine fyne.io/fyne/v2/internal/async.SetMainGoroutine
func SetMainGoroutine()

//go:linkname GetWidgetRenderer fyne.io/fyne/v2/internal/cache.Renderer
func GetWidgetRenderer(wid fyne.Widget) fyne.WidgetRenderer

//go:linkname applyTheme fyne.io/fyne/v2/app.(*settings).applyTheme
func applyTheme(settings unsafe.Pointer, theme fyne.Theme, variant fyne.ThemeVariant)

func ApplyTheme(settings fyne.Settings, th fyne.Theme, variant fyne.ThemeVariant) {
	applyTheme(reflect.ValueOf(settings).UnsafePointer(), th, variant)
}

func SetTheme(settings fyne.Settings, th fyne.Theme) {
	ApplyTheme(settings, th, settings.ThemeVariant())
}

func SetThemeVariant(settings fyne.Settings, variant fyne.ThemeVariant) {
	ApplyTheme(settings, settings.Theme(), variant)
}

//go:linkname SetupLang fyne.io/fyne/v2/lang.setupLang
func SetupLang(lang string)

//go:linkname ClosestSupportedLocale fyne.io/fyne/v2/lang.closestSupportedLocale
func ClosestSupportedLocale(locs []string) fyne.Locale

//go:linkname LocaleFromTag fyne.io/fyne/v2/lang.localeFromTag
func LocaleFromTag(in language.Tag) fyne.Locale
