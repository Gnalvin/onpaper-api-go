package cache

// CtxCacheVale 缓存中间件传递的值类型
type CtxCacheVale struct {
	Val       interface{}
	HaveCache bool
}

// NeedFindData 缓存中少读取的数据 id
type NeedFindData struct {
	Id    string
	Index int
}
