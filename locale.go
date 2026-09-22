package fynekit

import (
	"log"

	"fyne.io/fyne/v2"
	"github.com/jeandeaual/go-locale"
	"golang.org/x/text/language"
)

var (
// overridenLocale fyne.Locale
)

/*// OverridenLocale returns the currently overriden locale. If no locale has been overriden previously,
// it returns an empty string and the package uses its default behavior.
func OverridenLocale() fyne.Locale {
	return overridenLocale
}*/

// OverrideLocale overrides the locale with the closest supported locale
// to the given locale. If an empty string is given, the package uses its default
// behavior.
func OverrideLocale(loc fyne.Locale) {
	//overridenLocale = ""
	if loc != "" {
		/*overridenLocale = ClosestSupportedLocale([]string{loc.String()})
		SetupLang(overridenLocale.LanguageString())*/
		SetupLang(ClosestSupportedLocale([]string{loc.String()}).LanguageString())
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
