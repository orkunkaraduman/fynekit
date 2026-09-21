package fynekit

import (
	"log"

	"fyne.io/fyne/v2"
	"github.com/jeandeaual/go-locale"
	"golang.org/x/text/language"
)

var (
	currentLocale fyne.Locale
)

// CurrentLocale returns the currently set locale. If no locale has been set previously,
// it returns an empty string and the package uses the system locale.
func CurrentLocale() fyne.Locale {
	return currentLocale
}

// SetLocale sets the closest supported locale to the given locale as the current locale.
// If an empty string is given, the package sets the system locale as the current locale.
func SetLocale(loc fyne.Locale) {
	currentLocale = ""
	if loc != "" {
		currentLocale = ClosestSupportedLocale([]string{loc.String()})
		SetupLang(currentLocale.LanguageString())
		return
	}

	all, err := locale.GetLocales()
	if err != nil {
		log.Printf("failed to load user locales: %v", err)
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

func LocaleFromTag(in language.Tag) fyne.Locale {
	b, s, r := in.Raw()
	ret := b.String()

	if r.String() != "ZZ" {
		ret += "-" + r.String()

		if s.String() != "Zzzz" {
			ret += "-" + s.String()
		}
	}

	return fyne.Locale(ret)
}
