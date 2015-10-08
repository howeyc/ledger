package staticbin

import "github.com/go-martini/martini"

// ClassicWithoutStatic creates a classic Martini without a default Static.
func ClassicWithoutStatic() *martini.ClassicMartini {
	r := martini.NewRouter()
	m := martini.New()
	m.Use(martini.Logger())
	m.Use(martini.Recovery())
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)
	return &martini.ClassicMartini{m, r}
}

// Classic creates a classic Martini with a default Static.
func Classic(asset func(string) ([]byte, error)) *martini.ClassicMartini {
	m := ClassicWithoutStatic()
	m.Use(Static(defaultDir, asset))
	return m
}
