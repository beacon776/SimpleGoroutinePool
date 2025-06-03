package workerpool2

type Option func(*Pool)

func WithPreAlloc(preAlloc bool) Option {
	return func(p *Pool) {
		p.preAlloc = preAlloc
	}
}

func WithBlock(block bool) Option {
	return func(p *Pool) {
		p.block = block
	}
}
