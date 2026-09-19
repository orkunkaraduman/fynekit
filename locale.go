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

func CurrentLocale() fyne.Locale {
	return currentLocale
}

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
	SetupLang(ClosestSupportedLocale(all).LanguageString())
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
