package managers

import (
	"eclipse/pkg/interfaces"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

var _ interfaces.ProxyManagerInterface = (*ProxyManager)(nil)

type ProxyManager struct {
	proxies          []string
	accountsCount    int
	accountsPerProxy int
	currentIndex     int
}

func NewProxyManager(proxies []string, accountsCount int) *ProxyManager {
	var accountsPerProxy int
	if len(proxies) >= accountsCount {
		accountsPerProxy = 1
	} else {
		accountsPerProxy = accountsCount / len(proxies)
		if accountsCount%len(proxies) != 0 {
			accountsPerProxy++
		}
	}

	log.Printf("Прокси настроены | подгружены %d прокси и %d аккаунтов | будет использоваться %d аккаунтов на 1 прокси",
		len(proxies),
		accountsCount,
		accountsPerProxy,
	)

	return &ProxyManager{
		proxies:          proxies,
		accountsCount:    accountsCount,
		accountsPerProxy: accountsPerProxy,
		currentIndex:     0,
	}
}

func (pm *ProxyManager) GetHttpClient(accountIndex int) *http.Client {
	proxyURL := pm.GetProxyForAccount(accountIndex)

	transport := &http.Transport{
		Proxy: http.ProxyURL(pm.parseProxy(proxyURL)),
	}

	return &http.Client{
		Transport: transport,
	}
}

func (pm *ProxyManager) GetProxyForAccount(accountIndex int) string {
	if len(pm.proxies) >= pm.accountsCount {
		return pm.proxies[accountIndex]
	}

	proxyIndex := accountIndex / pm.accountsPerProxy
	if proxyIndex >= len(pm.proxies) {
		proxyIndex = len(pm.proxies) - 1
	}

	proxy := pm.proxies[proxyIndex]
	log.Printf("Аккаунт %d использует прокси: http://%s", accountIndex+1, proxy)
	return proxy
}

func (pm *ProxyManager) parseProxy(proxyURL string) *url.URL {
	parsedURL, err := url.Parse(fmt.Sprintf("http://%s", proxyURL))
	if err != nil {
		log.Printf("Ошибка парсинга прокси %s: %v", proxyURL, err)
		return nil
	}
	return parsedURL
}

func (pm *ProxyManager) GetProxyURL(accountIndex int) string {
	proxy := pm.GetProxyForAccount(accountIndex)
	return fmt.Sprintf("http://%s", proxy)
}
