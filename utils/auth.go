package utils

import "sync"

type AuthValidator struct {
	sync.RWMutex
	data map[interface{}]int // key && auth
}

// NewAuthValidator 创建一个权限检验器
func NewAuthValidator() *AuthValidator {
	au := &AuthValidator{
		data: make(map[interface{}]int),
	}
	return au
}

// AddAuthData 添加对象权限
func (au *AuthValidator) AddAuthData(key interface{}, maxAuth int) {
	au.Lock()
	defer au.Unlock()
	au.data[key] = maxAuth
}

// GetAuth 获取对象的权限值
func (au *AuthValidator) GetAuth(key interface{}) int {
	au.RLock()
	defer au.RUnlock()
	if val, ok := au.data[key]; ok {
		return val
	}
	return 0
}

// Validate 验证操作是否有权限
func (au *AuthValidator) Validate(key interface{}, operation int) bool {
	au.RLock()
	defer au.RUnlock()
	if val, ok := au.data[key]; ok {
		if val&operation == operation {
			return true
		} else {
			return false
		}
	}
	// 没有在权限管控内
	return true
}
