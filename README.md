# LRU Cache

### LRU_Cache

LRU_Cache реализует следующий интерфейс:
```go
type ICache interface {
    Add(key string, value interface{})
    Get(key string) (value interface{}, ok bool)
    Remove(key string)
}
```
LRU_Cache помещает новые или уже существующие запрашиваемые элементы в начало связного списка. 
Если кэш заполнен, то последний элемент удаляется. Таким образом, из кэша вытесняются значения, 
которые дольше всего не запрашивались.

### LRU_Cache_WithTTL

LRU_Cache_WithTTL реализует следующий интерфейс:
```go
    type ICache interface {
    Cap() int
    Clear()
    Add(key string, value interface{})
    AddWithTTL(key string, value interface{}, ttl time.Duration)
    Get(key string) (value interface{}, ok bool)
    Remove(key string)
}
```
LRU_Cache_WithTTL работает по такому же принципу, что и LRU_Cache, но также имеет возможность
добавлять элементы с определенным временным лимитом на хранение. 
Это достигается за счет того, что при создании кэша запускается горутина, 
которая отслеживает в очереди наличие элементов с истекшим временем хранения и удаляет их.
Эту горутину можно остановить с помощью CancelFunc, которая возвращается при создании кэша.


### LRU_Cache_WithTTL_v2

LRU_Cache_WithTTL_v2 реализует следующий интерфейс:
```go
type ICache interface {
    Cap() int
    Clear()
    Add(key string, value interface{})
    AddWithTTL(key string, value interface{}, ttl time.Duration)
    Get(key string) (value interface{}, ok bool)
    Remove(key string)
}
```
LRU_Cache_WithTTL_v2 это вторая версия кэша с возможностью добавления элементов с TTL.
Отличие от первой версии состоит в том, что здесь нет отслеживающей горутины – 
проверка и удаление элементов с истекших сроком хранения происходит при обращении к кэшу, 
а также может явно вызываться с помощью вызова метода UpdateExpirations().
