package font

type CheckErrFn func(error) bool

type SubsetOption func(*subsetConfig)

type subsetConfig struct {
	concurrent bool
	checkGlyph bool
	fn         CheckErrFn
}

func WithConcurrent() SubsetOption {
	return func(c *subsetConfig) {
		c.concurrent = true
	}
}

func WithCheckGlyph() SubsetOption {
	return func(c *subsetConfig) {
		c.checkGlyph = true
	}
}
func WithCheckErr(fn CheckErrFn) SubsetOption {
	return func(c *subsetConfig) {
		c.fn = fn
	}
}
