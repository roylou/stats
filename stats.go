// Package stats defines a lightweight interface for collecting statistics. It
// doesn't provide an implementation, just the shared interface.
package stats

// Client provides methods to collection statistics.
type Client interface {
	// BumpAvg bumps the average for the given key.
	BumpAvg(key string, val float64, tags ...string)

	// BumpSum bumps the sum for the given key.
	BumpSum(key string, val float64, tags ...string)

	// BumpHistogram bumps the histogram for the given key.
	BumpHistogram(key string, val float64, tags ...string)

	// BumpTime is a special version of BumpHistogram which is specialized for
	// timers. Calling it starts the timer, and it returns a value on which End()
	// can be called to indicate finishing the timer. A convenient way of
	// recording the duration of a function is calling it like such at the top of
	// the function:
	//
	//     defer s.BumpTime("my.function").End()
	BumpTime(key string, tags ...string) interface {
		End()
	}
}

// PrefixClient adds multiple keys for the same value, with each prefix
// added to the key and calls the underlying client.
func PrefixClient(prefixes []string, client Client) Client {
	return &prefixClient{
		Prefixes: prefixes,
		Client:   client,
	}
}

type prefixClient struct {
	Prefixes []string
	Client   Client
}

func (p *prefixClient) BumpAvg(key string, val float64, tags ...string) {
	for _, prefix := range p.Prefixes {
		p.Client.BumpAvg(prefix+key, val, tags...)
	}
}

func (p *prefixClient) BumpSum(key string, val float64, tags ...string) {
	for _, prefix := range p.Prefixes {
		p.Client.BumpSum(prefix+key, val, tags...)
	}
}

func (p *prefixClient) BumpHistogram(key string, val float64, tags ...string) {
	for _, prefix := range p.Prefixes {
		p.Client.BumpHistogram(prefix+key, val, tags...)
	}
}

func (p *prefixClient) BumpTime(key string, tags ...string) interface {
	End()
} {
	var m multiEnder
	for _, prefix := range p.Prefixes {
		m = append(m, p.Client.BumpTime(prefix+key, tags...))
	}
	return m
}

// multiEnder combines many enders together.
type multiEnder []interface {
	End()
}

func (m multiEnder) End() {
	for _, e := range m {
		e.End()
	}
}

// HookClient is useful for testing. It provides optional hooks for each
// expected method in the interface, which if provided will be called. If a
// hook is not provided, it will be ignored.
type HookClient struct {
	BumpAvgHook       func(key string, val float64, tags ...string)
	BumpSumHook       func(key string, val float64, tags ...string)
	BumpHistogramHook func(key string, val float64, tags ...string)
	BumpTimeHook      func(key string, tags ...string) interface {
		End()
	}
}

// BumpAvg will call BumpAvgHook if defined.
func (c *HookClient) BumpAvg(key string, val float64, tags ...string) {
	if c.BumpAvgHook != nil {
		c.BumpAvgHook(key, val, tags...)
	}
}

// BumpSum will call BumpSumHook if defined.
func (c *HookClient) BumpSum(key string, val float64, tags ...string) {
	if c.BumpSumHook != nil {
		c.BumpSumHook(key, val, tags...)
	}
}

// BumpHistogram will call BumpHistogramHook if defined.
func (c *HookClient) BumpHistogram(key string, val float64, tags ...string) {
	if c.BumpHistogramHook != nil {
		c.BumpHistogramHook(key, val, tags...)
	}
}

// BumpTime will call BumpTimeHook if defined.
func (c *HookClient) BumpTime(key string, tags ...string) interface {
	End()
} {
	if c.BumpTimeHook != nil {
		return c.BumpTimeHook(key, tags...)
	}
	return NoOpEnd
}

type noOpEnd struct{}

func (n noOpEnd) End() {}

// NoOpEnd provides a dummy value for use in tests as valid return value for
// BumpTime().
var NoOpEnd = noOpEnd{}

// BumpAvg calls BumpAvg on the Client if it isn't nil. This is useful when a
// component has an optional stats.Client.
func BumpAvg(c Client, key string, val float64, tags ...string) {
	if c != nil {
		c.BumpAvg(key, val, tags...)
	}
}

// BumpSum calls BumpSum on the Client if it isn't nil. This is useful when a
// component has an optional stats.Client.
func BumpSum(c Client, key string, val float64, tags ...string) {
	if c != nil {
		c.BumpSum(key, val, tags...)
	}
}

// BumpHistogram calls BumpHistogram on the Client if it isn't nil. This is
// useful when a component has an optional stats.Client.
func BumpHistogram(c Client, key string, val float64, tags ...string) {
	if c != nil {
		c.BumpHistogram(key, val, tags...)
	}
}

// BumpTime calls BumpTime on the Client if it isn't nil. If the Client is nil
// it still returns a valid return value which will be a no-op. This is useful
// when a component has an optional stats.Client.
func BumpTime(c Client, key string, tags ...string) interface {
	End()
} {
	if c != nil {
		return c.BumpTime(key, tags...)
	}
	return NoOpEnd
}
