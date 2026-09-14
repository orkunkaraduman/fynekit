//go:build ios

package fynekit

/*
#import <UIKit/UIKit.h>

BOOL isDarkMode(void) {
	if (@available(iOS 12.0, *)) {
		UIUserInterfaceStyle style =
			UIScreen.mainScreen.traitCollection.userInterfaceStyle;

		return style == UIUserInterfaceStyleDark;
	}
	return NO;
}

*/
import "C"
import (
	_ "unsafe"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

//go:linkname systemThemeVariant fyne.io/fyne/v2/internal/app.SystemTheme
var systemThemeVariant fyne.ThemeVariant

func init() {
	systemThemeVariant = theme.VariantLight
	if C.isDarkMode() {
		systemThemeVariant = theme.VariantDark
	}
}
