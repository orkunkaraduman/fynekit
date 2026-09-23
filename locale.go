package fynekit

import (
	"fyne.io/fyne/v2"
	"github.com/jeandeaual/go-locale"
	"golang.org/x/text/language"
)

var (
	currentLocale fyne.Locale
)

// CurrentLocale returns the closest supported locale to either the system locale or the overridden locale.
// The package uses this locale.
func CurrentLocale() fyne.Locale {
	return currentLocale
}

// OverrideLocale overrides the locale with the closest supported locale
// to the given locale. If an empty string is given, the package uses its default
// behavior.
func OverrideLocale(loc fyne.Locale) {
	if loc != "" {
		currentLocale = ClosestSupportedLocale([]string{loc.String()})
		SetupLang(currentLocale.LanguageString())
		return
	}

	all, err := locale.GetLocales()
	if err != nil {
		fyne.LogError("Failed to load user locales", err)
		all = []string{"en"}
	}
	currentLocale = ClosestSupportedLocale(all)
	SetupLang(currentLocale.LanguageString())
}

func LocaleFromLang(in string) fyne.Locale {
	t, e := language.Parse(in)
	if e != nil {
		return ""
	}

	return LocaleFromTag(t)
}
