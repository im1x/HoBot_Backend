package vkplay

type VkplService2 interface {
	GetVkplToken() string
	refreshVkplToken() error
	isAuthNeedRefresh() bool
}
