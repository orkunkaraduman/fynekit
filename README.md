# Fyne Kit

***This Go module is still under core development. There may be breaking changes even between minor versions. If you
plan to use it in production, you should develop against a specific version. Some experimental features may be removed
in future versions, and new ones may be added.***

*[Fyne](https://fyne.io) is an easy-to-use UI toolkit and app API written in Go. It is designed to build applications
that run on desktop and mobile devices with a single codebase. View on [GitHub](https://github.com/fyne-io/fyne).*  
That was their own definition. To me, it is an open-source, cross-platform GUI solution in today's world, where
companies dictate their own APIs, ABIs, and programming languages. What makes it even better is that it uses Go, a
language that is close to both system-level and user-level programming and has strong concurrency support. This is a
GUI solution for Go; we have been waiting for years. I think [Fyne](https://fyne.io) has done a great job providing this need
and continues to improve.  
Of course, there are other GUI projects for Go as well. If you are interested in GUI development with Go, I recommend
checking them out too.

**Fyne Kit** is a Go module that contains a package or packages with various helper tools and utilities
for [Fyne](https://fyne.io). **Fyne Kit** is also an independent open-source project that is not affiliated with the
original [Fyne](https://fyne.io) project. It just aims to provide an extended toolkit for [Fyne](https://fyne.io).

## Highlights

### Model - View - View Controller | MVVC

#### View

While developing with ***Fyne***, I wanted a base widget named `View` to extend new `struct`s and scopes. This widget
should also support binding a `binding.DataItem` to a `binding.DataListener` or a function. When the View is disposed
of, `(*View).UnbindAll()` should remove all bindings from their `binding.DataListener`s.

```go
package main

import (
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/orkunkaraduman/fynekit"
)

type MyView struct {
	fynekit.View
	TextBinding1 binding.String
	TextBinding2 binding.String
}

func NewMyView() *MyView {
	v := &MyView{
		TextBinding1: binding.NewString(),
		TextBinding2: binding.NewString(),
	}
	v.ExtendBaseWidget()
	v.SetContent(container.NewVBox(
		widget.NewLabelWithData(v.TextBinding1),
		widget.NewEntryWithData(v.TextBinding2),
	))
	v.BindItemToFunction(v.TextBinding2, func() {
		x, _ := v.TextBinding2.Get()
		v.TextBinding1.Set(x)
	})
	return v
}

func main() {
	v := NewMyView()
	/* ... */
	v.UnbindAll()
}
```

#### View Controller

Binding works the same way in `ViewController` as it does in `View`. However, `ViewController` is not a widget.

### Map Widget for Fyne

I implemented an online map widget for Fyne. Currently, it only supports OpenStreetMap.

I haven't prepared any documentation for it yet, but I believe you can quickly get started by looking at the source
code.

### Internals & Locale & Utilities

#### Internals

I needed to access some Fyne internals for features that I believe will be improved in future Fyne versions. So I had to
use `//go:linkname` to create some workarounds.

You can find the related source code in files matching the regex pattern `internals(_.*)?.go`.

#### Locale

By default, Fyne uses the system locale. However, in multilingual applications, users may need to override the locale
and language.

For this reason, I added support for overriding the locale in `locale.go`, using some code from Fyne's `lang` package.

#### Utilities

I added some useful utility functions, such as `Walk`, `Wrap`, and `WrapWithMinSize`. You can also find their shorter
versions: `R`, `W`, and `WM`.

`Walk` allows you to traverse objects and their child objects in the view tree. I used it effectively in
`(*View).UnbindAll()`.

The `Wrap` functions, especially the `W` shortcut, are among my favorites. They are very useful for building the view
tree inline. Instead of assigning each widget to a variable and then using those variables to build the view tree, we
can modify widgets directly while creating the tree inline.

## Contributing

We welcome contributions from the community to improve and expand `fynekit`'s capabilities. If you find a bug, have a
feature request, or want to contribute code, please follow our guidelines for contributing
[CONTRIBUTING.md](CONTRIBUTING.md) and submit a pull request.

## License

`fynekit` is open-source software released under
the [BSD 3-Clause License](https://opensource.org/licenses/BSD-3-Clause).

## Acknowledgments

I would like to thank the open-source community and the developers of the libraries and tools that `fynekit` depends on.

## Contact

If you have any questions, suggestions, or need support, you can reach me
at [ok@orkunkaraduman.com](mailto:ok@orkunkaraduman.com).
