package types

type Callable func()

// Noop function.
var Noop Callable = func() {}
