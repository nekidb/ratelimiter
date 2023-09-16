package ratelimiter

import (
	"strings"
	"sync"
	"time"
)

type RateLimiter struct {
	counter    map[string]int       // количество выполненных запросов из подсети
	lastHits   map[string]time.Time // время последнего запроса из подсети
	prefixSize int                  // размер префикса, считываемый из ip адреса
	limit      int                  // максимальное количество запросов до выставления лимита
	cooldown   time.Duration        // время, после которого лимит сбрасывается
	mu         sync.RWMutex         // мьютекс для потокобезопасности мапы
}

func NewRateLimiter(prefixSize int, limit int, cooldown time.Duration) *RateLimiter {
	return &RateLimiter{
		counter:    make(map[string]int),
		lastHits:   make(map[string]time.Time),
		prefixSize: prefixSize,
		limit:      limit,
		cooldown:   cooldown,
	}
}

// Функция увеличивает счетчик и обновляет время последнего запроса для подсети переданного ip
func (l *RateLimiter) Increment(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Вытаскиваем подсеть из ip
	subnet := extractSubnet(ip, l.prefixSize)

	// Если время кулдауна прошло, то сбрасываем счетчик
	if lastHit, ok := l.lastHits[subnet]; ok && time.Since(lastHit) > l.cooldown {
		l.counter[subnet] = 0
	}

	l.counter[subnet]++
	l.lastHits[subnet] = time.Now()
}

// Функция проверяет, выставлен ли лимит для подсети переданного ip
func (l *RateLimiter) IsLimited(ip string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Вытаскиваем подсеть из ip
	subnet := extractSubnet(ip, l.prefixSize)

	// Если время кулдауна прошло, то лимита нет
	if lastHit, ok := l.lastHits[subnet]; ok && time.Since(lastHit) > l.cooldown {
		return false
	}

	// Если количество запросов для подсети превышает число лимита, то лимит выставлен на него
	if l.counter[subnet] >= l.limit {
		return true
	}

	return false
}

// Функция сбрасывает лимит для подсети
func (l *RateLimiter) Reset(subnet string) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	delete(l.counter, subnet)
	delete(l.lastHits, subnet)
}

// Функция вытаскивает подсеть из ip по размеру префикса
func extractSubnet(ip string, prefixSizeInBits int) string {
	// Для упрощения предполагается, что подсеть состоит из полных чисел представленных в ip
	prefixSizeInBytes := prefixSizeInBits / 8

	parts := strings.Split(ip, ".")

	return strings.Join(parts[:prefixSizeInBytes], ".")
}
