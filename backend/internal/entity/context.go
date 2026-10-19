package entity

// ContextKey ...
type ContextKey string

// ContextValue ...
type ContextValue struct {
	M map[string]interface{}
}

func (c ContextValue) Get(key string) interface{} {
	return c.M[key]
}
