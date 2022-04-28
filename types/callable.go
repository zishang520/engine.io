package types

type Callable func()

var Noop Callable = func() {}
